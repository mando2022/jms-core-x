// File: internal/bootstrap/config.go
// Package bootstrap implements Block 20 — Runtime Bootstrap & Integration Layer.
// مسؤول عن تحميل إعدادات التشغيل وتحويلها إلى RuntimeConfig بلا أى منطق Business.
package bootstrap

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	runtime "jms-core-x/internal/runtime"
)

// AppConfig يعكس إعدادات التشغيل كما حددها Blueprint Block 20 PRO.
type AppConfig struct {
	DB struct {
		Driver string
		DSN    string
	}
	HTTP struct {
		Port int
	}
	Runtime struct {
		EnableHistory     bool
		EnableEvents      bool
		EnableAlerts      bool
		EnableWSHub       bool
		EnableHTTP        bool
		EnableDiagnostics bool
		EnableCleanup     bool

		HistoryInterval     time.Duration
		EventsInterval      time.Duration
		AlertsInterval      time.Duration
		DiagnosticsInterval time.Duration
		CleanupInterval     time.Duration

		HTTPShutdownTimeout time.Duration
	}
}

// LoadConfigFromEnv يحمّل الإعدادات من متغيرات البيئة فقط.
// لا يضع أى قيم افتراضية خاصة بالـ Business Logic؛
// القيم الغير موجودة يتم التعامل معها كخطأ (ماعدا الفواصل الزمنية و الـ Flags الاختيارية).
func LoadConfigFromEnv(ctx context.Context) (AppConfig, error) {
	_ = ctx // حالياً لا نستخدمه لكن الاحتفاظ به يحافظ على التوقيع المستقبلى.

	var cfg AppConfig

	// DB
	cfg.DB.Driver = os.Getenv("JMS_DB_DRIVER")
	cfg.DB.DSN = os.Getenv("JMS_DB_DSN")
	if cfg.DB.Driver == "" {
		return cfg, fmt.Errorf("bootstrap: JMS_DB_DRIVER is required")
	}
	if cfg.DB.DSN == "" {
		return cfg, fmt.Errorf("bootstrap: JMS_DB_DSN is required")
	}

	// HTTP
	portStr := os.Getenv("JMS_HTTP_PORT")
	if portStr == "" {
		return cfg, fmt.Errorf("bootstrap: JMS_HTTP_PORT is required")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		return cfg, fmt.Errorf("bootstrap: invalid JMS_HTTP_PORT %q", portStr)
	}
	cfg.HTTP.Port = port

	// Runtime flags (اختيارية، الافتراض = false)
	cfg.Runtime.EnableHistory = parseBoolEnv("JMS_ENABLE_HISTORY", false)
	cfg.Runtime.EnableEvents = parseBoolEnv("JMS_ENABLE_EVENTS", false)
	cfg.Runtime.EnableAlerts = parseBoolEnv("JMS_ENABLE_ALERTS", false)
	cfg.Runtime.EnableWSHub = parseBoolEnv("JMS_ENABLE_WSHUB", false)
	cfg.Runtime.EnableHTTP = parseBoolEnv("JMS_ENABLE_HTTP", false)
	cfg.Runtime.EnableDiagnostics = parseBoolEnv("JMS_ENABLE_DIAGNOSTICS", false)
	cfg.Runtime.EnableCleanup = parseBoolEnv("JMS_ENABLE_CLEANUP", false)

	// Intervals (اختيارية؛ صفر = معطّل)
	cfg.Runtime.HistoryInterval = parseDurationEnv("JMS_HISTORY_INTERVAL", 0)
	cfg.Runtime.EventsInterval = parseDurationEnv("JMS_EVENTS_INTERVAL", 0)
	cfg.Runtime.AlertsInterval = parseDurationEnv("JMS_ALERTS_INTERVAL", 0)
	cfg.Runtime.DiagnosticsInterval = parseDurationEnv("JMS_DIAGNOSTICS_INTERVAL", 0)
	cfg.Runtime.CleanupInterval = parseDurationEnv("JMS_CLEANUP_INTERVAL", 0)

	// HTTP shutdown timeout (اختياري؛ الافتراض 5 ثوانى آمنة منطقياً للتشغيل فقط)
	shutdownTimeout := parseDurationEnv("JMS_HTTP_SHUTDOWN_TIMEOUT", 5*time.Second)
	cfg.Runtime.HTTPShutdownTimeout = shutdownTimeout

	return cfg, nil
}

func parseBoolEnv(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func parseDurationEnv(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

// ToRuntimeConfig يحوّل AppConfig.Runtime إلى runtime.RuntimeConfig
// بدون أى تعديل على المعنى المنطقى.
func (c AppConfig) ToRuntimeConfig() runtime.RuntimeConfig {
	return runtime.RuntimeConfig{
		HistoryInterval:     c.Runtime.HistoryInterval,
		EventStoreInterval:  c.Runtime.EventsInterval,
		AlertsInterval:      c.Runtime.AlertsInterval,
		DiagnosticsInterval: c.Runtime.DiagnosticsInterval,
		CleanupInterval:     c.Runtime.CleanupInterval,

		EnableHistory:     c.Runtime.EnableHistory,
		EnableEvents:      c.Runtime.EnableEvents,
		EnableAlerts:      c.Runtime.EnableAlerts,
		EnableDiagnostics: c.Runtime.EnableDiagnostics,
		EnableCleanup:     c.Runtime.EnableCleanup,
		EnableWSHub:       c.Runtime.EnableWSHub,
		EnableHTTP:        c.Runtime.EnableHTTP,

		HTTPShutdownTimeout: c.Runtime.HTTPShutdownTimeout,
	}
}
