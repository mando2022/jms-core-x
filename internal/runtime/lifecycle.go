// File: internal/runtime/lifecycle.go
// يحتوي على واجهات مساعدة لإدارة lifecycle لمكوّنات التشغيل.
// Block 19 نفسه (Orchestrator) يطبّق هذه الواجهة.
package runtime

import "context"

// Runner هي واجهة عامة لأي مكوّن يدعم Start / Shutdown.
type Runner interface {
	Start()
	Shutdown(ctx context.Context) error
}

// GracefulShutdown يستدعي Shutdown على مجموعة من الـ runners.
// هذه الدالة لا تحتوي على أي منطق خاص بـ JMS؛ مجرد أداة مساعدة عامة.
func GracefulShutdown(ctx context.Context, runners ...Runner) {
	for _, r := range runners {
		if r == nil {
			continue
		}
		_ = r.Shutdown(ctx)
	}
}
