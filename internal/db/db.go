package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

// Config represents database boot configuration.
// For JMS Core X this is expected to point to a SQLite database,
// but the driver name and DSN are provided by the caller.
type Config struct {
	Driver string
	DSN    string
}

// Manager manages a single process-wide database connection.
type Manager struct {
	mu     sync.RWMutex
	db     *sql.DB
	cfg    Config
	booted bool
}

var defaultManager Manager

// Boot initializes the default database connection for the process.
// It is safe to call Boot multiple times with the same configuration;
// calling it again with a different configuration will return an error.
func Boot(ctx context.Context, cfg Config) (*sql.DB, error) {
	return defaultManager.Boot(ctx, cfg)
}

// Get returns the default *sql.DB instance if Boot has succeeded.
// It returns nil if Boot has not been called successfully yet.
func Get() *sql.DB {
	return defaultManager.Get()
}

// Close closes the default database connection if it is open.
func Close() error {
	return defaultManager.Close()
}

// Boot initializes the manager with the given configuration.
func (m *Manager) Boot(ctx context.Context, cfg Config) (*sql.DB, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.booted {
		// If configuration is identical, just return the existing handle.
		if cfg.Driver == m.cfg.Driver && cfg.DSN == m.cfg.DSN {
			return m.db, nil
		}
		return nil, fmt.Errorf("db: already booted with different configuration")
	}

	if cfg.Driver == "" {
		return nil, fmt.Errorf("db: driver is required")
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("db: DSN is required")
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("db: open failed: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("db: ping failed: %w", err)
	}

	if err := runMigrations(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	m.db = db
	m.cfg = cfg
	m.booted = true

	return db, nil
}

// Get returns the managed *sql.DB instance, or nil if not booted.
func (m *Manager) Get() *sql.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}

// Close closes the managed *sql.DB instance if it is open.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.booted || m.db == nil {
		return nil
	}

	err := m.db.Close()
	m.db = nil
	m.booted = false
	m.cfg = Config{}
	return err
}
