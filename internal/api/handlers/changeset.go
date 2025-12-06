package handlers

import (
    "encoding/json"
    "net/http"

    "jms-core-x/internal/change"
)

// ChangeSetProvider — الواجهة المطلوبة من Block 14
type ChangeSetProvider interface {
    BuildChangeSet(ctx context.Context) (*change.ChangeSet, error)
}

// Response JSON structure
type ChangeSetResponse struct {
    SnapshotID string               `json:"snapshot_id"`
    Timestamp  string               `json:"timestamp"`
    Events     []change.ChangeEvent `json:"events"`
}

// ChangeSetHandler — يرجع ChangeSet كامل
func ChangeSetHandler(provider ChangeSetProvider) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        ctx := r.Context()

        cs, err := provider.BuildChangeSet(ctx)
        if err != nil {
            http.Error(w, "failed to build changeset", http.StatusInternalServerError)
            return
        }

        resp := ChangeSetResponse{
            SnapshotID: cs.SnapshotID,
            Timestamp:  cs.Timestamp.UTC().Format(time.RFC3339),
            Events:     cs.Events,
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
    }
}
