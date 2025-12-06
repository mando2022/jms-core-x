package api

import (
    "net/http"

    "jms-core-x/internal/alerts"
)

// AlertsHandler يصدّر تنبيهات Block 16 عبر HTTP GET فقط.
type AlertsHandler struct {
    reader alerts.AlertReader
}

func NewAlertsHandler(r alerts.AlertReader) *AlertsHandler {
    return &AlertsHandler{reader: r}
}

// HandleAlerts يرجع قائمة التنبيهات بناءً على AlertFilter من query params.
func (h *AlertsHandler) HandleAlerts(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        methodNotAllowed(w, r, http.MethodGet)
        return
    }
    if h == nil || h.reader == nil {
        errorJSON(w, http.StatusInternalServerError, "alert reader not configured")
        return
    }

    q := r.URL.Query()
    filter := alerts.AlertFilter{
        DeviceID: q.Get("device_id"),
        Type:     q.Get("type"),
    }
    filter.Limit = parseLimit(r, "limit", 0, 1000)
    filter.Since = parseTimeQuery(r, "since")
    filter.Until = parseTimeQuery(r, "until")

    list, err := h.reader.ListAlerts(r.Context(), filter)
    if err != nil {
        errorJSON(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondJSON(w, http.StatusOK, list)
}

// HandleDeviceAlerts اختصار لقراءة تنبيهات جهاز واحد: /api/alerts/device/{id}
func (h *AlertsHandler) HandleDeviceAlerts(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        methodNotAllowed(w, r, http.MethodGet)
        return
    }
    if h == nil || h.reader == nil {
        errorJSON(w, http.StatusInternalServerError, "alert reader not configured")
        return
    }

    deviceID := pathParam(r, "/api/alerts/device/")
    if deviceID == "" {
        errorJSON(w, http.StatusBadRequest, "missing device id")
        return
    }

    filter := alerts.AlertFilter{
        DeviceID: deviceID,
    }
    filter.Limit = parseLimit(r, "limit", 0, 1000)

    list, err := h.reader.ListAlerts(r.Context(), filter)
    if err != nil {
        errorJSON(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondJSON(w, http.StatusOK, list)
}
