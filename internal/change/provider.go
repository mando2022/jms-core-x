// File: internal/change/provider.go
// Block 14 – Change Detection
//
// هذا الملف يعرّف الواجهة (interface) التي يستخدمها Block 15 / Block 16
// للحصول على أحدث ChangeSet من طبقة التغييرات (Block 14).
// أي نوع في package change يحقق هذه الدالة:
//     BuildChangeSet(ctx) (ChangeSet, error)
// يعتبر ChangeSetProvider صالح.

package change

import "context"

// ChangeSetProvider يمثل أي مكوّن قادر على بناء ChangeSet واحد
// (عادةً من آخر FinalSnapshot + History).
type ChangeSetProvider interface {
	// BuildChangeSet يبني ChangeSet حالي استنادًا إلى الحالة
	// الداخلية (History + Final Snapshot أو أي مصادر أخرى داخل Block 14).
	//
	// في حال النجاح يرجع ChangeSet جاهز للتخزين في Event Store.
	// في حال الخطأ يرجع خطأ واحد يوضح سبب الفشل.
	BuildChangeSet(ctx context.Context) (ChangeSet, error)
}
