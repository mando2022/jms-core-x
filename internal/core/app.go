package core

import (
	"context"
	"fmt"
	"sync"
)

// App يمثل الـ Core Runtime لـ JMS Core X
// مسؤول عن تسجيل الطبقات، وادارة lifecycle الخاص بها
type App struct {
	mu     sync.RWMutex
	layers map[LayerName]Layer
	order  []LayerName
}

// NewApp ينشئ App جديد مع ترتيب افتراضي للطبقات
// حسب Block 3: Auth → Settings → Discovery → Unified → Ping → Map → Events → Backup
func NewApp() *App {
	return &App{
		layers: make(map[LayerName]Layer),
		order: []LayerName{
			LayerAuth,
			LayerSettings,
			LayerDiscovery,
			LayerUnified,
			LayerPing,
			LayerMap,
			LayerEvents,
			LayerBackup,
		},
	}
}

// RegisterLayer يضيف Layer جديد للـ App
// يمنع التكرار بنفس الاسم
func (a *App) RegisterLayer(layer Layer) error {
	if layer == nil {
		return fmt.Errorf("cannot register nil layer")
	}

	name := layer.Name()

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.layers[name]; exists {
		return fmt.Errorf("layer %q already registered", name)
	}

	a.layers[name] = layer
	return nil
}

// GetLayer يرجع Layer حسب الاسم (لو موجود)
func (a *App) GetLayer(name LayerName) (Layer, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	l, ok := a.layers[name]
	return l, ok
}

// InitAll يقوم بتشغيل Init لكل Layer مسجل
// بالترتيب المحدد في Block 3
func (a *App) InitAll(ctx context.Context, configs map[LayerName]Config) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, name := range a.order {
		layer, ok := a.layers[name]
		if !ok {
			// layer مش لازم يكون موجود من دلوقتي – بيتكمل في البلوكات التالية
			continue
		}

		cfg := Config(nil)
		if configs != nil {
			if c, exists := configs[name]; exists {
				cfg = c
			}
		}

		// لو الـ Layer بيستخدم BaseLayer نقدر نمرر الـ Config من بره
		if bl, ok := layer.(interface {
			setConfig(Config)
		}); ok {
			bl.setConfig(cfg)
		}

		if err := layer.Init(ctx, cfg); err != nil {
			return fmt.Errorf("init layer %q: %w", name, err)
		}
	}

	return nil
}

// StartAll يقوم بتشغيل كل الطبقات بالترتيب
func (a *App) StartAll(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, name := range a.order {
		layer, ok := a.layers[name]
		if !ok {
			continue
		}

		if err := layer.Start(ctx); err != nil {
			return fmt.Errorf("start layer %q: %w", name, err)
		}
	}

	return nil
}

// StopAll يوقف كل الطبقات بالعكس (من الآخر للأول)
// عشان نحترم التبعيات (مثلاً: نوقف Ping قبل Unified)
func (a *App) StopAll(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for i := len(a.order) - 1; i >= 0; i-- {
		name := a.order[i]
		layer, ok := a.layers[name]
		if !ok {
			continue
		}

		if err := layer.Stop(ctx); err != nil {
			return fmt.Errorf("stop layer %q: %w", name, err)
		}
	}

	return nil
}

// StatusSnapshot يرجع حالة كل Layer مسجل في لحظة واحدة
func (a *App) StatusSnapshot() map[LayerName]LayerStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	out := make(map[LayerName]LayerStatus, len(a.layers))
	for name, layer := range a.layers {
		out[name] = layer.Status()
	}
	return out
}
