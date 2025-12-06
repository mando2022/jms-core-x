// File: jms-core-x/internal/change/models.go
// Package change implements Block 14 — Change Detection Engine.
// مسؤول عن تعريف واجهات القراءة + هياكل التغيرات + أحداث Block 14.

package change

import (
    "context"
    "time"

    "jms-core-x/internal/core/final"
    "jms-core-x/internal/history"
)

//
// ─────────────────────────────────────────────────────────────
//      ███████╗██╗  ██╗███╗   ██╗███████╗███████╗
//      ██╔════╝██║  ██║████╗  ██║██╔════╝██╔════╝
//      ███████╗███████║██╔██╗ ██║█████╗  ███████╗
//      ╚════██║██╔══██║██║╚██╗██║██╔══╝  ╚════██║
//      ███████║██║  ██║██║ ╚████║███████╗███████║
//      ╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚══════╝╚══════╝
// ─────────────────────────────────────────────────────────────
//

// FinalProvider هي الواجهة التي يحصل Block 14 من خلالها
// على FinalSnapshot (القادم من Block 12).
type FinalProvider interface {
    FinalSnapshot(ctx context.Context) final.FinalSnapshot
}

// HistoryReader يحتوي على كل وظائف قراءة التاريخ التي يحتاجها Block 14.
// يتم تطبيقه الآن داخل history.Store بعد تعديل Block 13.
type HistoryReader interface {
    DeviceHistory(ctx context.Context) ([]history.DeviceHistory, error)
    DeviceTimeline(ctx context.Context, deviceID string) ([]history.DeviceTimeline, error)
    LoadDeviceHistory(deviceID string) (history.DeviceHistory, error)
}

// SnapshotReader يوفّر SnapshotHistory السابق (آخر Snapshot مخزن).
type SnapshotReader interface {
    LoadLastSnapshot() (history.SnapshotHistory, error)
}

//
// ─────────────────────────────────────────────────────────────
//      هياكل الأحداث والتغيرات الخاصة ببلوك 14
// ─────────────────────────────────────────────────────────────
//

// أنواع الأحداث الأساسية التي ينتجها Block 14.
const (
    DEVICE_NEW         = "device_new"
    DEVICE_DISAPPEARED = "device_disappeared"
    IP_CHANGED         = "ip_changed"
    HOSTNAME_CHANGED   = "hostname_changed"
    VENDOR_CHANGED     = "vendor_changed"
    PORTS_CHANGED      = "ports_changed"
    SOURCES_CHANGED    = "sources_changed"
)

// ChangeEvent يمثل تغييرًا واحدًا حدث بين Snapshot سابق والحالة الحالية.
type ChangeEvent struct {
    Type     string
    DeviceID string
    Field    string
    OldValue interface{}
    NewValue interface{}
}

// ChangeSet يمثل المخرج الرسمي لبلوك 14.
// يحتوي على كل التغيرات التي حدثت بين Snapshot سابق والحالي.
type ChangeSet struct {
    SnapshotID string
    Timestamp  time.Time
    Events     []ChangeEvent
}
