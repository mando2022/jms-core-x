package normalization

import (
	"strings"
	"unicode"
)

// isZeroMAC يتحقق إن الـ MAC كله أصفار (بعد إزالة الفواصل).
func isZeroMAC(s string) bool {
	return s == "000000000000"
}

// isHexString يتحقق أن كل الحروف في السلسلة حروف hex صالحة.
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// safeLowerTrim → lowercase + trim لمسافات البداية والنهاية.
func safeLowerTrim(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// stripSuffixes يزيل أي suffix من القائمة إن وجد في نهاية السلسلة.
func stripSuffixes(s string, suffixes []string) string {
	for _, suf := range suffixes {
		if strings.HasSuffix(s, suf) {
			s = strings.TrimSuffix(s, suf)
		}
	}
	return strings.TrimSpace(s)
}

// cleanHostnameNoise يزيل الرموز الغير مرغوبة من hostname
// (مثل علامات الاقتباس وبعض الضوضاء البسيطة).
func cleanHostnameNoise(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "'", "")
	return s
}

// stripNonPrintable يزيل المحارف غير المطبوعة/الغريبة من النص.
func stripNonPrintable(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == ' ' || unicode.IsPrint(r) {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
