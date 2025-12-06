package api

import (
    "encoding/json"
    "net/http"
    "time"
)

// ---------------------------
// Standard Success Response
// ---------------------------

type SuccessResponse struct {
    Status    string      `json:"status"`
    Timestamp time.Time   `json:"timestamp"`
    Data      interface{} `json:"data"`
}

func JSON(w http.ResponseWriter, data interface{
	return respondJSON(w, status, data)
}
}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    resp := SuccessResponse{
        Status:    "ok",
        Timestamp: time.Now().UTC(),
        Data:      data,
    }

    json.NewEncoder(w).Encode(resp)
}

// ---------------------------
// Standard Error Response
// ---------------------------

type ErrorResponse struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Error     string    `json:"error"`
    Code      int       `json:"code"`
}

func JSONError(w http.ResponseWriter, code int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)

    resp := ErrorResponse{
        Status:    "error",
        Timestamp: time.Now().UTC(),
        Error:     msg,
        Code:      code,
    }

    json.NewEncoder(w).Encode(resp)
}
