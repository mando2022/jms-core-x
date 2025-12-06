// File: internal/runtime/util.go
// دوال مساعدة عامة لبلوك 19 — لا تحتوي على أي منطق متعلق بالأجهزة أو البيانات.
package runtime

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Logger هي واجهة بسيطة تتوافق مع log.Logger القياسي.
// يتم استخدامها لتسجيل رسائل تشغيلية فقط.
type Logger interface {
	Printf(format string, args ...interface{})
}

// init يقوم بضبط seed عشوائي بسيط لاستخدامه في الـ jitter.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// withJitter يضيف نسبة بسيطة من العشوائية على Interval معين.
// الهدف هو توزيع الحمل على الزمن وتفادي الـ spikes.
func withJitter(base time.Duration, pct float64) time.Duration {
	if base <= 0 || pct <= 0 {
		return base
	}
	delta := time.Duration(float64(base) * pct)
	if delta <= 0 {
		return base
	}
	// قيمة بين -delta و +delta
	offset := time.Duration(rand.Int63n(int64(2*delta+1))) - delta
	return base + offset
}

// safeGo يساعد في تشغيل goroutine مع إدارة WaitGroup والتقاط أي panic.
// يتم تسجيل أي panic داخل Supervisor / Logger بدون إيقاف البرنامج بالكامل.
func safeGo(rt *RuntimeContext, sup *Supervisor, name string, fn func(ctx context.Context)) {
	if rt == nil || fn == nil {
		return
	}

	rt.WG.Add(1)
	go func() {
		defer rt.WG.Done()
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in loop %s: %v", name, r)
				if sup != nil {
					sup.ReportFailure(name, err)
				}
				if sup != nil && sup.logger != nil {
					sup.logger.Printf("runtime: recovered panic in %s: %v", name, r)
				}
			}
		}()

		fn(rt.Ctx)
	}()
}
