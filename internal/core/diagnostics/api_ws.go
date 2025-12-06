package diagnostics

import (
    "context"
    "encoding/json"
    "net/http"

    "jms-core-x/internal/ws/transport"
)

// WebSocketConn is a minimal abstraction over a WebSocket connection used by
// WSAPI. In the current architecture it is an alias of the shared ws/transport.Conn
// interface so that Block 9 can reuse the same WebSocket transport layer as Block 17.
type WebSocketConn = transport.Conn

// WebSocketUpgrader upgrades an HTTP connection to a WebSocket connection.
// It is also defined as an alias to ws/transport.Upgrader to avoid coupling
// higher-level layers back to the diagnostics package.
type WebSocketUpgrader = transport.Upgrader

// WSAPI exposes DiagnosticsLayer ping samples over WebSocket.
//
// Block 9 does not select a concrete WebSocket implementation; that decision
// is delegated to higher-level wiring code which will provide a concrete
// WebSocketUpgrader.
type WSAPI struct {
    diag     *DiagnosticsLayer
    upgrader WebSocketUpgrader
}

// NewWSAPI constructs a new WebSocket diagnostics API using the provided
// DiagnosticsLayer and WebSocketUpgrader.
func NewWSAPI(diag *DiagnosticsLayer, upgrader WebSocketUpgrader) *WSAPI {
    return &WSAPI{
        diag:     diag,
        upgrader: upgrader,
    }
}

// HandlePingStream upgrades the incoming HTTP connection to a WebSocket and
// streams PingSample values for a given device in real time.
//
// Expected route (configured externally):
//   GET /diagnostics/devices/{id}/ping/ws
func (api *WSAPI) HandlePingStream(w http.ResponseWriter, r *http.Request) {
    if api == nil || api.diag == nil {
        http.Error(w, "diagnostics WSAPI not configured", http.StatusInternalServerError)
        return
    }
    if api.upgrader == nil {
        http.Error(w, "websocket upgrader not configured", http.StatusInternalServerError)
        return
    }

    deviceID := extractDeviceID(r)
    if deviceID == "" {
        http.Error(w, "missing device id", http.StatusBadRequest)
        return
    }

    // Upgrade HTTP connection to WebSocket.
    conn, err := api.upgrader.Upgrade(w, r)
    if err != nil {
        // Upgrader is responsible for writing any HTTP-level error response.
        return
    }
    defer conn.Close()

    // Subscribe to the active diagnostics session for this device.
    listener, err := api.diag.Subscribe(deviceID)
    if err != nil {
        _ = conn.WriteJSON(map[string]string{
            "error": err.Error(),
        })
        return
    }
    defer api.diag.Unsubscribe(deviceID, listener)

    ctx := withWSContext(r)

    // Forward samples until the context is cancelled or the session ends.
    for {
        select {
        case <-ctx.Done():
            return
        case sample, ok := <-listener:
            if !ok {
                // Channel closed by the session manager when the session finishes.
                return
            }
            if err := conn.WriteJSON(sample); err != nil {
                // Any write error terminates the stream.
                return
            }
        }
    }
}

// encodeJSON is a small helper to keep WebSocket JSON encoding consistent with
// the HTTP API helpers. It intentionally mirrors the behaviour of writeJSON
// in api_http.go but returns raw bytes so callers can decide how to send them.
func encodeJSON(v interface{}) ([]byte, error) {
    return json.Marshal(v)
}

// extractDeviceID resolves the device ID from the incoming HTTP request.
// For now it looks at the URL query parameters. It can be extended later
// to parse path parameters if the router injects them differently.
func extractDeviceID(r *http.Request) string {
    if r == nil {
        return ""
    }
    q := r.URL.Query()
    if id := q.Get("id"); id != "" {
        return id
    }
    if id := q.Get("device_id"); id != "" {
        return id
    }
    return ""
}

// withWSContext allows higher-level code to derive a cancellable context from
// an HTTP request when using alternate upgrade mechanisms. It is included here
// to mirror the HTTP API withTimeout helper and to keep the WebSocket handling
// aligned with context cancellation patterns used across the codebase.
func withWSContext(r *http.Request) context.Context {
    if r == nil {
        return context.Background()
    }
    return r.Context()
}
