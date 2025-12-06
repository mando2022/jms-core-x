// File: jms-core-x/internal/ws/client.go
// Client يمثل عميل WebSocket واحد متصل بالـ Hub.
// لا يحتوي أي منطق DB أو تحليل، فقط تمرير WSMessage إلى الاتصال.
package ws

import "jms-core-x/internal/ws/transport"

type Conn = transport.Conn


// Client يمثل عميل واحد مسجل داخل الـ Hub.
type Client struct {
	hub  *Hub
	conn Conn
	send chan WSMessage
}

// NewClient ينشئ Client جديد مربوط بـ Hub واتصال WebSocket مجرد (Conn).
func NewClient(hub *Hub, conn Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan WSMessage, 16), // Buffer صغير لتقليل الضغط على الـ Hub
	}
}

// SendQueue يعيد قناة الإرسال الداخلية للعميل.
// الاستخدام الأساسي يكون من الـ Hub داخل نفس الباكدج.
func (c *Client) SendQueue() chan<- WSMessage {
	return c.send
}

// Start يشغل حلقة الكتابة للعميل داخل Goroutine مستقلة.
func (c *Client) Start() {
	go c.writeLoop()
}

// writeLoop يستقبل الرسائل من قناة send ويكتبها كـ JSON على اتصال WebSocket.
func (c *Client) writeLoop() {
	defer func() {
		if c.hub != nil {
			c.hub.Unregister(c)
		}
		safeClose(c.conn)
	}()

	for msg := range c.send {
		if c.conn == nil {
			break
		}
		if err := c.conn.WriteJSON(msg); err != nil {
			// أي خطأ في الكتابة ينهي العميل، ويتم تنظيفه عبر defer.
			break
		}
	}
}
