// File: jms-core-x/internal/rawsnapshot/collector.go
package rawsnapshot

import (
	"context"
	"time"
)

type Collector struct {
	engine *Engine
}

func NewCollector(engine *Engine) *Collector {
	return &Collector{engine: engine}
}

func (c *Collector) Snapshot(ctx context.Context) RawSnapshot {
	if c == nil || c.engine == nil {
		return RawSnapshot{
			Time:   time.Now(),
			Errors: []error{SnapshotError{Source: "collector"}},
		}
	}
	return c.engine.BuildSnapshot(ctx)
}
