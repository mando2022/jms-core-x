package api

import (
    "net/http"

    "jms-core-x/internal/ws"
    "jms-core-x/internal/ws/transport"
)

// WSHandler مسؤول عن ترقية الاتصال إلى WebSocket وتمريره إلى Hub (Block 17).
// ملاحظة: تم فصل نوع الـ WebSocketUpgrader عن diagnostics واستخدام واجهة
// عامة من الحزمة ws/transport حتى لا يعتمد Block 18 على Block 9 القديم.
type WSHandler struct {
    hub      *ws.Hub
    upgrader transport.Upgrader
}

// NewWSHandler ينشئ Handler جديد لطلبات WebSocket مع Hub و Upgrader مجرد.
func NewWSHandler(hub *ws.Hub, upgrader transport.Upgrader) *WSHandler {
    return &WSHandler{
        hub:      hub,
        upgrader: upgrader,
    }
}

// HandleWebSocket يقوم بترقية الاتصال إلى WebSocket وتسجيل Client جديد في Hub.
func (h *WSHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    if h.hub == nil {
        errorJSON(w, http.StatusInternalServerError, "websocket hub not configured")
        return
    }
    if h.upgrader == nil {
        errorJSON(w, http.StatusInternalServerError, "websocket upgrader not configured")
        return
    }

    conn, err := h.upgrader.Upgrade(w, r)
    if err != nil {
        // أي أخطاء upgrade يتم التعامل معها بواسطة الـ upgrader نفسه (مثل كتابة الرد المناسب).
        return
    }

    client := ws.NewClient(h.hub, conn)
    h.hub.Register(client)
    client.Start()
}
