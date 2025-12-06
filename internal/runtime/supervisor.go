// File: internal/runtime/supervisor.go
// Supervisor مسؤول عن مراقبة حالة الـ loops (نجاح/فشل + آخر خطأ)
// دون أي منطق Business أو تفسير معنوي للأخطاء.
package runtime

import (
	"sync"
	"time"
)

// LoopStats يحتوي على إحصائيات أساسية لكل loop.
type LoopStats struct {
	LastError    error
	LastRun      time.Time
	FailCount    int
	SuccessCount int
}

// Supervisor يخزن إحصائيات الـ loops بطريقة آمنة على تعدد الـ goroutines.
type Supervisor struct {
	mu     sync.RWMutex
	stats  map[string]LoopStats
	logger Logger
}

// NewSupervisor يبني Supervisor جديد.
func NewSupervisor(log Logger) *Supervisor {
	return &Supervisor{
		stats:  make(map[string]LoopStats),
		logger: log,
	}
}

// ReportSuccess يتم استدعاؤها عند نجاح تنفيذ loop معين.
func (s *Supervisor) ReportSuccess(loopName string) {
	if s == nil || loopName == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	st := s.stats[loopName]
	st.LastRun = time.Now()
	st.SuccessCount++
	s.stats[loopName] = st
}

// ReportFailure يتم استدعاؤها عند فشل تنفيذ loop معين.
func (s *Supervisor) ReportFailure(loopName string, err error) {
	if s == nil || loopName == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	st := s.stats[loopName]
	st.LastRun = time.Now()
	st.FailCount++
	st.LastError = err
	s.stats[loopName] = st

	if s.logger != nil && err != nil {
		s.logger.Printf("runtime: loop %s failed: %v", loopName, err)
	}
}

// Snapshot يرجع نسخة من إحصائيات جميع الـ loops.
func (s *Supervisor) Snapshot() map[string]LoopStats {
	if s == nil {
		return map[string]LoopStats{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]LoopStats, len(s.stats))
	for k, v := range s.stats {
		out[k] = v
	}
	return out
}
