package api

import (
    "log"
    "net/http"
    "runtime/debug"
    "time"
)

//
// -------------------------------
// Logging Middleware
// -------------------------------
//
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        start := time.Now()

        next.ServeHTTP(w, r)

        log.Printf("[HTTP] %s %s | %s", r.Method, r.URL.Path, time.Since(start))
    })
}

//
// -------------------------------
// CORS Middleware
// -------------------------------
//
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Expose-Headers", "Content-Length")

        // Handle preflight requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

//
// -------------------------------
// Recovery (Panic Protection)
// -------------------------------
//
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        defer func() {
            if err := recover(); err != nil {
                log.Printf("[HTTP] Panic Recovered: %v\nStack Trace:\n%s", err, debug.Stack())
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()

        next.ServeHTTP(w, r)
    })
}
