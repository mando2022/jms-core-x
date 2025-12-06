package api

import (
    "net/http"
    "time"

    "jms-core-x/internal/core/final"
)

// SnapshotHandler يعرض FinalSnapshot القادم من Block 12 كـ JSON جاهز للاستهلاك.
type SnapshotHandler struct {
    provider FinalSnapshotProvider
}

func NewSnapshotHandler(p FinalSnapshotProvider) *SnapshotHandler {
    return &SnapshotHandler{provider: p}
}

// snapshotResponse هو الغلاف الخارجي لـ FinalSnapshot فى الـ HTTP API.
type snapshotResponse struct {
    Timestamp time.Time           `json:"timestamp"`
    Devices   []final.FinalDevice `json:"devices"`
    Meta      map[string]interface{} `json:"meta,omitempty"`
}

// HandleSnapshot يرجع snapshot نهائي واحد من FinalSnapshotProvider.
func (h *SnapshotHandler) HandleSnapshot(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        methodNotAllowed(w, r, http.MethodGet)
        return
    }
    if h == nil || h.provider == nil {
        errorJSON(w, http.StatusInternalServerError, "snapshot provider not configured")
        return
    }

    snap := h.provider.FinalSnapshot(r.Context())

    resp := snapshotResponse{
        Timestamp: snap.Time,
        Devices:   snap.Devices,
        Meta: map[string]interface{}{
            "device_count": len(snap.Devices),
        },
    }

    respondJSON(w, http.StatusOK, resp)
}
