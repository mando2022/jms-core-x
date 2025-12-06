package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "jms-core-x/internal/events"
)

// EventReader — الواجهة اللي Block 15 يوفّرها
type EventReader interface {
    ListEvents(ctx context.Context, filter events.EventFilter) ([]events.Event, error)
}

// EventsHandler — يرجّع الأحداث من Event Store
func EventsHandler(reader EventReader) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        ctx := r.Context()
        q := r.URL.Query()

        // Extract query params
        deviceID := q.Get("device_id")
        eventType := q.Get("type")

        var sincePtr *time.Time
        var untilPtr *time.Time

        // since=
        if raw := q.Get("since"); raw != "" {
            t, err := time.Parse(time.RFC3339, raw)
            if err == nil {
                sincePtr = &t
            }
        }

        // until=
        if raw := q.Get("until"); raw != "" {
            t, err := time.Parse(time.RFC3339, raw)
            if err == nil {
                untilPtr = &t
            }
        }

        // limit=
        limit := 100
        if raw := q.Get("limit"); raw != "" {
            fmt.Sscanf(raw, "%d", &limit)
        }

        filter := events.EventFilter{
            DeviceID: deviceID,
            Type:     eventType,
            Since:    sincePtr,
            Until:    untilPtr,
            Limit:    limit,
        }

        evts, err := reader.ListEvents(ctx, filter)
        if err != nil {
            http.Error(w, "failed to load events", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(evts)
    }
}
