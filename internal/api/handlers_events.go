package api

import (
    "net/http"

    "jms-core-x/internal/events"
)

// EventsHandler يصدّر أحداث Block 15 عبر HTTP GET فقط.
type EventsHandler struct {
    reader events.EventReader
}

func NewEventsHandler(r events.EventReader) *EventsHandler {
    return &EventsHandler{reader: r}
}

// HandleEvents يرجع قائمة الأحداث بناءً على EventFilter مبني من query params.
func (h *EventsHandler) HandleEvents(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        methodNotAllowed(w, r, http.MethodGet)
        return
    }
    if h == nil || h.reader == nil {
        errorJSON(w, http.StatusInternalServerError, "event reader not configured")
        return
    }

    q := r.URL.Query()
    filter := events.EventFilter{
        DeviceID: q.Get("device_id"),
        Type:     q.Get("type"),
    }
    filter.Limit = parseLimit(r, "limit", 0, 1000)
    filter.Since = parseTimeQuery(r, "since")
    filter.Until = parseTimeQuery(r, "until")

    list, err := h.reader.ListEvents(r.Context(), filter)
    if err != nil {
        errorJSON(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondJSON(w, http.StatusOK, list)
}
