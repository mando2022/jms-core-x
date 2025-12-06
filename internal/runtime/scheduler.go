// File: internal/runtime/scheduler.go
// Scheduler يوفر primitive عام لتشغيل وظائف دورية (Jobs)
// بدون أي منطق متعلق بنوع البيانات أو الأجهزة.
package runtime

import (
	"context"
	"time"
)

// JobFunc هي الوظيفة الأساسية التي سيتم تشغيلها دوريًا.
type JobFunc func(ctx context.Context) error

// ScheduleConfig يحدد إعدادات Job واحد.
type ScheduleConfig struct {
	Name      string
	Interval  time.Duration
	JitterPct float64
	Enabled   bool
}

// RunJobLoop يشغّل Job دوريًا وفق ScheduleConfig داخل goroutine منفصلة.
// - لا يضيف أي منطق جديد على البيانات.
// - مسؤول فقط عن التكرار الزمني واستدعاء job().
func RunJobLoop(rt *RuntimeContext, sup *Supervisor, log Logger, cfg ScheduleConfig, job JobFunc) {
	if rt == nil || !cfg.Enabled || cfg.Interval <= 0 || job == nil {
		return
	}
	name := cfg.Name
	if name == "" {
		name = "job"
	}

	safeGo(rt, sup, name, func(ctx context.Context) {
		interval := cfg.Interval
		if interval <= 0 {
			interval = time.Second
		}

		ticker := time.NewTicker(withJitter(interval, cfg.JitterPct))
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := job(ctx); err != nil {
					if sup != nil {
						sup.ReportFailure(name, err)
					}
					if log != nil {
						log.Printf("runtime: job %s error: %v", name, err)
					}
				} else {
					if sup != nil {
						sup.ReportSuccess(name)
					}
				}
			}
		}
	})
}
