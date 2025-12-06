package bootstrap

import (
    "context"
    "database/sql"

    "jms-core-x/internal/alerts"
    "jms-core-x/internal/events"
    "jms-core-x/internal/history"
)

// Stores هو التجميع الرسمي لكل Stores الخاصة بالبلوكات:
//  - History Store  (Block 13)
//  - Event Store    (Block 15)
//  - Alerts Store   (Block 16)
type Stores struct {
    History *history.Store
    Events  events.Store
    Alerts  alerts.Store
}

// BuildStores يقوم ببناء جميع الـ Stores الخاصة بالطبقات العليا.
// هذه الدالة يتم استدعاؤها فقط من Block 20 (Runtime Wiring)
// داخل BuildRuntime().
func BuildStores(ctx context.Context, db *sql.DB) (*Stores, error) {

    // 1) History Store
    histStore, err := history.NewStore(ctx, db)
    if err != nil {
        return nil, err
    }

    // 2) Events Store
    evStore := events.NewDBStore(db)

    // 3) Alerts Store
    alStore := alerts.NewDBStore(db)

    // 4) تجميع كل شيء
    return &Stores{
        History: histStore,
        Events:  evStore,
        Alerts:  alStore,
    }, nil
}
