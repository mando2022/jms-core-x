// File: jms-core-x/internal/events/models.go
// Package events implements Block 15 — Event Store.
// مسؤول عن تعريف نماذج الأحداث وواجهات القراءة الخاصة بها.
package events

import (
	"context"
	"time"
)

// Event يمثل سطر واحد داخل جدول events_log.
// هذا الهيكل هو الشكل المخزن لأي ChangeEvent قادم من Block 14.
type Event struct {
	ID         string    // معرف الحدث داخل Event Store (UUID أو hash ثابت)
	DeviceID   string    // معرف الجهاز كما هو مستخدم في Blocks 12–14
	SnapshotID string    // Snapshot الذي نتج عنه هذا الحدث (ChangeSet.SnapshotID)
	Type       string    // نوع الحدث (DEVICE_NEW, DEVICE_DISAPPEARED, FIELD_CHANGED, ...)
	Field      string    // اسم الحقل الذي تغيّر (ip, hostname, vlan, ...)
	OldValue   string    // القيمة القديمة بعد تحويلها لنص
	NewValue   string    // القيمة الجديدة بعد تحويلها لنص
	Timestamp  time.Time // وقت حدوث التغير (قادم من ChangeSet.Timestamp)
}

// EventFilter يحدد شروط البحث عند قراءة الأحداث من Event Store.
type EventFilter struct {
	DeviceID string     // اختياري: تقييد النتائج على جهاز معين
	Type     string     // اختياري: تقييد النتائج على نوع حدث معين
	Since    *time.Time // اختياري: من وقت معين
	Until    *time.Time // اختياري: إلى وقت معين
	Limit    int        // اختياري: أقصى عدد نتائج
}

// EventReader يحدد واجهات القراءة التي ستستخدمها البلوكات 16 و17 و18.
type EventReader interface {
	// ListEvents يرجع قائمة بالأحداث بناءً على الفلتر المعطى.
	ListEvents(ctx context.Context, filter EventFilter) ([]Event, error)

	// ListDeviceEvents يرجع آخر الأحداث الخاصة بجهاز واحد.
	ListDeviceEvents(ctx context.Context, deviceID string, limit int) ([]Event, error)

	// ListSnapshotEvents يرجع كل الأحداث التي تخص Snapshot واحد.
	ListSnapshotEvents(ctx context.Context, snapshotID string) ([]Event, error)
}
