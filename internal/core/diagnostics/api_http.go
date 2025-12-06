// ============================================================================
// DEPRECATED (Block 9 Prototype)
// ----------------------------------------------------------------------------
// This file belongs to the OLD Diagnostics HTTP API prototype (Block 9).
// It is NOT used by the new system architecture (Blocks 17–20).
// DO NOT IMPORT OR USE THIS API IN ANY NEW CODE.
// The official HTTP API is under: internal/api/ (Block 18)
// This file is kept only as historical reference.
// ============================================================================

package diagnostics

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "strconv"
    "strings"
    "time"
)

// HTTPAPI exposes DiagnosticsLayer functionality over REST-style HTTP handlers.
// Block 9: this type only wraps existing DiagnosticsLayer methods without
// introducing any new diagnostics logic.
type HTTPAPI struct {
    diag *DiagnosticsLayer
}

// NewHTTPAPI constructs a new HTTPAPI facade for the provided DiagnosticsLayer.
func NewHTTPAPI(diag *DiagnosticsLayer) *HTTPAPI {
    if diag == nil {
        panic("diagnostics: HTTPAPI requires non-nil DiagnosticsLayer")
    }
    return &HTTPAPI{diag: diag}
}

// startPingRequest represents the optional JSON body for HandleStartPing.
type startPingRequest struct {
    DurationSeconds int `json:"duration,omitempty"`
    IntervalSeconds int `json:"interval,omitempty"`
}

// pingStatusResponse is a small helper DTO returned after starting a ping session.
type pingStatusResponse struct {
    Status string `json:"status"`
}

// watchlistRequest is the JSON payload for adding a device to the watchlist.
type watchlistRequest struct {
    DeviceID string `json:"device_id"`
    Note     string `json:"note,omitempty"`
}

// diagnosticsError is a simple JSON error envelope.
type diagnosticsError struct {
    Error string `json:"error"`
}

// ----------------------------------------------------------------------
// HandleStartPing  (FIXED to accept session,error from StartDevicePing)
// ----------------------------------------------------------------------
func (api *HTTPAPI) HandleStartPing(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    deviceID := extractDeviceID(r)
    if deviceID == "" {
        writeJSONError(w, http.StatusBadRequest, "missing device id")
        return
    }

    // Defaults from Block9_Blueprint_v2.md
    duration := 60 * time.Second
    interval := 1 * time.Second

    if r.Body != nil {
        defer r.Body.Close()
        var body startPingRequest
        if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
            if body.DurationSeconds > 0 {
                duration = time.Duration(body.DurationSeconds) * time.Second
            }
            if body.IntervalSeconds > 0 {
                interval = time.Duration(body.IntervalSeconds) * time.Second
            }
        }
    }

    _, err := api.diag.StartDevicePing(ctx, deviceID, duration, interval)
    if err != nil {
        writeJSONError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, pingStatusResponse{Status: "started"})
}

// ----------------------------------------------------------------------
// HandlePingSummary
// ----------------------------------------------------------------------
func (api *HTTPAPI) HandlePingSummary(w http.ResponseWriter, r *http.Request) {
    deviceID := extractDeviceID(r)
    if deviceID == "" {
        writeJSONError(w, http.StatusBadRequest, "missing device id")
        return
    }

    summary, ok := api.diag.LastSummary(deviceID)
    if !ok {
        writeJSONError(w, http.StatusNotFound, "no ping summary for device")
        return
    }

    writeJSON(w, http.StatusOK, summary)
}

// ----------------------------------------------------------------------
// HandleDeviceDiagnostics  (FIXED to use diag,ok := ... )
// ----------------------------------------------------------------------
func (api *HTTPAPI) HandleDeviceDiagnostics(w http.ResponseWriter, r *http.Request) {
    deviceID := extractDeviceID(r)
    if deviceID == "" {
        writeJSONError(w, http.StatusBadRequest, "missing device id")
        return
    }

    diag, ok := api.diag.DeviceDiagnostics(deviceID)
    if !ok {
        writeJSONError(w, http.StatusNotFound, "device diagnostics not found")
        return
    }

    writeJSON(w, http.StatusOK, diag)
}

// ----------------------------------------------------------------------
// Watchlist operations
// ----------------------------------------------------------------------
func (api *HTTPAPI) HandleAddToWatchlist(w http.ResponseWriter, r *http.Request) {
    if r.Body == nil {
        writeJSONError(w, http.StatusBadRequest, "missing request body")
        return
    }
    defer r.Body.Close()

    var req watchlistRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
        return
    }
    if req.DeviceID == "" {
        writeJSONError(w, http.StatusBadRequest, "device_id is required")
        return
    }

    api.diag.AddToWatchlist(req.DeviceID, req.Note)
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (api *HTTPAPI) HandleRemoveFromWatchlist(w http.ResponseWriter, r *http.Request) {
    deviceID := extractDeviceID(r)
    if deviceID == "" {
        writeJSONError(w, http.StatusBadRequest, "missing device id")
        return
    }

    api.diag.RemoveFromWatchlist(deviceID)
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (api *HTTPAPI) HandleListWatchlist(w http.ResponseWriter, r *http.Request) {
    entries := api.diag.ListWatchlist()
    writeJSON(w, http.StatusOK, entries)
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------
func extractDeviceID(r *http.Request) string {
    if id := r.URL.Query().Get("device_id"); id != "" {
        return id
    }

    path := strings.Trim(r.URL.Path, "/")
    if path == "" {
        return ""
    }
    segments := strings.Split(path, "/")
    if len(segments) == 0 {
        return ""
    }

    ignore := map[string]struct{}{
        "diagnostics": {},
        "devices":     {},
        "ping":        {},
        "summary":     {},
        "ws":          {},
        "watchlist":   {},
    }

    for i := len(segments) - 1; i >= 0; i-- {
        s := segments[i]
        if s == "" {
            continue
        }
        if _, ok := ignore[s]; ok {
            continue
        }
        return s
    }

    return ""
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}
) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    enc := json.NewEncoder(w)
    enc.SetEscapeHTML(true)
    _ = enc.Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
    if msg == "" {
        msg = http.StatusText(status)
    }
    writeJSON(w, status, diagnosticsError{Error: msg})
}

func withTimeoutSeconds(parent context.Context, seconds int) (context.Context, context.CancelFunc) {
    if seconds <= 0 {
        return parent, func() {}
    }
    return context.WithTimeout(parent, time.Duration(seconds)*time.Second)
}

func parseIntQuery(r *http.Request, key string) (int, error) {
    raw := r.URL.Query().Get(key)
    if raw == "" {
        return 0, errors.New("missing query parameter")
    }
    n, err := strconv.Atoi(raw)
    if err != nil {
        return 0, err
    }
    return n, nil
}
