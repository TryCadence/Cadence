package analysis

import (
	"context"
	"fmt"
	"sync"

	"github.com/TryCadence/Cadence/internal/logging"
)

// StrategyPlugin defines the interface that custom detection strategy plugins must implement.
// Plugins are loaded programmatically and registered with the PluginManager.
type StrategyPlugin interface {
	// Info returns metadata about the strategy this plugin provides.
	Info() StrategyInfo

	// Detect runs the plugin's detection logic against the provided source data.
	// It returns a slice of detections (may be empty) or an error.
	Detect(ctx context.Context, data *SourceData) ([]Detection, error)
}

// PluginDetector wraps a set of loaded plugins as a Detector so they integrate
// seamlessly with the existing DetectionRunner / StreamingRunner pipeline.
type PluginDetector struct {
	manager *PluginManager
}

func (pd *PluginDetector) Detect(ctx context.Context, data *SourceData) ([]Detection, error) {
	return pd.manager.RunAll(ctx, data)
}

// PluginManager manages the lifecycle of strategy plugins.
type PluginManager struct {
	mu      sync.RWMutex
	plugins map[string]StrategyPlugin
	enabled map[string]bool // nil means "all enabled"
	logger  *logging.Logger
}

// NewPluginManager creates a new PluginManager.
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]StrategyPlugin),
		logger:  logging.Default().With("component", "plugin_manager"),
	}
}

// NewPluginManagerWithLogger creates a PluginManager with a custom logger.
func NewPluginManagerWithLogger(logger *logging.Logger) *PluginManager {
	if logger == nil {
		logger = logging.Default()
	}
	return &PluginManager{
		plugins: make(map[string]StrategyPlugin),
		logger:  logger.With("component", "plugin_manager"),
	}
}

// Register adds a plugin. If a plugin with the same strategy name already exists
// it is replaced (allowing hot-reload patterns).
func (pm *PluginManager) Register(p StrategyPlugin) error {
	if p == nil {
		return fmt.Errorf("cannot register nil plugin")
	}
	info := p.Info()
	if info.Name == "" {
		return fmt.Errorf("plugin must have a non-empty strategy name")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.plugins[info.Name] = p
	pm.logger.Info("plugin registered", "strategy", info.Name, "category", info.Category)
	return nil
}

// Unregister removes a plugin by strategy name.
func (pm *PluginManager) Unregister(name string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if _, ok := pm.plugins[name]; ok {
		delete(pm.plugins, name)
		pm.logger.Info("plugin unregistered", "strategy", name)
		return true
	}
	return false
}

// Get returns a specific plugin by name.
func (pm *PluginManager) Get(name string) (StrategyPlugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, ok := pm.plugins[name]
	return p, ok
}

// List returns info about all registered plugins.
func (pm *PluginManager) List() []StrategyInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	infos := make([]StrategyInfo, 0, len(pm.plugins))
	for _, p := range pm.plugins {
		infos = append(infos, p.Info())
	}
	return infos
}

// Count returns the number of registered plugins.
func (pm *PluginManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.plugins)
}

// SetEnabled configures which plugins are active. Pass nil to enable all.
// The map values indicate true=enabled, false=disabled.
func (pm *PluginManager) SetEnabled(enabled map[string]bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.enabled = enabled
}

// IsEnabled checks whether a specific plugin is currently enabled.
func (pm *PluginManager) IsEnabled(name string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if pm.enabled == nil {
		return true // nil means all enabled
	}
	v, ok := pm.enabled[name]
	return ok && v
}

// RunAll executes all enabled plugins against the source data and collects
// detections. A single plugin failure does not abort the remaining plugins.
func (pm *PluginManager) RunAll(ctx context.Context, data *SourceData) ([]Detection, error) {
	pm.mu.RLock()
	plugins := make([]StrategyPlugin, 0, len(pm.plugins))
	for name, p := range pm.plugins {
		if pm.isEnabledLocked(name) {
			plugins = append(plugins, p)
		}
	}
	pm.mu.RUnlock()

	var allDetections []Detection
	var errs []error

	for _, p := range plugins {
		select {
		case <-ctx.Done():
			return allDetections, ctx.Err()
		default:
		}

		info := p.Info()
		detections, err := pm.safeDetect(ctx, p, data)
		if err != nil {
			pm.logger.Error("plugin detection failed",
				"strategy", info.Name,
				"error", err.Error(),
			)
			errs = append(errs, fmt.Errorf("plugin %s: %w", info.Name, err))
			continue
		}
		allDetections = append(allDetections, detections...)
	}

	if len(errs) > 0 && len(allDetections) == 0 {
		return nil, fmt.Errorf("all plugins failed: %v", errs)
	}

	return allDetections, nil
}

// safeDetect runs a plugin with panic recovery.
func (pm *PluginManager) safeDetect(ctx context.Context, p StrategyPlugin, data *SourceData) (detections []Detection, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("plugin panicked: %v", r)
		}
	}()
	return p.Detect(ctx, data)
}

// isEnabledLocked checks if a plugin is enabled (caller must hold at least RLock).
func (pm *PluginManager) isEnabledLocked(name string) bool {
	if pm.enabled == nil {
		return true
	}
	v, ok := pm.enabled[name]
	return ok && v
}

// Detector returns a Detector wrapping all registered plugins for use with
// DetectionRunner or StreamingRunner.
func (pm *PluginManager) Detector() Detector {
	return &PluginDetector{manager: pm}
}

// MergeIntoRegistry copies all plugin strategy metadata into a StrategyRegistry.
func (pm *PluginManager) MergeIntoRegistry(registry *StrategyRegistry) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	for _, p := range pm.plugins {
		registry.Register(p.Info())
	}
}
