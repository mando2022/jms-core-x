// File: jms-core-x/internal/ws/alerts.go
// Helpers لبث Alerts القادمة من Block 16 (Alerts Engine) عبر WebSocket.
package ws

import (
	"time"

	"jms-core-x/internal/alerts"
)

// AlertToWS يحوّل Alert واحد من Block 16 إلى WSMessage جاهز للبث.
func AlertToWS(a alerts.Alert) WSMessage {
	return WSMessage{
		Type:      "alert",
		Payload:   a,
		Timestamp: time.Now(),
	}
}
