// File: jms-core-x/internal/change/engine.go
package change

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "jms-core-x/internal/core/final"
    "jms-core-x/internal/history"
)

// Engine يدير اكتشاف التغيرات بين Snapshot سابق والحالي.
type Engine struct {
    finalProvider   FinalProvider    // Block 12
    historyReader   HistoryReader    // Block 13 (Store)
    snapshotReader  SnapshotReader   // Block 13 (Store)
}

func NewEngine(
    finalProvider FinalProvider,
    historyReader HistoryReader,
    snapshotReader SnapshotReader,
) *Engine {
    return &Engine{
        finalProvider:  finalProvider,
        historyReader:  historyReader,
        snapshotReader: snapshotReader,
    }
}

// BuildChangeSet ينشئ ChangeSet واحد بناءً على Snapshot حالية والسابق.
func (e *Engine) BuildChangeSet(ctx context.Context) (ChangeSet, error) {
    if e.finalProvider == nil {
        return ChangeSet{}, fmt.Errorf("change: final provider is nil")
    }

    currentSnap := e.finalProvider.FinalSnapshot(ctx)

    // قراءة SnapshotHistory السابق
    var lastSnap history.SnapshotHistory
    var prevSnap final.FinalSnapshot

    snapHist, err := e.snapshotReader.LoadLastSnapshot()
    if err == nil && len(snapHist.RawJSON) > 0 {
        lastSnap = snapHist
        _ = json.Unmarshal(snapHist.RawJSON, &prevSnap)
    }

    // مقارنة Snapshots
    events := diffFinalSnapshots(prevSnap, currentSnap)

    return ChangeSet{
        SnapshotID: lastSnap.SnapshotID,
        Timestamp:  time.Now(),
        Events:     events,
    }, nil
}
