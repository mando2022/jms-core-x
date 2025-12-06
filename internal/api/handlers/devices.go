package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "jms-core-x/internal/unified"
)

// UnifiedProvider interface المتوقع من Block 20
type UnifiedProvider interface {
    BuildUnifiedSnapshot() (*unified.Snapshot, error)
}

// Response model للـ JSON
type DevicesResponse struct {
    Timestamp time.Time           `json:"timestamp"`
    Devices   []unified.Device    `json:"devices"`
}

// DevicesHandler — يرجّع Unified Snapshot
func DevicesHandler(provider UnifiedProvider) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        snapshot, err := provider.BuildUnifiedSnapshot()
        if err != nil {
            http.Error(w, "failed to build unified snapshot", http.StatusInternalServerError)
            return
        }

        resp := DevicesResponse{
            Timestamp: snapshot.Timestamp,
            Devices:   snapshot.Devices,
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
    }
}
