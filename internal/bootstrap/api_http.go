package bootstrap

import (
    "time"

    "jms-core-x/internal/api"
    "jms-core-x/internal/api/handlers"
    "jms-core-x/internal/change"
    "jms-core-x/internal/events"
    "jms-core-x/internal/alerts"
    "jms-core-x/internal/timeline"
    "jms-core-x/internal/snapshots"
    "jms-core-x/internal/unified"
)

// APIServices يمثل الخدمات الكاملة التي Block 18 يقدّمها
type APIServices struct {
    Router     http.Handler
    HTTPServer *api.HTTPServer
}

// BuildHTTPAPI يبني HTTP Router + Server
func BuildHTTPAPI(
    unifiedProvider unified.Provider,
    changeEngine change.ChangeSetProvider,
    eventsReader events.EventReader,
    alertsReader alerts.AlertsReader,
    timelineReader timeline.TimelineReader,
    snapshotReader snapshots.SnapshotReader,
    wsHandler http.Handler,
) (*APIServices, error) {

    // 1) تجميع الديبندنسيز للراوتر
    deps := api.RouterDeps{
        UnifiedProvider: unifiedProvider,
        ChangeEngine:    changeEngine,
        EventsReader:    eventsReader,
        AlertsReader:    alertsReader,
        TimelineStore:   timelineReader,
        SnapshotStore:   snapshotReader,
        WSUpgrade:       wsHandler,
    }

    // 2) إنشاء الراوتر
    router := api.BuildRouter(deps)

    // 3) بناء السيرفر
    cfg := api.Config{
        Address:      ":8080",
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  30 * time.Second,
        EnableCORS:   true,
    }

    server := api.NewHTTPServer(cfg, router)

    return &APIServices{
        Router:     router,
        HTTPServer: server,
    }, nil
}
