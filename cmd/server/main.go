package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jms-core-x/internal/core"
	"jms-core-x/internal/runtime"

	// إضافات Auto-Ping V2 (Option A)
	"jms-core-x/internal/core/diagnostics"
	"jms-core-x/internal/bootstrap"
)

// ======================================================
//   main.go — JMS Core X
//   Boot → Core (Block 3) + Runtime (Block 19)
//   (Option A) تفعيل Auto-Ping فقط داخل الـ Orchestrator
// ======================================================

func main() {

	//-----------------------------------------------
	// 1) Context + Logger
	//-----------------------------------------------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.New(os.Stdout, "[jms-core-x] ", log.LstdFlags|log.Lmicroseconds)
	logger.Println("Bootstrapping JMS Core X...")

	//-----------------------------------------------
	// 2) Core App (Block 3)
	//-----------------------------------------------
	app := core.NewApp()

	if err := app.InitAll(ctx, nil); err != nil {
		logger.Fatalf("init layers failed: %v", err)
	}

	if err := app.StartAll(ctx); err != nil {
		logger.Fatalf("start layers failed: %v", err)
	}

	logger.Println("Core App (Block 3) started successfully")

	//-----------------------------------------------
	// 3) RuntimeConfig (Block 19) مع Auto-Ping فقط
	//-----------------------------------------------
	runtimeCfg := runtime.RuntimeConfig{
		EnableHistory: false,
		EnableEvents:  false,
		EnableAlerts:  false,
		EnableWSHub:   false,
		EnableHTTP:    false,

		// -------------------------
		// Auto-Ping V2 (Option A)
		// -------------------------
		EnableDiagnostics:   true,
		DiagnosticsInterval: 5 * time.Second,

		EnableCleanup:       false,
		CleanupInterval:     0,

		HistoryInterval:     0,
		EventStoreInterval:  0,
		AlertsInterval:      0,

		HTTPShutdownTimeout: 5 * time.Second,
	}

	//-----------------------------------------------
	// 4) بناء DiagnosticsLayer + Job (Auto-Ping V2)
	//-----------------------------------------------
	diagLayer := diagnostics.NewDiagnosticsLayer(
		nil,                    // UnifiedSnapshotProvider — غير مفعل الآن في A
		diagnostics.ICMPProbe,  // Ping Function
		nil,                    // EventsReader — غير مفعل الآن في A
	)

	// إنشاء Job رسمي للـ Orchestrator
	diagnosticsJob := bootstrap.NewDiagnosticsJob(diagLayer)

	//-----------------------------------------------
	// 5) OrchestratorDeps
	//-----------------------------------------------
	var deps runtime.OrchestratorDeps
	deps.DiagnosticsJob = diagnosticsJob

	logger.Println("Diagnostics Layer + Auto-Ping Job initialized (Option A)")

	//-----------------------------------------------
	// 6) إنشاء Orchestrator (Block 19)
	//-----------------------------------------------
	orch := runtime.NewOrchestrator(ctx, runtimeCfg, logger, deps)

	//-----------------------------------------------
	// 7) تشغيل الـ Runtime
	//-----------------------------------------------
	orch.Start()
	logger.Println("Runtime Orchestrator (Block 19) started successfully")

	logger.Println("JMS Core X is running... Press CTRL+C to exit.")

	//-----------------------------------------------
	// 8) انتظار إشارة الإيقاف
	//-----------------------------------------------
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Println("Shutdown signal received, stopping services...")

	//-----------------------------------------------
	// 9) إيقاف Runtime
	//-----------------------------------------------
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := orch.Shutdown(shutdownCtx); err != nil {
		logger.Printf("orchestrator shutdown error: %v", err)
	}

	//-----------------------------------------------
	// 10) إيقاف Core Layers
	//-----------------------------------------------
	if err := app.StopAll(ctx); err != nil {
		logger.Printf("stop layers error: %v", err)
	}

	logger.Println("JMS Core X shut down cleanly.")
}
