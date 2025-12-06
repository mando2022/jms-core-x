package diagnostics

import (
    "context"
    "time"
)

// ProbeFunc موجود أصلاً فى session.go:
// type ProbeFunc func(ctx context.Context, ip string) (rtt time.Duration, ok bool)

var ICMPProbe ProbeFunc = func(ctx context.Context, ip string) (time.Duration, bool) {
    // هنا إما تعمل منطق Ping حقيقى
    // أو مبدئياً ترجع قيمة ثابتة لحين ما تكملها:
    return 0, true
}
