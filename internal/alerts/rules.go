// File: jms-core-x/internal/alerts/rules.go
// يحتوي على قواعد إطلاق التنبيهات بناءً على Events القادمة من Block 15.
// لا يحتوي على أي منطق DB أو WebSocket أو HTTP.
package alerts

import (
	"jms-core-x/internal/change"
	"jms-core-x/internal/events"
)

// applyRules يطبق كل قواعد التنبيه على Event واحد.
// يرجع قائمة بـ Alerts (قد تكون فارغة، أو أكثر من Alert واحد لنفس الحدث).
func applyRules(ev events.Event) []Alert {
	alerts := make([]Alert, 0, 2)

	// Rule 1 — جهاز جديد ظهر على الشبكة.
	if ev.Type == change.DEVICE_NEW {
		alerts = append(alerts, newAlertFromEvent(
			ev,
			ALERT_NEW_DEVICE,
			SEVERITY_INFO,
			"New device appeared",
		))
	}

	// Rule 2 — جهاز اختفى من الشبكة.
	if ev.Type == change.DEVICE_DISAPPEARED {
		alerts = append(alerts, newAlertFromEvent(
			ev,
			ALERT_DEVICE_DISAPPEARED,
			SEVERITY_WARNING,
			"Device disappeared",
		))
	}

	// Rule 3 — تغيير IP للجهاز.
	if ev.Type == change.IP_CHANGED {
		alerts = append(alerts, newAlertFromEvent(
			ev,
			ALERT_IP_CHANGED,
			SEVERITY_INFO,
			"Device IP changed",
		))
	}

	// Rule 4 — تغيير Vendor بشكل غير متوقع.
	if ev.Type == change.VENDOR_CHANGED {
		alerts = append(alerts, newAlertFromEvent(
			ev,
			ALERT_VENDOR_CHANGED,
			SEVERITY_CRITICAL,
			"Vendor changed unexpectedly",
		))
	}

	// Rule 5 — Too Many Events (High Frequency)
	// طبقًا للبلو برنت؛ المنطق التفصيلي يمكن توسيعه لاحقًا.
	if tooManyEventsInShortTime(ev.DeviceID) {
		alerts = append(alerts, newAlertFromEvent(
			ev,
			ALERT_HIGH_FREQUENCY,
			SEVERITY_WARNING,
			"High event rate detected",
		))
	}

	return alerts
}

// newAlertFromEvent يبني Alert واحد بناءً على Event معين
// مع تحديد نوع التنبيه ومستوى الخطورة والوصف.
func newAlertFromEvent(ev events.Event, alertType, severity, description string) Alert {
	return Alert{
		ID:          generateAlertID(ev.DeviceID, ev.ID, alertType, ev.Timestamp),
		DeviceID:    ev.DeviceID,
		EventID:     ev.ID,
		Type:        alertType,
		Description: description,
		Severity:    severity,
		Timestamp:   ev.Timestamp,
	}
}

// tooManyEventsInShortTime هو placeholder للقاعدة الخاصة بمعدل الأحداث العالي.
// طبقًا لقوانين المشروع (عدم التخمين)، نترك المنطق الفعلي لبلوك لاحق
// ونرجع false حاليًا لتجنب Alerts غير دقيقة.
func tooManyEventsInShortTime(deviceID string) bool {
	_ = deviceID
	return false
}
