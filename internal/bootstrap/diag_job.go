// File: internal/bootstrap/diag_job.go
// الغرض: تعريف JobFunc موحّد لتشغيل Auto-Ping V2 (DiagnosticsLayer)
// بحيث يشتغل من خلال Orchestrator (Block 19)، بدل loop عشوائى فى runtime20.

package bootstrap

import (
    "context"
    "time"

    "jms-core-x/internal/core/diagnostics"
    "jms-core-x/internal/runtime"
)

// NewDiagnosticsJob يبنى JobFunc جاهز لتوصيله مع OrchestratorDeps.DiagnosticsJob.
//
// الفكرة:
//   - كل مرة الـ Orchestrator يشغّل الـ Job (حسب DiagnosticsInterval),
//     نقرأ الـ Watchlist من DiagnosticsLayer ونبدأ Ping لكل جهاز.
//   - نفس المنطق اللى كان مكتوب كـ goroutine + ticker فى runtime20.go
//     لكن بشكل أنظف وتحت إدارة Block 19.
func NewDiagnosticsJob(layer *diagnostics.DiagnosticsLayer) runtime.JobFunc {
    if layer == nil {
        // احتياطى: لو مفيش layer، نرجّع JobFunc بيعمل NO-OP،
        // عشان ما نكسّرش الـ Orchestrator.
        return func(ctx context.Context) error {
            _ = ctx
            return nil
        }
    }

    return func(ctx context.Context) error {
        // NOTE:
        // هنا بننفّذ "دورة" واحدة من Auto-Ping:
        //   - نقرأ الأجهزة من الـ Watchlist.
        //   - نشغّل StartDevicePing لكل جهاز.
        //
        // الـ Orchestrator هو اللى هيكرّر تنفيذ الـ Job دى
        // حسب DiagnosticsInterval الموجود فى RuntimeConfig.

        ids := layer.WatchlistItems()
        if len(ids) == 0 {
            // مفيش أجهزة فى الـ Watchlist حاليًا — نرجع بدون Error.
            return nil
        }

        // الإعدادات الافتراضية الحالية (زى ما كانت فى runtime20.go):
        const maxAge = 30 * time.Second   // مدة صلاحية جلسة الـ Ping
        const interval = 2 * time.Second  // الفترة بين كل Ping والتانى

        for _, id := range ids {
            select {
            case <-ctx.Done():
                // لو الـ Runtime بيوقف، نخرج فورًا.
                return ctx.Err()
            default:
            }

            layer.StartDevicePing(ctx, id, maxAge, interval)
        }

        return nil
    }
}
