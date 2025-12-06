// File: internal/api/handlers_history.go
package api

import (
    "context"
    "encoding/json"
    "net/http"

    "jms-core-x/internal/history"
)

// HistoryReader تأتي من Block 18 (interface في router.go)
// ويتم تنفيذها عبر historyAPIAdapter في bootstrap/api_ws.go.
type HistoryReader interface {
    DeviceHistory(ctx context.Context) ([]history.DeviceHistory, error)
    DeviceTimeline(ctx context.Context, deviceID string) ([]history.DeviceTimeline, error)
    LoadDeviceHistory(deviceID string) (history.DeviceHistory, error)
}

// HandlersHistory يحتوي على كل دوال REST الخاصة ببلوك 18 (History API).
type HandlersHistory struct {
    reader HistoryReader
}

func NewHandlersHistory(r HistoryReader) *HandlersHistory {
    return &HandlersHistory{reader: r}
}

// GET /api/history/devices
func (h *HandlersHistory) HandleDevicesHistory(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    list, err := h.reader.DeviceHistory(ctx)
    if err != nil {
        http.Error(w, "failed to load device history: "+err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, list)
}

// GET /api/history/device/{id}
func (h *HandlersHistory) HandleDeviceHistory(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    deviceID := Param(r, "id")

    if deviceID == "" {
        http.Error(w, "missing device id", http.StatusBadRequest)
        return
    }

    d, err := h.reader.LoadDeviceHistory(deviceID)
    if err != nil {
        http.Error(w, "device not found: "+err.Error(), http.StatusNotFound)
        return
    }

    writeJSON(w, d)
}

// GET /api/history/device/{id}/timeline
func (h *HandlersHistory) HandleDeviceTimeline(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    deviceID := Param(r, "id")

    if deviceID == "" {
        http.Error(w, "missing device id", http.StatusBadRequest)
        return
    }

    list, err := h.reader.DeviceTimeline(ctx, deviceID)
    if err != nil {
        http.Error(w, "failed to load timeline: "+err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, list)
}

// ﹣﹣ Utility Writer ﹣﹣
func writeJSON(w http.ResponseWriter, v interface{
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(v)
}
