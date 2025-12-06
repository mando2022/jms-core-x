// File: jms-core-x/internal/ws/hub.go
// Package ws implements Block 17 — WebSocket Hub.
// مسؤول عن إدارة قائمة العملاء وبث WSMessage إليهم عبر قنوات داخلية فقط.
// لا يوجد أي تعامل مباشر مع DB ولا أي منطق تحليلي هنا.
package ws

// Hub هو مركز الاتصالات لجميع عملاء WebSocket.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan WSMessage
	register   chan *Client
	unregister chan *Client
}

// NewHub ينشئ Hub جديد مع تهيئة القنوات والهياكل الداخلية.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan WSMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run يشغل الحلقة الرئيسية للـ Hub.
// يفضّل تشغيله داخل Goroutine مستقلة.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if client != nil {
				h.clients[client] = true
			}

		case client := <-h.unregister:
			if client != nil {
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
				}
			}

		case msg := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// العميل لا يستهلك الرسائل (بطيء أو عالق).
					// نحذفه لحماية الـ Hub.
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

// Register يسجل عميل جديد داخل الـ Hub.
func (h *Hub) Register(c *Client) {
	if h == nil || c == nil {
		return
	}
	h.register <- c
}

// Unregister يزيل عميل من الـ Hub وينظّف قناته.
func (h *Hub) Unregister(c *Client) {
	if h == nil || c == nil {
		return
	}
	h.unregister <- c
}

// Broadcast يدفع WSMessage إلى كل العملاء المسجلين.
func (h *Hub) Broadcast(msg WSMessage) {
	if h == nil {
		return
	}
	h.broadcast <- msg
}
