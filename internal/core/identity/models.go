package identity

import "time"

// IdentitySnapshot هو المخرج الرئيسي لطبقة الهوية (Block 11).
// يحتوي على حالة الهوية النهائية للأجهزة بعد تطبيق قواعد
// المطابقة على NormalizedSnapshot القادم من Block 10.
type IdentitySnapshot struct {
	Time    time.Time
	Devices []IdentityDevice
}

// IdentityDevice يمثل كيان جهاز واحد بعد حل الهوية.
// كل جهاز حقيقي في الشبكة يجب أن يقابله IdentityDevice واحد فقط.
type IdentityDevice struct {
	DeviceID  string   // الهوية الثابتة للجهاز داخل النظام
	MAC       string   // MAC النهائي للجهاز (إن وجد)
	IPs       []string // كل الـ IPs المنسوبة لنفس الكيان
	Hostnames []string // كل أسماء الجهاز بعد التطبيع
	Vendor    string   // البائع/المصنّع النهائي بعد الترجيح
	Ports     []string // كل الـ Ports/Interfaces التي ظهر عليها الجهاز
	Sources   []string // قائمة المصادر (lldp/ubnt/ipscan/trusted/...)
}
