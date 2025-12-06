package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "jms-core-x/internal/alerts"
)

// AlertsReader — الواجهة القادمة من Block 16
type AlertsReader interface {
    ListAlerts(ctx context.Context, filter alerts.AlertFilter) ([]alerts.Alert, error)
}

// AlertsHandler — يرجّع Alerts للـ Frontend
func AlertsHandler(reader AlertsReader) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        ctx := r.Context()
        q := r.URL.Query()

        deviceID := q.Get("device_id")
        alertType := q.Get("type")
        severity := q.Get("severity")

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

        filter := alerts.AlertFilter{
            DeviceID: deviceID,
            Type:     alertType,
            Severity: severity,
            Since:    sincePtr,
            Until:    untilPtr,
            Limit:    limit,
        }

        out, err := reader.ListAlerts(ctx, filter)
        if err != nil {
            http.Error(w, "failed to load alerts", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(out)
    }
}
