package api

import (
    "net/http"

    "github.com/gorilla/mux"
)

// Dependencies needed by the router
type RouterDeps struct {
    UnifiedProvider UnifiedProvider       // Block 6–8 Unified Snapshot
    ChangeEngine    ChangeSetProvider     // Block 14
    EventsReader    EventReader           // Block 15
    AlertsReader    AlertsReader          // Block 16
    TimelineStore   TimelineReader        // Block 13
    SnapshotStore   SnapshotReader        // Block 12
    WSUpgrade       http.Handler          // Block 17 WebSocket Hub
}

// BuildRouter يقوم ببناء الراوتر كاملاً
func BuildRouter(deps RouterDeps) http.Handler {
    r := mux.NewRouter()

    // Middleware
    r.Use(loggingMiddleware)
    r.Use(corsMiddleware)
    r.Use(recoveryMiddleware)

    api := r.PathPrefix("/api").Subrouter()

    // ----------- BASIC & HEALTH -----------
    api.HandleFunc("/health", HealthHandler).Methods("GET")

    // ----------- DEVICES (Unified Snapshot) -----------
    api.HandleFunc("/devices", DevicesHandler(deps.UnifiedProvider)).Methods("GET")

    // ----------- SNAPSHOTS -----------
    api.HandleFunc("/snapshots", SnapshotsHandler(deps.SnapshotStore)).Methods("GET")

    // ----------- TIMELINE -----------
    api.HandleFunc("/timeline", TimelineHandler(deps.TimelineStore)).Methods("GET")

    // ----------- CHANGESET -----------
    api.HandleFunc("/changeset", ChangeSetHandler(deps.ChangeEngine)).Methods("GET")

    // ----------- EVENTS -----------
    api.HandleFunc("/events", EventsHandler(deps.EventsReader)).Methods("GET")

    // ----------- ALERTS -----------
    api.HandleFunc("/alerts", AlertsHandler(deps.AlertsReader)).Methods("GET")

    // ----------- DIAGNOSTICS -----------
    api.HandleFunc("/diagnostics", DiagnosticsHandler(deps)).Methods("GET")

    // ----------- WEBSOCKET -----------
    r.Handle("/ws", deps.WSUpgrade)

    // ----------- NOT FOUND -----------
    r.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

    return r
}
