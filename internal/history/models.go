// File: jms-core-x/internal/history/models.go
package history

import "time"

// DeviceHistory يمثل السجل التاريخي المختصر لكل جهاز.
// يعكس مباشرة البيانات المخزنة في جدول devices_history.
type DeviceHistory struct {
    DeviceID  string
    FirstSeen time.Time
    LastSeen  time.Time
    SeenCount int
    MAC       string
    Vendor    string
}

// SnapshotHistory يمثل نسخة مؤرشفة من FinalSnapshot كما تم
// تخزينها في جدول snapshots_history.
type SnapshotHistory struct {
    SnapshotID string
    Timestamp  time.Time
    RawJSON    []byte
}

// DeviceTimeline يمثل نقطة زمنية واحدة في الـ Timeline لجهاز معين.
// الحقول مطابقة لجدول devices_timeline.
type DeviceTimeline struct {
    DeviceID  string
    Timestamp time.Time
    Source    string
}
