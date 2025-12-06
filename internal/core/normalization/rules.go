package normalization

import (
	"net"
	"strings"
)

// NormalizeMAC يطبّق قواعد MAC حسب الـ Blueprint:
// - إزالة ":" و "-"
// - lowercase
// - التحقق من الطول 12 حقل hex
// - رفض MAC كله أصفار
// - إعادة الصياغة aa:bb:cc:dd:ee:ff
// - invalid → "".
func NormalizeMAC(in string) string {
	if in == "" {
		return ""
	}

	s := strings.ToLower(strings.TrimSpace(in))
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "")

	if len(s) != 12 {
		return ""
	}
	if isZeroMAC(s) {
		return ""
	}
	if !isHexString(s) {
		return ""
	}

	// إعادة بناء الشكل القياسي.
	return s[0:2] + ":" + s[2:4] + ":" + s[4:6] + ":" +
		s[6:8] + ":" + s[8:10] + ":" + s[10:12]
}

// NormalizeIP يطبّق قواعد IP حسب الـ Blueprint:
// - Trim للمسافات
// - رفض "0.0.0.0"
// - IPv4 parsing فقط
// - IP invalid → "".
func NormalizeIP(in string) string {
	s := strings.TrimSpace(in)
	if s == "" || s == "0.0.0.0" {
		return ""
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return ""
	}

	ipv4 := ip.To4()
	if ipv4 == nil {
		return ""
	}

	return ipv4.String()
}

// NormalizeVendor يطبّق قواعد Vendor حسب الـ Blueprint:
// - lowercase + trim
// - إزالة الرموز الزائدة البسيطة
// - "unknown" → ""
// - "ubnt" → "Ubiquiti".
func NormalizeVendor(in string) string {
	if in == "" {
		return ""
	}

	s := safeLowerTrim(in)
	s = stripNonPrintable(s)

	if s == "" || s == "unknown" {
		return ""
	}

	if s == "ubnt" {
		return "Ubiquiti"
	}

	return s
}

// NormalizeHostname يطبّق قواعد Hostname حسب الـ Blueprint:
// - lowercase + trim
// - إزالة suffix مثل ".local" و ".lan"
// - تنظيف ضوضاء UBNT/LLDP البسيطة.
// أي نتيجة فارغة/ضوضاء تعاد كـ "".
func NormalizeHostname(in string) string {
	if in == "" {
		return ""
	}

	s := safeLowerTrim(in)
	if s == "" {
		return ""
	}

	s = stripSuffixes(s, []string{".local", ".lan"})
	s = cleanHostnameNoise(s)

	if s == "" || s == "-" {
		return ""
	}

	return s
}

// NormalizePort يطبّق قواعد Port حسب الـ Blueprint:
// - توحيد الصيغ الأساسية:
//     eth1 → Ethernet1
//     lan0 → LAN0
// - إزالة الرموز غير المرغوبة البسيطة.
// ملاحظة: ما فيش source فعلي للـ Port حاليًا في unified.UnifiedDevice
// لذا يتم استدعاء هذه الدالة على القيمة (إن وجدت) بدون اختراع بيانات جديدة.
func NormalizePort(in string) string {
	s := safeLowerTrim(in)
	if s == "" {
		return ""
	}

	// إزالة بعض الرموز الشائعة غير المرغوبة.
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, ";", "")

	if strings.HasPrefix(s, "eth") {
		return "Ethernet" + s[3:]
	}
	if strings.HasPrefix(s, "lan") {
		return "LAN" + s[3:]
	}

	return s
}
