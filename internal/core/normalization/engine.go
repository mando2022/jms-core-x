package normalization

import (
	"context"
	"time"

	"jms-core-x/internal/unified"
)

// NormalizationEngine هو قلب Block 10.
// يستقبل UnifiedSnapshot من Block 5 ويخرج NormalizedSnapshot
// بدون أي Merge جديد، بدون DB، وبدون أي Identity Logic.
type NormalizationEngine struct {
	unified UnifiedSnapshotProvider
}

// NewEngine ينشئ محرك التطبيع مع UnifiedSnapshotProvider محقون من الخارج.
func NewEngine(unified UnifiedSnapshotProvider) *NormalizationEngine {
	return &NormalizationEngine{unified: unified}
}

// تأكيد أن NormalizationEngine يحقق واجهة NormalizedProvider.
var _ NormalizedProvider = (*NormalizationEngine)(nil)

// NormalizedSnapshot هو المدخل الرسمي للاستخدام الخارجي (Block 11).
func (e *NormalizationEngine) NormalizedSnapshot(ctx context.Context) NormalizedSnapshot {
	return e.Build(ctx)
}

// Build هي الدالة الأساسية التي تطبق قواعد التطبيع على unified.Snapshot.
//
// ملاحظات مهمة:
// - لا يتم تعديل unified.Snapshot الأصلي.
// - لا يتم إنشاء أجهزة جديدة أو حذف أجهزة بناءً على منطق Merge.
// - كل UnifiedDevice ↔ NormalizedDevice (1:1).
func (e *NormalizationEngine) Build(ctx context.Context) NormalizedSnapshot {
	snap := e.unified.UnifiedSnapshot(ctx)

	out := NormalizedSnapshot{
		Time:    time.Now(),
		Devices: make([]NormalizedDevice, 0, len(snap.Devices)),
	}

	for _, d := range snap.Devices {
		nd := NormalizedDevice{
			ID:       d.ID,
			MAC:      NormalizeMAC(d.MAC),
			IP:       NormalizeIP(d.IP),
			Hostname: NormalizeHostname(d.Name),
			Vendor:   NormalizeVendor(d.Vendor),

			// ملاحظة: unified.UnifiedDevice الحالي لا يوفّر Port،
			// لذا يتم تركه فارغًا حاليًا حتى يتم تمديد Block 5 مستقبلًا.
			Port: "",

			UnifiedSource: string(primarySource(d.Sources)),
		}

		out.Devices = append(out.Devices, nd)
	}

	return out
}

// primarySource يختار مصدرًا واحدًا من قائمة Sources الخاصة بالجهاز
// بدون أي منطق Identity أو Merge. اختيار أول مصدر متاح يعتبر
// أبسط تمثيل آمن داخل نطاق Block 10.
func primarySource(sources []unified.Source) unified.Source {
	if len(sources) == 0 {
		return ""
	}
	return sources[0]
}
