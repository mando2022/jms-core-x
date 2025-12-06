// File: jms-core-x/internal/ws/router.go
// Router مسؤول عن استقبال أنواع البيانات المختلفة (Snapshot / Event / Alert / Diagnostic)
// وتحويلها إلى WSMessage ثم دفعها إلى Hub. لا يكتب في DB ولا يغيّر أي بيانات.
package ws

import (
	"context"

	"jms-core-x/internal/alerts"
	"jms-core-x/internal/core/final"
	"jms-core-x/internal/events"
)

// Router يربط بين مصادر البيانات (Blocks 12/15/16/9) وبين Hub.
type Router struct {
	hub *Hub
}

// NewRouter ينشئ Router مربوط بـ Hub معين.
func NewRouter(hub *Hub) *Router {
	return &Router{hub: hub}
}

// PublishEvent يبث Event واحد من Block 15 عبر Hub.
func (r *Router) PublishEvent(ev events.Event) {
	if r == nil || r.hub == nil {
		return
	}
	r.hub.Broadcast(EventToWS(ev))
}

// PublishAlert يبث Alert واحد من Block 16 عبر Hub.
func (r *Router) PublishAlert(a alerts.Alert) {
	if r == nil || r.hub == nil {
		return
	}
	r.hub.Broadcast(AlertToWS(a))
}

// PublishSnapshot يبث FinalSnapshot جاهز من Block 12 عبر Hub.
func (r *Router) PublishSnapshot(snap final.FinalSnapshot) {
	if r == nil || r.hub == nil {
		return
	}
	r.hub.Broadcast(SnapshotToWS(snap))
}

// PublishSnapshotFromProvider يجلب FinalSnapshot الحالي من FinalProvider
// ثم يبثه عبر Hub بدون أي تعديل على البيانات.
func (r *Router) PublishSnapshotFromProvider(ctx context.Context, provider FinalProvider) {
	if r == nil || r.hub == nil || provider == nil {
		return
	}
	snap := provider.FinalSnapshot(ctx)
	r.hub.Broadcast(SnapshotToWS(snap))
}

// PublishDiagnostic يبث DiagnosticPayload واحد (قادماً من Block 9)
// داخل WSMessage من نوع "diagnostic".
func (r *Router) PublishDiagnostic(p DiagnosticPayload) {
	if r == nil || r.hub == nil {
		return
	}
	r.hub.Broadcast(DiagnosticToWS(p))
}
