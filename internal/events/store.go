// File: internal/events/store.go
package events

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DBStore هو التخزين الفعلي للأحداث فوق SQL DB.
type DBStore struct {
	db *sql.DB
}

// تأكيد أن DBStore يحقق EventReader
var _ EventReader = (*DBStore)(nil)

func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{db: db}
}

// SaveEvents يسجل مجموعة أحداث داخل Transaction واحدة
func (s *DBStore) SaveEvents(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("events: begin tx failed: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO events_log
(id, device_id, snapshot_id, type, field, old_value, new_value, timestamp)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("events: prepare insert failed: %w", err)
	}
	defer stmt.Close()

	for _, ev := range events {
		ts := parseTimestamp(ev.Timestamp)

		_, err := stmt.ExecContext(
			ctx,
			ev.ID,
			safeDBValue(ev.DeviceID),
			safeDBValue(ev.SnapshotID),
			safeDBValue(ev.Type),
			safeDBValue(ev.Field),
			safeDBValue(ev.OldValue),
			safeDBValue(ev.NewValue),
			ts,
		)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("events: insert event failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("events: commit tx failed: %w", err)
	}

	return nil
}

// ListEvents يرجع الأحداث مع الفلترة
func (s *DBStore) ListEvents(ctx context.Context, filter EventFilter) ([]Event, error) {
	query := `SELECT id, device_id, type, payload, created_at FROM events_log ORDER BY created_at DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.DeviceID, &e.Type, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
}

// ListDeviceEvents يرجع أحدث الأحداث المتعلقة بجهاز واحد
func (s *DBStore) ListDeviceEvents(ctx context.Context, deviceID string, limit int) ([]Event, error) {
	f := EventFilter{
		DeviceID: deviceID,
		Limit:    limit,
	}
	return s.ListEvents(ctx, f)
}

// ListSnapshotEvents ← يجب تطبيق snapshot_id هنا مباشرة
func (s *DBStore) ListSnapshotEvents(ctx context.Context, snapshotID string) ([]Event, error) {
	query := `
SELECT id, device_id, snapshot_id, type, field, old_value, new_value, timestamp
FROM events_log
WHERE snapshot_id = ?
ORDER BY timestamp DESC
`

	rows, err := s.db.QueryContext(ctx, query, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("events: list snapshot events failed: %w", err)
	}
	defer rows.Close()

	out := make([]Event, 0, 32)

	for rows.Next() {
		var ev Event
		var ts sql.NullInt64

		if err := rows.Scan(
			&ev.ID,
			&ev.DeviceID,
			&ev.SnapshotID,
			&ev.Type,
			&ev.Field,
			&ev.OldValue,
			&ev.NewValue,
			&ts,
		); err != nil {
			return nil, fmt.Errorf("events: scan snapshot event failed: %w", err)
		}

		if ts.Valid {
			ev.Timestamp = fromTimestamp(ts.Int64)
		} else {
			ev.Timestamp = time.Time{}
		}

		out = append(out, ev)
	}

	return out, rows.Err()
}

func joinConditions(conds []string) string {
	out := conds[0]
	for _, c := range conds[1:] {
		out += " AND " + c
	}
	return out
}
