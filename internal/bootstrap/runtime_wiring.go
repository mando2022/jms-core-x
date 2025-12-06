// File: internal/bootstrap/runtime_wiring.go
// Block 20 — Runtime Bootstrap & Integration Layer
// مسؤول عن تجميع كل التبعيات (Stores + Engines + Hub + Jobs + API)
// وتحويلها إلى RuntimeWiring جاهز للتمرير إلى Orchestrator (Block 19).

package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"jms-core-x/internal/api"
	"jms-core-x/internal/api/handlers"
	"jms-core-x/internal/api/readers"
	"jms-core-x/internal/core/diagnostics"
	runtime "jms-core-x/internal/runtime"
	"jms-core-x/internal/unified"
	"jms-core-x/internal/ws"
)

// RuntimeWiring هو bundle يحتوي جميع مكونات الـ Runtime:
// (RuntimeConfig + OrchestratorDeps + DB + Stores + Engines + Hub)
type RuntimeWiring struct {
	Config runtime.RuntimeConfig
	Deps   runtime.OrchestratorDeps

	DB      *sql.DB
	Stores  *Stores
	Engines *Engines
	Hub     *ws.Hub
}

// BuildRuntime يبني RuntimeWiring بالكامل
// من AppConfig + RuntimeConfig بدون أي منطق أعمال إضافي.
func BuildRuntime(
	ctx context.Context,
	appCfg AppConfig,
	rtCfg runtime.RuntimeConfig,
) (*RuntimeWiring, error) {

	// -----------------------------------
	// 1) فتح قاعدة البيانات (Block 4)
	// -----------------------------------
	db, err := OpenDatabase(ctx, appCfg)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: open database failed: %w", err)
	}

	// -----------------------------------
	// 2) بناء Stores (Blocks 13–16)
	// -----------------------------------
	stores, err := BuildStores(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: build stores failed: %w", err)
	}

	// -----------------------------------
	// 3) Unified Engine + Provider (Block 8)
	// -----------------------------------
	unifiedEngine := unified.NewEngine()
	unifiedProvider := NewUnifiedSnapshotProvider(unifiedEngine)

	// -----------------------------------
	// 4) بناء المحركات العليا (Blocks 10–16)
	// -----------------------------------
	engines, err := BuildEngines(ctx, stores, unifiedProvider)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: build engines failed: %w", err)
	}

	// -----------------------------------
	// 5) WebSocket Hub (Block 17)
	// -----------------------------------
	hub := ws.NewHub()

	// -----------------------------------
	// 6) تجهيز OrchestratorDeps (Block 19)
	// -----------------------------------
	deps := runtime.OrchestratorDeps{
		HistoryEngine:  engines.History,
		EventsEngine:   engines.EventStore,
		AlertsEngine:   engines.Alerts,
		Hub:            hub,
		HTTPServer:     nil,
		SnapshotJob:    nil,
		DiagnosticsJob: nil,
		CleanupJob:     nil,
	}

	// -----------------------------------
	// 7) Phase 5 – Jobs Integration
	// -----------------------------------

	// Snapshot Job
	snapshotJob := NewSnapshotJob(engines)
	deps.SnapshotJob = snapshotJob

	// Diagnostics Layer + Diagnostics Job (Auto-Ping V2)
	diagLayer := diagnostics.NewDiagnosticsLayer()
	diagJob := NewDiagnosticsJob(diagLayer)
	deps.DiagnosticsJob = diagJob

	// Cleanup Job
	cleanupJob := NewCleanupJob(stores)
	deps.CleanupJob = cleanupJob

	// -----------------------------------
	// 8) Phase 6 – HTTP API Integration
	// -----------------------------------

	// Readers
	diagReader := readers.NewDiagnosticsReader(diagLayer)
	historyReader := readers.NewHistoryReader(stores.History)
	eventsReader := readers.NewEventsReader(stores.Events)
	alertsReader := readers.NewAlertsReader(stores.Alerts)
	snapshotReader := readers.NewSnapshotReader(engines.Final)

	// Handlers
	h := handlers.New(
		diagReader,
		historyReader,
		eventsReader,
		alertsReader,
		snapshotReader,
		hub,
	)

	// Router (Block 18)
	router := api.NewRouter(h)

	// HTTP Server
	httpServer := &http.Server{
		Addr:    appCfg.HttpListenAddr,
		Handler: router,
	}

	// تمرير الـ HTTP Server إلى deps
	deps.HTTPServer = httpServer

	// -----------------------------------
	// 9) تجميع RuntimeWiring النهائي
	// -----------------------------------
	wiring := &RuntimeWiring{
		Config:  rtCfg,
		Deps:    deps,
		DB:      db,
		Stores:  stores,
		Engines: engines,
		Hub:     hub,
	}

	return wiring, nil
}
