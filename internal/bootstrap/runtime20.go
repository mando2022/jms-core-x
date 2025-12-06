package bootstrap

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gorilla/websocket"

    "jms-core-x/internal/api"
    "jms-core-x/internal/alerts"
    "jms-core-x/internal/change"
    "jms-core-x/internal/config"
    "jms-core-x/internal/core/diagnostics"
    "jms-core-x/internal/events"
    "jms-core-x/internal/history"
    "jms-core-x/internal/snapshots"
    "jms-core-x/internal/timeline"
    "jms-core-x/internal/unified"
    "jms-core-x/internal/ws"
    "jms-core-x/internal/ws/transport"
)


// Run هو نقطة الدخول الرسمية لتشغيل Block 20 Runtime.
func Run() error {

    ctx := context.Background()

    //
    // 1) تحميل الإعدادات
    //
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("runtime20: failed to load config: %w", err)
    }

    //
    // 2) تحميل Stores: Events + Alerts + History + Timeline + Snapshots
    //
    stores, err := LoadStores(cfg)
    if err != nil {
        return fmt.Errorf("runtime20: failed to load stores: %w", err)
    }

    //
    // 3) Unified Engine (Blocks 6–8)
    //
    unifiedEngine, err := buildUnifiedEngine(cfg)
    if err != nil {
        return fmt.Errorf("runtime20: failed to build unified engine: %w", err)
    }

    unifiedProvider := NewUnifiedSnapshotProvider(unifiedEngine)

    //
    // 4) بناء المحركات Blocks 10 → 16
    //
    engines, err := BuildEngines(ctx, stores, unifiedProvider)
    if err != nil {
        return fmt.Errorf("runtime20: failed to build engines: %w", err)
    }

    //
    // ────────────────────────────────────────────────
    // 5) DiagnosticsLayer الحقيقي (Auto-Ping)
    // ────────────────────────────────────────────────
    //

    probeFunc := diagnostics.ICMPProbe // لو ICMPProbe مش موجود أقدر أجهزهولك

    diagLayer := diagnostics.NewDiagnosticsLayer(
        unifiedProvider,
        probeFunc,
        stores.Events, // لربط Ping Events مع Event Store
    )

    diagReader := newDiagnosticsAdapter(diagLayer) // يربط مع API

    //
    // ────────────────────────────────────────────────
    // 6) WebSocket Hub (Block 17)
    // ────────────────────────────────────────────────
    //
    hub := ws.NewHub()
    go hub.Run()

    wsUpgrader := &transport.GorillaUpgrader{
        Upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool { return true },
        },
    }

    //
    // ────────────────────────────────────────────────
    // 7) Build Block 18 HTTP API
    // ────────────────────────────────────────────────
    //
    apiDeps := api.Deps{
        Final:       unifiedProvider,
        History:     stores.History,
        Events:      stores.Events,
        Alerts:      stores.Alerts,
        Diagnostics: diagReader,
        Hub:         hub,
        WSUpgrader:  wsUpgrader,
        SnapshotStore: stores.Snapshots,
        TimelineStore: stores.Timeline,
        Version:     "20.0",
    }

    router := api.NewRouter(apiDeps)

    serverCfg := api.Config{
        Address:      ":" + cfg.HTTP.Port,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  30 * time.Second,
        EnableCORS:   true,
    }

    httpServer := api.NewHTTPServer(serverCfg, router)

    //
    // ────────────────────────────────────────────────
    // 8) تشغيل Loops + HTTP Server
    // ────────────────────────────────────────────────
    //
    go func() {
        fmt.Println("Block 18 API: http://localhost:" + cfg.HTTP.Port)
        httpServer.Start()
    }()

    go engines.Change.StartLoop(ctx)
    go engines.Events.StartLoop(ctx)
    go engines.Alerts.StartLoop(ctx)
    go unifiedEngine.StartLoop(ctx)

    //
    // ────────────────────────────────────────────────
    // 9) تفعيل Auto-Ping (Watchlist Loop)
    // ────────────────────────────────────────────────
    //

    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for {
            <-ticker.C

            ids := diagLayer.WatchlistItems()

            for _, id := range ids {
                diagLayer.StartDevicePing(
                    context.Background(),
                    id,
                    30*time.Second,
                    2*time.Second,
                )
            }
        }
    }()

    //
    // 10) Prevent exit
    //
    select {}
}
