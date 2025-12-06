// File: jms-core-x/internal/events/schema.go
// مسؤول عن تعريف سكيمـا جدول events_log وضمان وجوده داخل DB.
package events

import (
	"context"
	"database/sql"
	"fmt"
)

// اسم جدول الأحداث داخل قاعدة البيانات.
const eventsLogTable = "events_log"

const createEventsLogTable = `
CREATE TABLE IF NOT EXISTS events_log (
    id          TEXT PRIMARY KEY,
    device_id   TEXT,
    snapshot_id TEXT,
    type        TEXT,
    field       TEXT,
    old_value   TEXT,
    new_value   TEXT,
    timestamp   INTEGER
);`

const createIdxEventsDevice = `
CREATE INDEX IF NOT EXISTS idx_events_device ON events_log(device_id);`

const createIdxEventsSnapshot = `
CREATE INDEX IF NOT EXISTS idx_events_snapshot ON events_log(snapshot_id);`

const createIdxEventsTimestamp = `
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events_log(timestamp);`

// EnsureSchema يتأكد من وجود جدول events_log والفهارس الأساسية.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		createEventsLogTable,
		createIdxEventsDevice,
		createIdxEventsSnapshot,
		createIdxEventsTimestamp,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("events: ensure schema failed: %w", err)
		}
	}

	return nil
}
