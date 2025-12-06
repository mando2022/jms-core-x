package handlers

import (
    "encoding/json"
    "net/http"
)

// HealthResponse الشكل النهائي للـ JSON
type HealthResponse struct {
    Status string `json:"status"`
}

// HealthHandler — نقطة اختبار جاهزية السيرفر
func HealthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    json.NewEncoder(w).Encode(HealthResponse{
        Status: "ok",
    })
}
