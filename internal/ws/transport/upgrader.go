// File: jms-core-x/internal/ws/transport/upgrader.go
package transport

import "net/http"

// Upgrader مسؤول عن ترقية اتصال HTTP عادي إلى WebSocket.
// التنفيذ الفعلي (gorilla/websocket أو غيره) يتم حقنه من الـ bootstrap (Block 20).
type Upgrader interface {
    Upgrade(w http.ResponseWriter, r *http.Request) (Conn, error)
}
