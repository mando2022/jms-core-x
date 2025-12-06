// File: jms-core-x/internal/core/final/models.go
package final

import (
    "context"
    "time"

    "jms-core-x/internal/core/identity"
    "jms-core-x/internal/core/normalization"
)

// FinalSnapshot هو المخرج الرسمي لبلوك 12 (Final Snapshot Engine).
// يمثل الصورة النهائية للأجهزة داخل النظام بعد تطبيق الهوية
// ثم تزيين البيانات (metadata) دون أي منطق هوية جديد.
type FinalSnapshot struct {
    Time    time.Time
    Devices []FinalDevice
}

// FinalDevice يمثل جهازًا واحدًا داخل FinalSnapshot.
// DeviceID يأتي كما هو من IdentitySnapshot ولا يتم تغييره هنا.
type FinalDevice struct {
    DeviceID  string   // كما هو من IdentitySnapshot (لا يتغير)
    MAC       string
    IPs       []string
    Hostnames []string
    Vendor    string
    Ports     []string

    // SeenSources توثّق المصادر التي ظهر من خلالها الجهاز
    // مثل: ["lldp","ubnt","ipscan","trusted"].
    SeenSources []string
}

// IdentityProvider هي الواجهة التي يحتاجها Block 12 من طبقة الهوية (Block 11).
// أي نوع يمكنه أن يوفّر IdentitySnapshot عبر هذه الدالة يمكن تمريره إلى FinalEngine.
type IdentityProvider interface {
    IdentitySnapshot(ctx context.Context) identity.IdentitySnapshot
}

// NormalizedProvider هي الواجهة التي يحتاجها Block 12 من طبقة التطبيع (Block 10).
// هذا المصدر اختياري ويستخدم فقط لتحسين بعض الـ metadata عند الحاجة.
type NormalizedProvider interface {
    NormalizedSnapshot(ctx context.Context) normalization.NormalizedSnapshot
}
