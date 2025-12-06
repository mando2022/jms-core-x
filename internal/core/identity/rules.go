package identity

import "strings"

// canUseMAC يحدد ما إذا كان يمكن استخدام MAC كمفتاح أساسي للهوية.
// في Block 11 نعتبر أن أي MAC غير فارغ (وقد تم تطبيعه في Block 10)
// صالح للاستخدام كهوية أساسية.
func canUseMAC(mac string) bool {
	return strings.TrimSpace(mac) != ""
}

// hostGroupKey يحول hostname إلى مفتاح تجميعي ثابت للمطابقة.
// يستخدم للمساعدة في دمج الأجهزة التي تشترك في نفس الاسم.
func hostGroupKey(hostname string) string {
	s := strings.TrimSpace(hostname)
	if s == "" {
		return ""
	}
	return strings.ToLower(s)
}

// chooseVendor يختار Vendor نهائي من قائمة المرشحين.
// المنطق بسيط ومحافظ: أول قيمة غير فارغة بعد التنظيف.
func chooseVendor(candidates []string) string {
	for _, v := range candidates {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}
