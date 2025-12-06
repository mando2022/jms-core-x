// File: jms-core-x/internal/history/engine.go
package history

import (
    "context"
    "fmt"

    "jms-core-x/internal/core/final"
)

// FinalProvider هي الواجهة التي يعتمد عليها Block 13
// للحصول على FinalSnapshot من Block 12.
type FinalProvider interface {
    FinalSnapshot(ctx context.Context) final.FinalSnapshot
}

// StoreInterface هو المخرج الرسمي لطبقة History Store داخل Block 13.
// (تم تحقيقه بالكامل داخل store.go بعد التعديلات).
type StoreInterface interface {
    StoreSnapshotHistory(ctx context.Context, snap final.FinalSnapshot) (string, error)
    UpdateDeviceHistory(ctx context.Context, snap final.FinalSnapshot) error
    AddTimelineEntries(ctx context.Context, snap final.FinalSnapshot) error
}

// HistoryEngine هو محرك Block 13 الفعلي.
// مسؤول عن تخزين Snapshot + History + Timeline في DB.
type HistoryEngine struct {
    provider FinalProvider
    store    StoreInterface
}

// NewHistoryEngine يبني HistoryEngine جديد.
func NewHistoryEngine(provider FinalProvider, store StoreInterface) *HistoryEngine {
    return &HistoryEngine{
        provider: provider,
        store:    store,
    }
}

// RunOnce يشغل دورة واحدة من عملية History:
// 1) يحصل على FinalSnapshot
// 2) يخزّنه كـ SnapshotHistory
// 3) يحدّث devices_history
// 4) يضيف timeline entries
func (h *HistoryEngine) RunOnce(ctx context.Context) error {
    if h.provider == nil {
        return fmt.Errorf("history: missing final snapshot provider")
    }
    if h.store == nil {
        return fmt.Errorf("history: missing store")
    }

    snap := h.provider.FinalSnapshot(ctx)

    // 1) تخزين Snapshot
    _, err := h.store.StoreSnapshotHistory(ctx, snap)
    if err != nil {
        return err
    }

    // 2) تحديث DeviceHistory
    if err := h.store.UpdateDeviceHistory(ctx, snap); err != nil {
        return err
    }

    // 3) timeline entries
    if err := h.store.AddTimelineEntries(ctx, snap); err != nil {
        return err
    }

    return nil
}
