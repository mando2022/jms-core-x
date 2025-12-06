package api

import (
    "context"
    "net/http"

    "jms-core-x/internal/alerts"
    "jms-core-x/internal/core/diagnostics"
    "jms-core-x/internal/core/final"
    "jms-core-x/internal/events"
    "jms-core-x/internal/history"
    "jms-core-x/internal/ws"
    "jms-core-x/internal/ws/transport"
)

// FinalSnapshotProvider هو المصدر الوحيد لـ FinalSnapshot من Block 12.
type FinalSnapshotProvider interface {
    FinalSnapshot(ctx context.Context) final.FinalSnapshot
}

// HistoryReader يوفّر واجهات القراءة التي يحتاجها HTTP API Layer من Block 13.
type HistoryReader interface {
    DeviceHistory(ctx context.Context) ([]history.DeviceHistory, error)
    DeviceTimeline(ctx context.Context, deviceID string) ([]history.TimelineEntry, error)
}

// AlertsReader يوفّر واجهات القراءة لمحرك الإنذارات (Block 16).
type AlertsReader interface {
    ListAlerts(ctx context.Context) ([]alerts.Alert, error)
}

// DiagnosticsReader يوفّر طريقة قراءة نتائج الـ Auto-Ping/Diagnostics (Block 9/Auto-Ping V2).
// التنفيذ الفعلي يتم حقنه من الـ Runtime (Block 20).
type DiagnosticsReader interface {
    // Sessions يرجع قائمة بالـ diagnostics snapshot لكل جهاز تمت متابعته.
    Sessions() []diagnostics.DeviceDiagnostics

    // DeviceDiagnostics يرجع آخر snapshot لـ جهاز محدد.
    DeviceDiagnostics(deviceID string) (diagnostics.DeviceDiagnostics, bool)

    // LastSummary يرجع ملخص آخر جلسة Ping لجهاز محدد (إن وُجدت).
    LastSummary(deviceID string) (diagnostics.PingSummary, bool)
}

// Deps يجمع كل الـ Providers وHub المطلوبة لبناء HTTP Router.
type Deps struct {
    Final       FinalSnapshotProvider
    History     HistoryReader
    Events      events.EventReader
    Alerts      alerts.AlertReader
    Diagnostics DiagnosticsReader

    Hub *ws.Hub
    // WSUpgrader لم يعد يعتمد على diagnostics.WebSocketUpgrader؛
    // تم نقله إلى واجهة عامة داخل ws/transport حتى يبقى Block 18 مستقل
    // عن تفاصيل Block 9 القديمة.
    WSUpgrader transport.Upgrader

    Version string
}

// NewRouter ينشئ HTTP Router كامل لبلوك 18 ويعيد handler جاهز للاستخدام.
func NewRouter(d Deps) http.Handler {
    mux := http.NewServeMux()

    // Final Snapshot (Block 12)
    finalHandler := NewFinalSnapshotHandler(d.Final)
    mux.HandleFunc("/api/final-snapshot", finalHandler.HandleFinalSnapshot)

    // History (Block 13)
    historyHandler := NewHistoryHandler(d.History)
    mux.HandleFunc("/api/history", historyHandler.HandleDeviceHistory)
    mux.HandleFunc("/api/history/timeline/", historyHandler.HandleDeviceTimeline)

    // Events (Block 15)
    eventsHandler := NewEventsHandler(d.Events)
    mux.HandleFunc("/api/events", eventsHandler.HandleEvents)

    // Alerts (Block 16)
    alertsHandler := NewAlertsHandler(d.Alerts)
    mux.HandleFunc("/api/alerts", alertsHandler.HandleAlerts)

    // Diagnostics (Block 9 / Auto-Ping V2)
    diagnosticsHandler := NewDiagnosticsHandler(d.Diagnostics)
    mux.HandleFunc("/api/diagnostics/sessions", diagnosticsHandler.HandleSessions)
    mux.HandleFunc("/api/diagnostics/session/", diagnosticsHandler.HandleSessionByID)
    mux.HandleFunc("/api/diagnostics/ping/", diagnosticsHandler.HandlePingSummary)

    // WebSocket Hub (Block 17)
    wsHandler := NewWSHandler(d.Hub, d.WSUpgrader)
    mux.HandleFunc("/api/ws", wsHandler.HandleWebSocket)

    // Version
    mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            methodNotAllowed(w, r, http.MethodGet)
            return
        }
        payload := map[string]string{"version": d.Version}
        respondJSON(w, http.StatusOK, payload)
    })

    return applyMiddleware(mux)
}
