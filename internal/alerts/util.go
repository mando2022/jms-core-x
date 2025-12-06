// File: jms-core-x/internal/alerts/util.go
// دوال مساعدة داخل Block 16 فقط.
// مسؤول عن توليد AlertID، والتعامل مع timestamps، وبعض عمليات التطبيع.
package alerts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// generateAlertID يولّد معرف ثابت للـ Alert.
// يعتمد على (DeviceID + EventID + Type + Timestamp).
// الهدف: نفس المدخلات -> نفس المعرف، بدون تعارض مع EventID.
func generateAlertID(deviceID, eventID, alertType string, ts time.Time) string {
	h := sha256.New()

	payload := fmt.Sprintf("%s|%s|%s|%d",
		strings.TrimSpace(deviceID),
		strings.TrimSpace(eventID),
		strings.TrimSpace(alertType),
		ts.UTC().Unix(),
	)

	if _, err := h.Write([]byte(payload)); err != nil {
		// في حالة فشل نادر، نرجع fallback بسيط.
		return hex.EncodeToString([]byte(payload))
	}

	sum := h.Sum(nil)

	// نستخدم أول 16 بايت فقط لتقليل طول المعرف (32 hex chars).
	return hex.EncodeToString(sum[:16])
}

// stringify يحاول تحويل أي قيمة إلى نص بشكل آمن.
// يستخدم JSON إن أمكن، ثم يعود إلى fmt.Sprint.
func stringify(v any) string {
	if v == nil {
		return ""
	}

	if b, err := json.Marshal(v); err == nil {
		return string(b)
	}

	return fmt.Sprint(v)
}

// toTimestamp يحوّل time.Time إلى Unix timestamp (ثواني) في UTC.
func toTimestamp(t time.Time) int64 {
	return t.UTC().Unix()
}

// fromTimestamp يحوّل Unix timestamp (ثواني) إلى time.Time في UTC.
func fromTimestamp(ts int64) time.Time {
	return time.Unix(ts, 0).UTC()
}

// safeJSON يعيد JSON نظيف، أو "{}" لو النص فاضي.
func safeJSON(b []byte) string {
	s := strings.TrimSpace(string(b))
	if s == "" {
		return "{}"
	}
	return s
}

// safeDBValue يقوم بتنظيف النص قبل تخزينه في DB.
// هنا تطبيق بسيط: trim فقط، متسق مع util في Block 15.
func safeDBValue(s string) string {
	return strings.TrimSpace(s)
}
