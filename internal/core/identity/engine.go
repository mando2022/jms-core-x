package identity

import (
	"context"

	"jms-core-x/internal/core/normalization"
)

// IdentityEngine هو قلب Block 11.
// يستقبل NormalizedSnapshot من Block 10 عبر normalization.NormalizedProvider
// ويُنتج IdentitySnapshot بدون أي تعامل مع DB أو HTTP أو Events.
type IdentityEngine struct {
	normalized normalization.NormalizedProvider
}

// NewIdentityEngine ينشئ مثيل جديد من IdentityEngine.
func NewIdentityEngine(p normalization.NormalizedProvider) *IdentityEngine {
	return &IdentityEngine{normalized: p}
}

// Build هي الدالة الأساسية لبناء IdentitySnapshot.
// الخطوات العامة:
//  1) قراءة NormalizedSnapshot من Block 10.
//  2) تمرير قائمة الأجهزة إلى matcher لبناء IdentityDevice لكل كيان.
//  3) إعادة IdentitySnapshot النهائي بدون أي تأثير جانبي خارجي.
func (e *IdentityEngine) Build(ctx context.Context) IdentitySnapshot {
	snap := e.normalized.NormalizedSnapshot(ctx)

	devices := BuildIdentityDevices(snap.Devices)

	return IdentitySnapshot{
		Time:    snap.Time,
		Devices: devices,
	}
}
