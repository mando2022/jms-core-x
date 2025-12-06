package bootstrap

import (
    "context"

    "jms-core-x/internal/alerts"
    "jms-core-x/internal/change"
    "jms-core-x/internal/events"
    "jms-core-x/internal/final"
    "jms-core-x/internal/history"
    "jms-core-x/internal/identity"
    "jms-core-x/internal/normalization"
)

// Engines يمثل تجميع كل المحركات العليا الخاصة بالبلوكات 10–16.
type Engines struct {
    Normalization *normalization.Engine
    Identity      *identity.Engine
    Final         *final.Engine
    History       *history.HistoryEngine
    Change        *change.Engine
    EventStore    *events.EventStoreEngine
    Alerts        *alerts.AlertsEngine
}

// BuildEngines يبني جميع الـ Engines اعتماداً على Stores و UnifiedSnapshotProvider.
// هذه الدالة يتم استدعاؤها فقط من Block 20 ضمن BuildRuntimeWiring().
func BuildEngines(
    ctx context.Context,
    stores *Stores,
    unifiedProvider normalization.UnifiedSnapshotProvider,
) (*Engines, error) {

    // 1) Normalization Engine (Block 10)
    normEngine := normalization.NewEngine(unifiedProvider)

    // 2) Identity Engine (Block 11)
    idEngine := identity.NewEngine(normEngine, unifiedProvider)

    // 3) Final Snapshot Engine (Block 12)
    finalEngine := final.NewEngine(normEngine, idEngine)

    // 4) History Engine (Block 13)
    histEngine := history.NewHistoryEngine(stores.History, finalEngine)

    // 5) Change Detection Engine (Block 14)
    changeEngine := change.NewEngine(unifiedProvider, stores.Events)

    // 6) Event Store Engine (Block 15)
    eventStoreEngine := events.NewEventStoreEngine(stores.Events, changeEngine)

    // 7) Alerts Engine (Block 16)
    alertsEngine := alerts.NewAlertsEngine(
        stores.Alerts,
        stores.Events,
        histEngine,
        idEngine,
        changeEngine,
    )

    // 8) تجميع كل المحركات في هيكل واحد
    return &Engines{
        Normalization: normEngine,
        Identity:      idEngine,
        Final:         finalEngine,
        History:       histEngine,
        Change:        changeEngine,
        EventStore:    eventStoreEngine,
        Alerts:        alertsEngine,
    }, nil
}
