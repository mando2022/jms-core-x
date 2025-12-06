package core

import (
	"context"
	"sync"
	"time"
)

// LayerName represents the logical name of a system layer
type LayerName string

const (
	LayerAuth       LayerName = "auth"
	LayerSettings   LayerName = "settings"
	LayerDiscovery  LayerName = "discovery"
	LayerUnified    LayerName = "unified"
	LayerPing       LayerName = "ping"
	LayerMap        LayerName = "map"
	LayerEvents     LayerName = "events"
	LayerBackup     LayerName = "backup"

	// ✨ Block 8 Addition
	LayerDiagnostics LayerName = "diagnostics"
)

type LayerStatus string

const (
	StatusCreated      LayerStatus = "created"
	StatusInitializing LayerStatus = "initializing"
	StatusRunning      LayerStatus = "running"
	StatusStopping     LayerStatus = "stopping"
	StatusStopped      LayerStatus = "stopped"
	StatusError        LayerStatus = "error"
)

type Config map[string]any

type Layer interface {
	Name() LayerName
	Init(ctx context.Context, cfg Config) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	Status() LayerStatus
	Config() Config
}

type BaseLayer struct {
	name   LayerName
	status LayerStatus
	cfg    Config

	mu      sync.RWMutex
	started time.Time
	err     error
}

func NewBaseLayer(name LayerName) BaseLayer {
	return BaseLayer{
		name:   name,
		status: StatusCreated,
		cfg:    Config{},
	}
}

func (b *BaseLayer) Name() LayerName {
	return b.name
}

func (b *BaseLayer) Config() Config {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.cfg
}

func (b *BaseLayer) setConfig(cfg Config) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if cfg == nil {
		b.cfg = Config{}
	} else {
		b.cfg = cfg
	}
}

func (b *BaseLayer) setStatus(s LayerStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = s
	if s == StatusRunning {
		b.started = time.Now()
	}
}

func (b *BaseLayer) Status() LayerStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

func (b *BaseLayer) setError(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.err = err
	if err != nil {
		b.status = StatusError
	}
}

func (b *BaseLayer) LastError() error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.err
}

func (b *BaseLayer) Uptime() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.started.IsZero() || b.status != StatusRunning {
		return 0
	}
	return time.Since(b.started)
}
