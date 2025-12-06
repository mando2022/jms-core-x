package api

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

// HTTPServer يمثل السيرفر الرئيسي للـ HTTP API
type HTTPServer struct {
    srv     *http.Server
    address string
}

// Config إعدادات السيرفر
type Config struct {
    Address      string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
    EnableCORS   bool
}

// NewHTTPServer يبني السيرفر بالروتر النهائي
func NewHTTPServer(cfg Config, handler http.Handler) *HTTPServer {

    srv := &http.Server{
        Addr:         cfg.Address,
        Handler:      handler,
        ReadTimeout:  cfg.ReadTimeout,
        WriteTimeout: cfg.WriteTimeout,
        IdleTimeout:  cfg.IdleTimeout,
    }

    return &HTTPServer{
        srv:     srv,
        address: cfg.Address,
    }
}

// Start يقوم بتشغيل السيرفر فعليًا
func (h *HTTPServer) Start() {
    // طباعة مؤكدة 100% بوضوح
    fmt.Println("[HTTP] Server starting on", h.address)

    go func() {
        if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Println("[HTTP] Server failed:", err)
        }
    }()
}

// Shutdown يوقف السيرفر بأمان
func (h *HTTPServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}
}
