// File: internal/runtime/orchestrator.go
// Package runtime implements Block 19 — Runtime Orchestrator.
// مسؤول عن تشغيل محركات النظام (Engines) وفق الـ Blueprint الرسمي:
//   - تشغيل الـ loops (history / events / alerts / diagnostics / cleanup)
//   - تشغيل WebSocket Hub (Block 17)
//   - تشغيل HTTP API Router (Block 18) عبر http.Server
// بدون إضافة أي Business Logic أو منطق تحليلي جديد.
package runtime

import (
	"context"
	"errors"
	"net/http"
	"time"

	"jms-core-x/internal/alerts"
	"jms-core-x/internal/events"
	"jms-core-x/internal/history"
	"jms-core-x/internal/ws"
)

// RuntimeConfig يحدد إعدادات التشغيل الخاصة بالـ Runtime Orchestrator.
type RuntimeConfig struct {
	// الفواصل الزمنية لكل loop. يتم تجاهل أي Interval <= 0.
	SnapshotInterval    time.Duration
	HistoryInterval     time.Duration
	EventStoreInterval  time.Duration
	AlertsInterval      time.Duration
	DiagnosticsInterval time.Duration
	CleanupInterval     time.Duration

	// تمكين / تعطيل كل loop على حدة.
	EnableSnapshot    bool
	EnableHistory     bool
	EnableEvents      bool
	EnableAlerts      bool
	EnableDiagnostics bool
	EnableCleanup     bool
	EnableWSHub       bool
	EnableHTTP        bool

	// زمن إيقاف الـ HTTP Server بطريقة سلسة (Graceful Shutdown).
	HTTPShutdownTimeout time.Duration
}

// Orchestrator هو المكوّن الأساسي لبلوك 19.
// يقوم فقط بالتشغيل (orchestration) ولا يضيف أي منطق جديد على البيانات.
type Orchestrator struct {
	rt   *RuntimeContext
	sup  *Supervisor
	cfg  RuntimeConfig
	log  Logger

	// المحركات القادمة من البلوكات السابقة.
	historyEngine *history.HistoryEngine
	eventsEngine  *events.EventStoreEngine
	alertsEngine  *alerts.AlertsEngine

	// مكونات الـ I/O العليا.
	hub    *ws.Hub
	server *http.Server

	// وظائف اختيارية يتم تمريرها من طبقة أعلى بدون منطق داخل Block 19.
	SnapshotJob    JobFunc
	DiagnosticsJob JobFunc
	CleanupJob     JobFunc
}

// OrchestratorDeps يجمع كل التبعيات التي يحتاجها Block 19 للتشغيل.
type OrchestratorDeps struct {
	HistoryEngine *history.HistoryEngine
	EventsEngine  *events.EventStoreEngine
	AlertsEngine  *alerts.AlertsEngine
	Hub           *ws.Hub
	HTTPServer    *http.Server
	SnapshotJob   JobFunc
	DiagnosticsJob JobFunc
	CleanupJob     JobFunc
}

// NewOrchestrator يبني Orchestrator جديد اعتمادًا على RuntimeConfig و OrchestratorDeps.
func NewOrchestrator(parent context.Context, cfg RuntimeConfig, log Logger, deps OrchestratorDeps) *Orchestrator {
	if parent == nil {
		parent = context.Background()
	}

	rt := NewRuntimeContext(parent)
	sup := NewSupervisor(log)

	o := &Orchestrator{
		rt:            rt,
		sup:           sup,
		cfg:           cfg,
		log:           log,
		historyEngine: deps.HistoryEngine,
		eventsEngine:  deps.EventsEngine,
		alertsEngine:  deps.AlertsEngine,
		hub:           deps.Hub,
		server:        deps.HTTPServer,
		SnapshotJob:   deps.SnapshotJob,
		DiagnosticsJob: deps.DiagnosticsJob,
		CleanupJob:     deps.CleanupJob,
	}

	return o
}

// Start يطلق جميع الـ loops والمحركات وفقًا للـ RuntimeConfig.
func (o *Orchestrator) Start() {
	if o == nil || o.rt == nil {
		return
	}

	// 1) Snapshot Loop (اختياري بالكامل).
	if o.cfg.EnableSnapshot && o.cfg.SnapshotInterval > 0 && o.SnapshotJob != nil {
		RunJobLoop(o.rt, o.sup, o.log, ScheduleConfig{
			Name:      "snapshot_loop",
			Interval:  o.cfg.SnapshotInterval,
			JitterPct: 0.05,
			Enabled:   true,
		}, o.SnapshotJob)
	}

	// 2) History Loop — يعتمد فقط على HistoryEngine.
	if o.cfg.EnableHistory && o.cfg.HistoryInterval > 0 && o.historyEngine != nil {
		RunJobLoop(o.rt, o.sup, o.log, ScheduleConfig{
			Name:      "history_loop",
			Interval:  o.cfg.HistoryInterval,
			JitterPct: 0.05,
			Enabled:   true,
		}, func(ctx context.Context) error {
			return o.historyEngine.RunOnce(ctx)
		})
	}

    // 3) Event Store Loop — يقوم بدورة Change Detection + Event Store.
    if o.cfg.EnableEvents && o.cfg.EventStoreInterval > 0 && o.eventsEngine != nil {
    RunJobLoop(o.rt, o.sup, o.log, ScheduleConfig{
        Name:      "event_store_loop",
        Interval:  o.cfg.EventStoreInterval,
        JitterPct: 0.05,
        Enabled:   true,
    }, func(ctx context.Context) error {
        return o.eventsEngine.RunOnce(ctx)
    })
}


	// 4) Alerts Loop — يعتمد على AlertsEngine فقط.
	if o.cfg.EnableAlerts && o.cfg.AlertsInterval > 0 && o.alertsEngine != nil {
		// Filter يتم تمريره من طبقة أعلى لاحقًا إن لزم؛
		// هنا نستخدم فلتر فارغ/افتراضي دون منطق Business.
		filter := events.EventFilter{}
		RunJobLoop(o.rt, o.sup, o.log, ScheduleConfig{
			Name:      "alerts_loop",
			Interval:  o.cfg.AlertsInterval,
			JitterPct: 0.05,
			Enabled:   true,
		}, func(ctx context.Context) error {
			return o.alertsEngine.Run(ctx, filter)
		})
	}

	// 5) Diagnostics Loop (اختياري) — يتم تمرير منطق الـ Job من طبقة أعلى.
	if o.cfg.EnableDiagnostics && o.cfg.DiagnosticsInterval > 0 && o.DiagnosticsJob != nil {
		RunJobLoop(o.rt, o.sup, o.log, ScheduleConfig{
			Name:      "diagnostics_loop",
			Interval:  o.cfg.DiagnosticsInterval,
			JitterPct: 0.05,
			Enabled:   true,
		}, o.DiagnosticsJob)
	}

	// 6) Cleanup Loop (اختياري) — مسئول عن تنظيف موارد Runtime فقط.
	if o.cfg.EnableCleanup && o.cfg.CleanupInterval > 0 && o.CleanupJob != nil {
		RunJobLoop(o.rt, o.sup, o.log, ScheduleConfig{
			Name:      "cleanup_loop",
			Interval:  o.cfg.CleanupInterval,
			JitterPct: 0.10,
			Enabled:   true,
		}, o.CleanupJob)
	}

	// 7) WebSocket Hub (Block 17) — Loop مستمرة واحدة.
	if o.cfg.EnableWSHub && o.hub != nil {
		safeGo(o.rt, o.sup, "ws_hub", func(ctx context.Context) {
			// Hub.Run() لا يدعم context، لكنه يستمع على قنوات داخلية فقط.
			// عند shutdown سيتم إنهاء الـ goroutine عبر إلغاء الـ context وإغلاق الموارد من طبقة أعلى إن لزم.
			o.hub.Run()
		})
	}

	// 8) HTTP API Router (Block 18) عبر http.Server.
	if o.cfg.EnableHTTP && o.server != nil {
		safeGo(o.rt, o.sup, "http_server", func(ctx context.Context) {
			err := o.server.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				if o.sup != nil {
					o.sup.ReportFailure("http_server", err)
				}
				if o.log != nil {
					o.log.Printf("runtime: http server error: %v", err)
				}
			}
		})
	}
}

// Shutdown يقوم بإيقاف جميع الـ loops والمحركات بشكل منظم.
func (o *Orchestrator) Shutdown(ctx context.Context) error {
	if o == nil || o.rt == nil {
		return nil
	}

	// إيقاف الـ HTTP Server أولاً بحيث لا يتم قبول طلبات جديدة.
	if o.cfg.EnableHTTP && o.server != nil {
		timeout := o.cfg.HTTPShutdownTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}

		shCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if err := o.server.Shutdown(shCtx); err != nil {
			if o.log != nil {
				o.log.Printf("runtime: http server shutdown error: %v", err)
			}
		}
	}

	// إلغاء الـ context الرئيسي لجميع الـ loops.
	o.rt.Cancel()

	// انتظار انتهاء جميع الـ goroutines المسجلة.
	o.rt.WG.Wait()

	return nil
}

// SupervisorStatsSnapshot يرجع نسخة من حالة الـ loops الحالية.
// يمكن استهلاكها من طبقة أعلى (API / Diagnostics) دون أي تعديل للبيانات.
func (o *Orchestrator) SupervisorStatsSnapshot() map[string]LoopStats {
	if o == nil || o.sup == nil {
		return map[string]LoopStats{}
	}
	return o.sup.Snapshot()
}
