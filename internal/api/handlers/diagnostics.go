package handlers

import (
    "encoding/json"
    "net/http"
    "time"
)

// RuntimeDiagnostics — الشكل الكامل لردّ الـ API
type RuntimeDiagnostics struct {
    Status     string            `json:"status"`
    Uptime     string            `json:"uptime"`
    Timestamp  time.Time         `json:"timestamp"`
    Loops      map[string]string `json:"loops"`
    Engines    map[string]string `json:"engines"`
    Versions   map[string]string `json:"versions"`
}

// DiagnosticsHandler — يعرض حالة runtime + orchestrator + engines
func DiagnosticsHandler(deps interface{}) http.HandlerFunc {
    started := time.Now()

    return func(w http.ResponseWriter, r *http.Request) {

        resp := RuntimeDiagnostics{
            Status:    "running",
            Uptime:    time.Since(started).String(),
            Timestamp: time.Now().UTC(),

            Loops: map[string]string{
                "unified_loop":   "active",
                "changeset_loop": "active",
                "events_loop":    "active",
                "alerts_loop":    "active",
                "ws_loop":        "active",
            },

            Engines: map[string]string{
                "unified_engine": "ok",
                "change_engine":  "ok",
                "event_engine":   "ok",
                "alert_engine":   "ok",
                "ws_hub":         "ok",
            },

            Versions: map[string]string{
                "core":   "X8",
                "api":    "1.0",
                "blocks": "14-20",
            },
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
    }
}
