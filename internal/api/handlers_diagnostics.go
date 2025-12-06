package api

import (
    "net/http"
)

// DiagnosticsHandler يقدّم واجهة Read-Only فوق DiagnosticsReader.
type DiagnosticsHandler struct {
    reader DiagnosticsReader
}

func NewDiagnosticsHandler(r DiagnosticsReader) *DiagnosticsHandler {
    return &DiagnosticsHandler{reader: r}
}

// HandleSessions يرجع جميع الـ diagnostics snapshots المتاحة.
func (h *DiagnosticsHandler) HandleSessions(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        methodNotAllowed(w, r, http.MethodGet)
        return
    }
    if h == nil || h.reader == nil {
        errorJSON(w, http.StatusInternalServerError, "diagnostics reader not configured")
        return
    }

    sessions := h.reader.Sessions()
    respondJSON(w, http.StatusOK, sessions)
}

// HandleSessionByID يرجع diagnostics snapshot لجهاز واحد.
func (h *DiagnosticsHandler) HandleSessionByID(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        methodNotAllowed(w, r, http.MethodGet)
        return
    }
    if h == nil || h.reader == nil {
        errorJSON(w, http.StatusInternalServerError, "diagnostics reader not configured")
        return
    }

    deviceID := pathParam(r, "/api/diagnostics/session/")
    if deviceID == "" {
        errorJSON(w, http.StatusBadRequest, "missing device id")
        return
    }

    diag, ok := h.reader.DeviceDiagnostics(deviceID)
    if !ok {
        errorJSON(w, http.StatusNotFound, "diagnostics not found")
        return
    }

    respondJSON(w, http.StatusOK, diag)
}

// HandlePingSummary يرجع آخر PingSummary لجهاز واحد.
func (h *DiagnosticsHandler) HandlePingSummary(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        methodNotAllowed(w, r, http.MethodGet)
        return
    }
    if h == nil || h.reader == nil {
        errorJSON(w, http.StatusInternalServerError, "diagnostics reader not configured")
        return
    }

    deviceID := pathParam(r, "/api/diagnostics/ping/")
    if deviceID == "" {
        errorJSON(w, http.StatusBadRequest, "missing device id")
        return
    }

    summary, ok := h.reader.LastSummary(deviceID)
    if !ok {
        errorJSON(w, http.StatusNotFound, "ping summary not found")
        return
    }

    respondJSON(w, http.StatusOK, summary)
}
