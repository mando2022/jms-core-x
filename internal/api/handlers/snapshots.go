package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "jms-core-x/internal/snapshots"
)

// SnapshotReader — الواجهة المطلوبة من Block 12
type SnapshotReader interface {
    ListSnapshots(ctx context.Context, filter snapshots.SnapshotFilter) ([]snapshots.Snapshot, error)
}

// SnapshotsHandler — يرجّع Snapshots للـ Frontend
func SnapshotsHandler(reader SnapshotReader) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        ctx := r.Context()
        q := r.URL.Query()

        deviceID := q.Get("device_id")

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
        limit := 50
        if raw := q.Get("limit"); raw != "" {
            fmt.Sscanf(raw, "%d", &limit)
        }

        filter := snapshots.SnapshotFilter{
            DeviceID: deviceID,
            Since:    sincePtr,
            Until:    untilPtr,
            Limit:    limit,
        }

        out, err := reader.ListSnapshots(ctx, filter)
        if err != nil {
            http.Error(w, "failed to load snapshots", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(out)
    }
}
