// File: internal/bootstrap/api_ws.go
package bootstrap

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "jms-core-x/internal/api"
    "jms-core-x/internal/alerts"
    "jms-core-x/internal/core/diagnostics"
    "jms-core-x/internal/core/final"
    "jms-core-x/internal/events"
    "jms-core-x/internal/history"
    "jms-core-x/internal/ws"
)

//
// =====================================================
//   FINAL SNAPSHOT PROVIDER ADAPTER  (Block 18)
// =====================================================
//

type finalProviderAdapter struct {
    engine *final.Engine
}

func newFinalProviderAdapter(engine *final.Engine) api.FinalSnapshotProvider {
    return &finalProviderAdapter{engine: engine}
}

func (a *finalProviderAdapter) FinalSnapshot(ctx context.Context) final.FinalSnapshot {
    return a.engine.FinalSnapshot(ctx)
}

//
// =====================================================
//   HISTORY API ADAPTER (Block 18 ↔ Block 13)
// =====================================================
//

type historyAPIAdapter struct {
    store *history.Store
}

func newHistoryAPIAdapter(store *history.Store) api.HistoryReader {
    return &historyAPIAdapter{store: store}
}

// كل وظائف Store أصبحت جاهزة داخل Block 13
func (a *historyAPIAdapter) DeviceHistory(ctx context.Context) ([]history.DeviceHistory, error) {
    return a.store.DeviceHistory(ctx)
}

func (a *historyAPIAdapter) DeviceTimeline(ctx context.Context, deviceID string) ([]history.DeviceTimeline, error) {
    return a.store.DeviceTimeline(ctx, deviceID)
}

func (a *historyAPIAdapter) LoadDeviceHistory(deviceID string) (history.DeviceHistory, error) {
    return a.store.LoadDeviceHistory(deviceID)
}

//
// =====================================================
//   EVENTS API ADAPTER  (Block 18)
// =====================================================
//

type eventsAPIAdapter struct {
    store *events.DBStore
}

func newEventsAPIAdapter(store *events.DBStore) api.EventReader {
    return &eventsAPIAdapter{store: store}
}

func (a *eventsAPIAdapter) ListEvents(ctx context.Context, limit int) ([]events.Event, error) {
    if limit <= 0 {
        limit = 200
    }
    return a.store.LoadRecent(ctx, limit)
}

//
// =====================================================
//   ALERTS API ADAPTER (Block 18)
// =====================================================
//

type alertsAPIAdapter struct {
    store *alerts.DBStore
}

func newAlertsAPIAdapter(store *alerts.DBStore) api.AlertReader {
    return &alertsAPIAdapter{store: store}
}

func (a *alertsAPIAdapter) ListAlerts(ctx context.Context, limit int) ([]alerts.Alert, error) {
    if limit <= 0 {
        limit = 200
    }
    return a.store.LoadRecent(ctx, limit)
}

//
// =====================================================
//   DIAGNOSTICS API ADAPTER (Block 18)
// =====================================================
//

type diagnosticsNOPAdapter struct{}

func newDiagnosticsNOPAdapter() api.DiagnosticsReader {
    return diagnosticsNOPAdapter{}
}

func (diagnosticsNOPAdapter) Sessions(context.Context) ([]string, error) {
    return []string{}, nil
}

func (diagnosticsNOPAdapter) DeviceDiagnostics(ctx context.Context, deviceID string) (diagnostics.DeviceDiagnostics, bool) {
    return diagnostics.DeviceDiagnostics{}, false
}

func (diagnosticsNOPAdapter) LastSummary(deviceID string) (diagnostics.PingSummary, bool) {
    return diagnostics.PingSummary{}, false
}

//
// =====================================================
//   HTTP + WS FRONTEND BUILDER
// =====================================================
//

type Frontend struct {
    Server     *http.Server
    Hub        *ws.Hub
    HTTPRouter http.Handler
}

func BuildFrontend(
    ctx context.Context,
    deps api.Deps,
    port int,
) (*Frontend, error) {

    hub := deps.Hub

    // HTTP Router
    router := api.NewRouter(deps)

    srv := &http.Server{
        Addr:    ":" + strconv.Itoa(port),
        Handler: router,
    }

    // تشغيل Hub في Goroutine منفصل
    go func() {
        hub.Run()
    }()

    return &Frontend{
        Server:     srv,
        Hub:        hub,
        HTTPRouter: router,
    }, nil
}

//
// =====================================================
//   INTERNAL — JSON UTIL
// =====================================================
//

// writeJSON is used only internally inside this bootstrap file.
// (Handlers in api package have their own writer.)
func writeJSON(w http.ResponseWriter, v interface{
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
}) {
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(v)
}

//
// =====================================================
//   ERROR HELPERS
// =====================================================
//

func httpError(w http.ResponseWriter, msg string, status int) {
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(map[string]interface{}{
        "error": msg,
    })
}

//
// =====================================================
//   PARAM EXTRACTOR (matching api.Param)
// =====================================================
//

// extractPathParam هو بديل آمن عند الحاجة
func extractPathParam(r *http.Request, key string) string {
    v := r.URL.Query().Get(key)
    return v
}

// END OF FILE
