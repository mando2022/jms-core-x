package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"jms-core-x/internal/change"
)

type EventStoreEngine struct {
	changes change.ChangeSetProvider
	store   *DBStore
}

func NewEventStoreEngine(changes change.ChangeSetProvider, store *DBStore) *EventStoreEngine {
	return &EventStoreEngine{
		changes: changes,
		store:   store,
	}
}

func (e *EventStoreEngine) RunOnce(ctx context.Context) error {
	cs, err := e.changes.BuildChangeSet(ctx)
	if err != nil {
		return fmt.Errorf("event-store: changeset error: %w", err)
	}

	if len(cs.Events) == 0 {
		return nil
	}

	out := make([]Event, 0, len(cs.Events))

	for _, ev := range cs.Events {
		oldStr := normalizeValue(ev.OldValue)
		newStr := normalizeValue(ev.NewValue)

		id, err := generateEventID(
			ev.DeviceID,
			cs.SnapshotID,
			ev.Type,
			ev.Field,
			oldStr,
			newStr,
			cs.Timestamp,
		)
		if err != nil {
			return fmt.Errorf("event-store: id error: %w", err)
		}

		out = append(out, Event{
			ID:         id,
			DeviceID:   ev.DeviceID,
			SnapshotID: cs.SnapshotID,
			Type:       ev.Type,
			Field:      ev.Field,
			OldValue:   oldStr,
			NewValue:   newStr,
			Timestamp:  cs.Timestamp,
		})
	}

	return e.store.SaveEvents(ctx, out)
}

func normalizeValue(v interface{
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%g", val)
	case fmt.Stringer:
		return val.String()
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%T", val)
		}
		return string(b)
	}
}
}) string {
	if v == nil {
		return ""
	}

	switch x := v.(type) {
	case string:
		return x
	case []string:
		b, _ := json.Marshal(x)
		return string(b)
	case []int:
		b, _ := json.Marshal(x)
		return string(b)
	case time.Time:
		return x.UTC().Format(time.RFC3339)
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}
