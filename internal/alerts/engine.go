// File: jms-core-x/internal/alerts/engine.go
// مسؤول عن تشغيل محرك التنبيهات:
// قراءة Events من Block 15 → تطبيق Rules → تخزين Alerts في DB.
package alerts

import (
	"context"
	"fmt"

	"jms-core-x/internal/events"
)

// AlertsEngine هو المحرك الأساسي لبلوك 16.
// لا يعرف أي تفاصيل عن HTTP أو WebSocket أو UI.
type AlertsEngine struct {
	reader events.EventReader // قادم من Block 15 — Event Store
	store  Store              // تنفيذ لواجهة Store داخل نفس البلوك
}

// NewEngine يبني AlertsEngine جديد بقراءة من EventReader وكتابة إلى Store.
func NewEngine(reader events.EventReader, store Store) *AlertsEngine {
	return &AlertsEngine{
		reader: reader,
		store:  store,
	}
}

// Run يقوم بتشغيل محرك التنبيهات مرة واحدة:
//
// 1. يستدعي EventReader.ListEvents للحصول على قائمة Events حسب الفلتر.
// 2. يطبق applyRules على كل Event.
// 3. يجمع كل Alerts الناتجة.
// 4. يستدعي store.SaveAlerts لحفظها في DB.
//
// يمكن استدعاؤه بشكل دوري من Layer أعلى (Scheduler / Core).
func (e *AlertsEngine) Run(ctx context.Context, filter events.EventFilter) error {
	evs, err := e.reader.ListEvents(ctx, filter)
	if err != nil {
		return fmt.Errorf("alerts: list events failed: %w", err)
	}

	var alertsOut []Alert

	for _, ev := range evs {
		a := applyRules(ev)
		if len(a) == 0 {
			continue
		}
		alertsOut = append(alertsOut, a...)
	}

	if len(alertsOut) == 0 {
		// لا توجد Alerts جديدة — لا يعتبر هذا خطأ.
		return nil
	}

	if err := e.store.SaveAlerts(ctx, alertsOut); err != nil {
		return fmt.Errorf("alerts: save alerts failed: %w", err)
	}

	return nil
}
