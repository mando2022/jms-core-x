// File: jms-core-x/internal/ws/transport/conn.go
// Package transport provides the low-level WebSocket connection abstraction
// shared by Block 17 (WS Hub) وطبقة Diagnostics القديمة (Block 9) عند الحاجة.
package transport

// Conn هو abstraction بسيط فوق اتصال WebSocket الفعلي.
// الهدف إن الكود الأعلى (Hub / Diagnostics) مايعرفش تفاصيل المكتبة المستخدمة.
type Conn interface {
    WriteJSON(v interface{}) error
    Close() error
}
