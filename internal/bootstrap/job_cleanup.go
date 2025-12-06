package bootstrap

import (
    "context"

    "jms-core-x/internal/runtime"
)

func NewCleanupJob(stores *Stores) runtime.JobFunc {
    return func(ctx context.Context) error {
        // تنظيف history
        if err := stores.History.Cleanup(ctx); err != nil {
            return err
        }

        // تنظيف events
        if err := stores.Events.Cleanup(ctx); err != nil {
            return err
        }

        // تنظيف alerts
        if err := stores.Alerts.Cleanup(ctx); err != nil {
            return err
        }

        return nil
    }
}
