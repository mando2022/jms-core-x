package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Migration represents a single schema change that can be applied
// to the underlying SQLite database.
type Migration struct {
	ID   int
	Name string
	Up   func(ctx context.Context, db *sql.DB) error
}

// migrations holds the ordered list of migrations for the project.
// Block 4 sets up the infrastructure; later blocks will extend this slice.
var migrations []Migration

// runMigrations ensures the base migrations table exists and then applies
// any pending migrations in order.
func runMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, baseMigrationsTableDDL); err != nil {
		return fmt.Errorf("db: create %s failed: %w", SchemaMigrationsTable, err)
	}

	applied, err := loadAppliedMigrations(ctx, db)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.ID] {
			continue
		}
		if m.Up == nil {
			return fmt.Errorf("db: migration %d (%s) has nil Up", m.ID, m.Name)
		}
		if err := m.Up(ctx, db); err != nil {
			return fmt.Errorf("db: migration %d (%s) failed: %w", m.ID, m.Name, err)
		}
		if err := recordMigration(ctx, db, m); err != nil {
			return err
		}
	}

	return nil
}

// baseMigrationsTableDDL defines the schema for the internal migrations table.
const baseMigrationsTableDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	applied_at TIMESTAMP NOT NULL
);
`

func loadAppliedMigrations(ctx context.Context, db *sql.DB) (map[int]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT id FROM "+SchemaMigrationsTable)
	if err != nil {
		return nil, fmt.Errorf("db: load applied migrations: %w", err)
	}
	defer rows.Close()

	result := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("db: scan applied migration id: %w", err)
		}
		result[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: iterate applied migrations: %w", err)
	}

	return result, nil
}

func recordMigration(ctx context.Context, db *sql.DB, m Migration) error {
	_, err := db.ExecContext(
		ctx,
		"INSERT INTO "+SchemaMigrationsTable+" (id, name, applied_at) VALUES (?, ?, ?)",
		m.ID,
		m.Name,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("db: record migration %d (%s): %w", m.ID, m.Name, err)
	}
	return nil
}
