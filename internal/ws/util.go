// File: jms-core-x/internal/ws/util.go
// Package ws implements shared helpers and data models for Block 17 — WebSocket Hub.
// هذه الوحدة لا تتعامل مع DB ولا تحلل البيانات، فقط أدوات مساعدة للـ WebSocket Hub.
package ws

import "time"

// WSMessage هو الشكل القياسي لأي رسالة يتم بثّها من Hub إلى العملاء.
type WSMessage struct {
	Type      string      `json:"type"`      // snapshot / event / alert / diagnostic
	Payload   interface{} `json:"payload"`   // FinalSnapshot / Event / Alert / DiagnosticPayload / ...
	Timestamp time.Time   `json:"timestamp"` // وقت إنتاج الرسالة للبث
}

// DiagnosticPayload يمثل حمولة رسائل الـ Diagnostics القادمة من Block 9.
type DiagnosticPayload struct {
	SessionID string      `json:"session_id"`
	Data      interface{} `json:"data"`
}

// DiagnosticToWS يغلّف بيانات الـ Diagnostics داخل WSMessage جاهز للبث.
func DiagnosticToWS(p DiagnosticPayload) WSMessage {
	return WSMessage{
		Type:      "diagnostic",
		Payload:   p,
		Timestamp: time.Now(),
	}
}

// safeClose تغلق اتصال WebSocket مع تجاهل أي خطأ.
// تستخدم لضمان تنظيف الاتصال بدون التأثير على استقرار الـ Hub.
func safeClose(c Conn) {
	if c == nil {
		return
	}
	_ = c.Close()
}
