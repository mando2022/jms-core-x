package bootstrap

import (
    "context"
    "time"

    "jms-core-x/internal/runtime"
)

// NewSnapshotJob يبني وظيفة snapshot لتستخدم داخل Orchestrator.
func NewSnapshotJob(engines *Engines) runtime.JobFunc {
    return func(ctx context.Context) error {
        // 1) تنفيذ Final Snapshot
        snapshot, err := engines.Final.Execute(ctx)
        if err != nil {
            return err
        }

        // 2) إضافة snapshot إلى history
        if err := engines.History.Append(ctx, snapshot); err != nil {
            return err
        }

        // 3) مقارنة snapshot القديم بالجديد (Change Detection)
        if err := engines.Change.Process(ctx, snapshot); err != nil {
            return err
        }

        // 4) تخزين الأحداث الناتجة (Event Store)
        if err := engines.EventStore.Process(ctx); err != nil {
            return err
        }

        // 5) تشغيل Alerts Engine
        if err := engines.Alerts.Process(ctx); err != nil {
            return err
        }

        return nil
    }
}
