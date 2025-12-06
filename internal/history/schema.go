// File: jms-core-x/internal/history/schema.go
package history

import (
    "context"
    "database/sql"
    "fmt"
)

// أسماء الجداول المستخدمة داخل Block 13 فقط.
const (
    devicesHistoryTable   = "devices_history"
    snapshotsHistoryTable = "snapshots_history"
    devicesTimelineTable  = "devices_timeline"
)


// ensureTableExists — دالة اختيارية (غير مستخدمة) للتحقق من وجود جدول
// تُترك كما هي لأنها لا تدخل في منطق Block 13 الأساسي.
func ensureTableExists(ctx context.Context, db *sql.DB, table string) error {
    row := db.QueryRowContext(
        ctx,
        "SELECT name FROM sqlite_master WHERE type='table' AND name=?",
        table,
    )

    var name string
    if err := row.Scan(&name); err != nil {
        return fmt.Errorf("history: ensure table %s failed: %w", table, err)
    }

    return nil
}
