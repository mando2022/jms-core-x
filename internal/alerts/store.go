// File: jms-core-x/internal/alerts/store.go
// مسؤول عن عمليات القراءة / الكتابة الخاصة بالـ Alerts داخل DB.
// لا يحتوي على أي منطق Rules أو WebSocket أو HTTP.
package alerts

import (
	"context"
	"database/sql"
	"fmt"
)

// اسم جدول التنبيهات داخل قاعدة البيانات.
const alertsLogTable = "alerts_log"

// تعريف سكيمـا جدول alerts_log والفهارس طبقًا للبلو برنت.
const createAlertsLogTable = `
CREATE TABLE IF NOT EXISTS alerts_log (
    id          TEXT PRIMARY KEY,
    device_id   TEXT,
    event_id    TEXT,
    type        TEXT,
    description TEXT,
    severity    TEXT,
    timestamp   INTEGER
);`

const createIdxAlertsDevice = `
CREATE INDEX IF NOT EXISTS idx_alerts_device ON alerts_log(device_id);`

const createIdxAlertsTimestamp = `
CREATE INDEX IF NOT EXISTS idx_alerts_timestamp ON alerts_log(timestamp);`

// EnsureSchema يتأكد من وجود جدول alerts_log والفهارس المرتبطة به.
// يمكن استدعاؤه من طبقة الـ DB/Core عند تهيئة التطبيق.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		createAlertsLogTable,
		createIdxAlertsDevice,
		createIdxAlertsTimestamp,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("alerts: ensure schema failed: %w", err)
		}
	}

	return nil
}

// Store Interface كما هو معرف في Block16_Blueprint_v2.md.
// مسؤول فقط عن حفظ مجموعة من Alerts داخل DB.
type Store interface {
	SaveAlerts(ctx context.Context, alerts []Alert) error
}

// DBStore هو التنفيذ الفعلي لـ Store فوق *sql.DB.
// يطبّق أيضًا AlertReader لتمكين القراءة من نفس الطبقة.
type DBStore struct {
	db *sql.DB
}

// NewDBStore يبني DBStore جديد فوق اتصال قاعدة بيانات موجود.
func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{db: db}
}

// SaveAlerts يحفظ مجموعة من التنبيهات داخل جدول alerts_log باستخدام معاملة واحدة.
func (s *DBStore) SaveAlerts(ctx context.Context, alerts []Alert) error {
	if len(alerts) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("alerts: begin tx failed: %w", err)
	}

	stmt := `
INSERT INTO ` + alertsLogTable + `
    (id, device_id, event_id, type, description, severity, timestamp)
VALUES (?, ?, ?, ?, ?, ?, ?)`

	for _, a := range alerts {
		ts := toTimestamp(a.Timestamp)

		if _, err := tx.ExecContext(
			ctx,
			stmt,
			a.ID,
			a.DeviceID,
			a.EventID,
			a.Type,
			safeDBValue(a.Description),
			a.Severity,
			ts,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("alerts: insert alert %q failed: %w", a.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("alerts: commit tx failed: %w", err)
	}

	return nil
}

// ListAlerts يطبق AlertFilter على جدول alerts_log.
// يطابق واجهة AlertReader المستخدمة في Block 17 و Block 18.
func (s *DBStore) ListAlerts(ctx context.Context, filter AlertFilter) ([]Alert, error) {
	query := `SELECT id, device_id, kind, severity, message, created_at FROM alerts_log ORDER BY created_at DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Alert
	for rows.Next() {
		var a Alert
		if err := rows.Scan(&a.ID, &a.DeviceID, &a.Kind, &a.Severity, &a.Message, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
}

// joinConditions توحد conditions متعددة داخل WHERE clause واحدة باستخدام AND.
// منسوخة بنفس النمط المستخدم في Block 15 (events) لضمان الاتساق.
func joinConditions(conds []string) string {
	switch len(conds) {
	case 0:
		return ""
	case 1:
		return conds[0]
	default:
		out := conds[0]
		for i := 1; i < len(conds); i++ {
			out += " AND " + conds[i]
		}
		return out
	}
}

// ضمان أن DBStore يطبّق AlertReader.
var _ AlertReader = (*DBStore)(nil)
