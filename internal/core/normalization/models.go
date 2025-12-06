package normalization

import (
	"context"
	"time"

	"jms-core-x/internal/unified"
)

// NormalizedSnapshot هو المخرج الرئيسي لطبقة التطبيع (Block 10).
// يحتوي على قائمة الأجهزة بعد تنظيف الحقول الأساسية.
type NormalizedSnapshot struct {
	Time    time.Time
	Devices []NormalizedDevice
}

// NormalizedDevice يمثل جهاز واحد بعد التطبيع.
// ملاحظة: Port موجود طبقًا للـ Blueprint حتى لو UnifiedDevice
// الحالي لا يوفّر منفذًا صريحًا بعد.
type NormalizedDevice struct {
	ID            string
	MAC           string
	IP            string
	Hostname      string
	Vendor        string
	Port          string
	UnifiedSource string
}

// UnifiedSnapshotProvider هو مدخل Block 10 القادم من Block 5.
// أي طبقة (Unified Layer أو غيرها) يمكن أن تطبق هذه الواجهة لتمرير
// unified.Snapshot إلى NormalizationEngine.
type UnifiedSnapshotProvider interface {
	UnifiedSnapshot(ctx context.Context) unified.Snapshot
}

// NormalizedProvider هو المخرج الرسمي لـ Block 10.
// Block 11 (Identity) وأي طبقة أخرى يجب أن تعتمد على هذه الواجهة.
type NormalizedProvider interface {
	NormalizedSnapshot(ctx context.Context) NormalizedSnapshot
}
