// File: jms-core-x/internal/ws/snapshot.go
// Helpers لبث FinalSnapshot القادم من Block 12 عبر WebSocket.
package ws

import (
	"context"
	"time"

	"jms-core-x/internal/core/final"
)

// FinalProvider هي الواجهة الدنيا التي يحتاجها Block 17 من Block 12
// للحصول على FinalSnapshot بدون معرفة أي تفاصيل تخزين.
type FinalProvider interface {
	FinalSnapshot(ctx context.Context) final.FinalSnapshot
}

// SnapshotToWS يغلّف FinalSnapshot داخل WSMessage جاهز للبث.
func SnapshotToWS(snap final.FinalSnapshot) WSMessage {
	return WSMessage{
		Type:      "snapshot",
		Payload:   snap,
		Timestamp: time.Now(),
	}
}
