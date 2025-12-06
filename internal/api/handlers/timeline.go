package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "jms-core-x/internal/timeline"
)

// TimelineReader — الواجهة المطلوبة من Block 13
type TimelineReader interface {
    ListTimeline(ctx context.Context, filter timeline.TimelineFilter) ([]timeline.TimelineEntry, error)
}

// TimelineHandler — يرجّع Device Timeline للـ Frontend
func TimelineHandler(reader TimelineReader) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        ctx := r.Context()
        q := r.URL.Query()

        deviceID := q.Get("device_id")

        if deviceID == "" {
            http.Error(w, "device_id is required", http.StatusBadRequest)
            return
        }

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
        limit := 200
        if raw := q.Get("limit"); raw != "" {
            fmt.Sscanf(raw, "%d", &limit)
        }

        filter := timeline.TimelineFilter{
            DeviceID: deviceID,
            Since:    sincePtr,
            Until:    untilPtr,
            Limit:    limit,
        }

        entries, err := reader.ListTimeline(ctx, filter)
        if err != nil {
            http.Error(w, "failed to load timeline", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entries)
    }
}
