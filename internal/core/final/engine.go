// File: jms-core-x/internal/core/final/engine.go
package final

import (
    "context"
    "time"

    "jms-core-x/internal/core/normalization"
)

// FinalEngine هو المحرك المسؤول عن بناء FinalSnapshot.
// يعتمد على:
//   - IdentityProvider (إجباري)
//   - NormalizedProvider (اختياري)
// ولا يقوم بأي منطق هوية جديد أو تعامل مع DB/HTTP/WebSocket.
type FinalEngine struct {
    identity   IdentityProvider
    normalized NormalizedProvider
}

// NewFinalEngine ينشئ FinalEngine جديدًا.
// يمكن تمرير normalized = nil في حالة عدم الحاجة لأي تزيين إضافي من Block 10.
func NewFinalEngine(id IdentityProvider, norm NormalizedProvider) *FinalEngine {
    return &FinalEngine{
        identity:   id,
        normalized: norm,
    }
}

// Build يبني FinalSnapshot معتمدًا على IdentitySnapshot القادم من Block 11
// (و NormalizedSnapshot اختياريًا للتزيين).
func (e *FinalEngine) Build(ctx context.Context) FinalSnapshot {
    // الحصول على IdentitySnapshot من Block 11 (إجباري).
    idSnap := e.identity.IdentitySnapshot(ctx)

    // الحصول على NormalizedSnapshot من Block 10 (اختياري).
    var normSnap *normalization.NormalizedSnapshot
    if e.normalized != nil {
        ns := e.normalized.NormalizedSnapshot(ctx)
        normSnap = &ns
    }

    devices := make([]FinalDevice, 0, len(idSnap.Devices))

    // في حالة وجود NormalizedSnapshot، نبني فهرسًا سريعًا للمساعدة في تزيين البيانات.
    var normIndex map[string][]normalization.NormalizedDevice
    if normSnap != nil {
        normIndex = buildNormalizedIndex(normSnap)
    }

    for _, d := range idSnap.Devices {
        // بناء FinalDevice مبدئي من IdentityDevice كما هو.
        fd := FinalDevice{
            DeviceID:    d.DeviceID,
            MAC:         d.MAC,
            IPs:         normalizeList(d.IPs),
            Hostnames:   normalizeList(d.Hostnames),
            Vendor:      d.Vendor,
            Ports:       normalizeList(d.Ports),
            SeenSources: normalizeList(d.Sources),
        }

        // تطبيق تزيين metadata من NormalizedSnapshot (إن وجد) دون أي منطق هوية جديد.
        if normSnap != nil && normIndex != nil {
            decorateFromNormalized(&fd, d, normIndex)
        }

        devices = append(devices, fd)
    }

    return FinalSnapshot{
        Time:    time.Now(),
        Devices: devices,
    }
}

// buildNormalizedIndex يبني فهرسًا بسيطًا للأجهزة المطَبَّعة
// لمساعدة merge.go على الوصول السريع للـ NormalizedDevice المناسب.
// لا يحتوي على أي منطق Identity جديد، ويعتمد غالبًا على MAC كمدخل رئيسي.
func buildNormalizedIndex(normSnap *normalization.NormalizedSnapshot) map[string][]normalization.NormalizedDevice {
    index := make(map[string][]normalization.NormalizedDevice, len(normSnap.Devices))
    for _, nd := range normSnap.Devices {
        key := nd.MAC
        if key == "" {
            // لو MAC فارغ يمكن استخدام مفتاح آخر بسيط مثل IP،
            // لكن دون تغيير منطق الهوية؛ الهدف هنا تزيين فقط.
            key = nd.IP
        }
        if key == "" {
            continue
        }
        index[key] = append(index[key], nd)
    }
    return index
}
