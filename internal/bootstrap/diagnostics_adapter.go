package bootstrap

import (
    "jms-core-x/internal/api"
    "jms-core-x/internal/core/diagnostics"
)

// diagnosticsAdapter يربط DiagnosticsLayer الحقيقي بالـ HTTP API
type diagnosticsAdapter struct {
    layer *diagnostics.DiagnosticsLayer
}

func newDiagnosticsAdapter(layer *diagnostics.DiagnosticsLayer) api.DiagnosticsReader {
    return &diagnosticsAdapter{layer: layer}
}

func (a *diagnosticsAdapter) Sessions() []diagnostics.DeviceDiagnostics {
    return a.layer.ListDiagnostics()
}

func (a *diagnosticsAdapter) DeviceDiagnostics(id string) (diagnostics.DeviceDiagnostics, bool) {
    return a.layer.DeviceDiagnostics(id)
}

func (a *diagnosticsAdapter) LastSummary(id string) (diagnostics.PingSummary, bool) {
    return a.layer.LastSummary(id)
}
