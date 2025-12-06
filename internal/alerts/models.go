// File: jms-core-x/internal/alerts/models.go
// Package alerts implements Block 16 — Alerts Engine.
// مسؤول عن تعريف نماذج التنبيهات وواجهات القراءة الخاصة بها.
package alerts

import (
	"context"
	"time"
)

// Alert يمثل سطر واحد داخل جدول alerts_log.
// هذا الهيكل مطابق للبلو برنت الرسمي v2 (Block16_Blueprint_v2.md).
type Alert struct {
	ID          string    // معرف التنبيه (يولد عبر generateAlertID)
	DeviceID    string    // معرف الجهاز كما هو مستخدم في Event Store
	EventID     string    // معرف الحدث الأصلي داخل events_log
	Type        string    // نوع التنبيه (ALERT_NEW_DEVICE, ALERT_IP_CHANGED, ...)
	Description string    // وصف بشري قصير للتنبيه
	Severity    string    // مستوى الخطورة (info, warning, critical)
	Timestamp   time.Time // وقت إنشاء / تسجيل التنبيه
}

// Alert Types — مجموعة الأنواع القياسية للتنبيهات داخل Block 16.
const (
	ALERT_NEW_DEVICE         = "alert_new_device"
	ALERT_DEVICE_DISAPPEARED = "alert_device_disappeared"
	ALERT_IP_CHANGED         = "alert_ip_changed"
	ALERT_VENDOR_CHANGED     = "alert_vendor_changed"
	ALERT_HIGH_FREQUENCY     = "alert_high_event_rate"
)

// Alert Severity Levels — مستويات الخطورة القياسية للـ Alerts.
const (
	SEVERITY_INFO     = "info"
	SEVERITY_WARNING  = "warning"
	SEVERITY_CRITICAL = "critical"
)

// AlertFilter يحدد شروط البحث عند قراءة التنبيهات من alerts_log.
// مطابق لما هو معرف في Block16_Blueprint_v2.md.
type AlertFilter struct {
	DeviceID string     // اختياري: تقييد النتائج على جهاز معين
	Type     string     // اختياري: نوع تنبيه معين
	Since    *time.Time // اختياري: من وقت معين (Timestamp >= Since)
	Until    *time.Time // اختياري: إلى وقت معين (Timestamp <= Until)
	Limit    int        // اختياري: أقصى عدد نتائج
}

// AlertReader هي الواجهة التي ستستخدمها البلوكات 17 و18
// لقراءة التنبيهات من Block 16 بدون معرفة تفاصيل التخزين.
type AlertReader interface {
	ListAlerts(ctx context.Context, filter AlertFilter) ([]Alert, error)
}
