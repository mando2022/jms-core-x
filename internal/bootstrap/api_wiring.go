package bootstrap

import (
	"net/http"

	"jms-core-x/internal/api"
	"jms-core-x/internal/api/handlers"
	"jms-core-x/internal/api/readers"
	"jms-core-x/internal/core/diagnostics"
	"jms-core-x/internal/ws"
)

// APIWiring يحتوي كل شيء يخص HTTP API + WebSocket API.
type APIWiring struct {
	Router     http.Handler
	HTTPServer *http.Server
}

// BuildAPI يقوم بإنشاء API كاملة ومتصلة بكل الطبقات داخل Runtime.
func BuildAPI(
	cfg AppConfig,
	engines *Engines,
	stores *Stores,
	diagLayer *diagnostics.DiagnosticsLayer,
	hub *ws.Hub,
) (*APIWiring, error) {

	// --------------------------
	// 1) Readers (لخدمة Handlers)
	// --------------------------
	diagReader := readers.NewDiagnosticsReader(diagLayer)
	historyReader := readers.NewHistoryReader(stores.History)
	eventsReader := readers.NewEventsReader(stores.Events)
	alertsReader := readers.NewAlertsReader(stores.Alerts)
	snapshotReader := readers.NewSnapshotReader(engines.Final)

	// --------------------------
	// 2) بناء Handlers
	// --------------------------
	h := handlers.New(
		diagReader,
		historyReader,
		eventsReader,
		alertsReader,
		snapshotReader,
		hub,
	)

	// --------------------------
	// 3) Router (Block 18)
	// --------------------------
	router := api.NewRouter(h)

	// --------------------------
	// 4) HTTP Server
	// --------------------------
	srv := &http.Server{
		Addr:    cfg.HttpListenAddr,
		Handler: router,
	}

	return &APIWiring{
		Router:     router,
		HTTPServer: srv,
	}, nil
}
