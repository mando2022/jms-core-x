// File: jms-core-x/internal/ws/events.go
// Helpers لبث Events القادمة من Block 15 (Event Store) عبر WebSocket.
package ws

import (
	"time"

	"jms-core-x/internal/events"
)

// EventToWS يحوّل Event واحد من Block 15 إلى WSMessage جاهز للبث.
func EventToWS(ev events.Event) WSMessage {
	return WSMessage{
		Type:      "event",
		Payload:   ev,
		Timestamp: time.Now(),
	}
}
