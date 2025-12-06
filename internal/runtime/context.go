// File: internal/runtime/context.go
// يوفر RuntimeContext الذي يجمع context + Cancel + WaitGroup
// ليتم استخدامه من جميع loops الخاصة ببلوك 19.
package runtime

import (
	"context"
	"sync"
)

// RuntimeContext هو الحاوية الأساسية لكل سياق التشغيل داخل Block 19.
type RuntimeContext struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	WG     *sync.WaitGroup
}

// NewRuntimeContext يبني RuntimeContext جديد من parent context.
func NewRuntimeContext(parent context.Context) *RuntimeContext {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	return &RuntimeContext{
		Ctx:    ctx,
		Cancel: cancel,
		WG:     &sync.WaitGroup{},
	}
}
