// File: internal/bootstrap/unified_provider.go
// Adapter رسمي لربط unified.Engine بواجهة UnifiedSnapshotProvider.
// هذا هو UnifiedSnapshotProvider الحقيقي الذي يحتاجه Block 10 و Block 20.

package bootstrap

import (
    "context"

    "jms-core-x/internal/core/normalization"
    "jms-core-x/internal/unified"
)

// unifiedSnapshotProvider هو الـ Adapter الحقيقي.
// يقوم بتحويل unified.Engine.Snapshot() → UnifiedSnapshot(ctx).
type unifiedSnapshotProvider struct {
    engine *unified.Engine
}

// NewUnifiedSnapshotProvider يبني UnifiedSnapshotProvider جاهز.
func NewUnifiedSnapshotProvider(engine *unified.Engine) normalization.UnifiedSnapshotProvider {
    return &unifiedSnapshotProvider{engine: engine}
}

// UnifiedSnapshot ينادى عليها Block 10 (Normalization Engine).
func (u *unifiedSnapshotProvider) UnifiedSnapshot(ctx context.Context) unified.Snapshot {
    // ctx غير مستخدم لأن unified.Engine لا يعتمد على سياق.
    _ = ctx
    return u.engine.Snapshot()
}
