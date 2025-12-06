// File: internal/bootstrap/db_boot.go
// مسؤول عن فتح قاعدة البيانات وتشغيل الـ migrations من خلال Block 4 (db.Manager).
package bootstrap

import (
	"context"
	"database/sql"
	"fmt"

	"jms-core-x/internal/db"
)

// OpenDatabase يفتح اتصال DB واحد للعملية بالكامل باستخدام db.Manager.
// لا ينشئ أى جداول بنفسه؛ كل شئ يتم عن طريق db.runMigrations داخل Block 4.
func OpenDatabase(ctx context.Context, cfg AppConfig) (*sql.DB, error) {
	dbCfg := db.Config{
		Driver: cfg.DB.Driver,
		DSN:    cfg.DB.DSN,
	}
	sqlDB, err := db.Boot(ctx, dbCfg)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: db boot failed: %w", err)
	}
	return sqlDB, nil
}
