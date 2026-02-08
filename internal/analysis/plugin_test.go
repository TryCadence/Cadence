package analysis

import (
	"context"
	"errors"
	"testing"
)

// mockPlugin is a test StrategyPlugin implementation.
type mockPlugin struct {
	info       StrategyInfo
	detections []Detection
	err        error
	panicMsg   string
}

func (m *mockPlugin) Info() StrategyInfo { return m.info }

func (m *mockPlugin) Detect(_ context.Context, _ *SourceData) ([]Detection, error) {
	if m.panicMsg != "" {
		panic(m.panicMsg)
	}
	return m.detections, m.err
}

func TestPluginManager_RegisterAndGet(t *testing.T) {
	pm := NewPluginManager()

	p := &mockPlugin{
		info: StrategyInfo{Name: "test_strategy", Category: CategoryPattern, Confidence: 0.8},
	}

	if err := pm.Register(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := pm.Get("test_strategy")
	if !ok {
		t.Fatal("expected plugin to be found")
	}
	if got.Info().Name != "test_strategy" {
		t.Fatalf("got name %q, want %q", got.Info().Name, "test_strategy")
	}

	if pm.Count() != 1 {
		t.Fatalf("got count %d, want 1", pm.Count())
	}
}

func TestPluginManager_RegisterNil(t *testing.T) {
	pm := NewPluginManager()
	if err := pm.Register(nil); err == nil {
		t.Fatal("expected error when registering nil plugin")
	}
}

func TestPluginManager_RegisterEmptyName(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{info: StrategyInfo{Name: ""}}
	if err := pm.Register(p); err == nil {
		t.Fatal("expected error when registering plugin with empty name")
	}
}

func TestPluginManager_Unregister(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{info: StrategyInfo{Name: "remove_me"}}
	_ = pm.Register(p)

	if !pm.Unregister("remove_me") {
		t.Fatal("expected Unregister to return true")
	}
	if pm.Unregister("remove_me") {
		t.Fatal("expected Unregister to return false for missing plugin")
	}
	if pm.Count() != 0 {
		t.Fatalf("got count %d, want 0", pm.Count())
	}
}

func TestPluginManager_List(t *testing.T) {
	pm := NewPluginManager()
	_ = pm.Register(&mockPlugin{info: StrategyInfo{Name: "a"}})
	_ = pm.Register(&mockPlugin{info: StrategyInfo{Name: "b"}})

	infos := pm.List()
	if len(infos) != 2 {
		t.Fatalf("got %d plugins, want 2", len(infos))
	}
}

func TestPluginManager_EnableDisable(t *testing.T) {
	pm := NewPluginManager()
	_ = pm.Register(&mockPlugin{info: StrategyInfo{Name: "always_on"}})
	_ = pm.Register(&mockPlugin{info: StrategyInfo{Name: "toggled"}})

	// All enabled by default
	if !pm.IsEnabled("always_on") || !pm.IsEnabled("toggled") {
		t.Fatal("all plugins should be enabled by default")
	}

	// Disable one
	pm.SetEnabled(map[string]bool{
		"always_on": true,
		"toggled":   false,
	})

	if !pm.IsEnabled("always_on") {
		t.Fatal("always_on should still be enabled")
	}
	if pm.IsEnabled("toggled") {
		t.Fatal("toggled should be disabled")
	}

	// Re-enable all
	pm.SetEnabled(nil)
	if !pm.IsEnabled("toggled") {
		t.Fatal("all plugins should be enabled after SetEnabled(nil)")
	}
}

func TestPluginManager_RunAll(t *testing.T) {
	pm := NewPluginManager()
	_ = pm.Register(&mockPlugin{
		info:       StrategyInfo{Name: "p1"},
		detections: []Detection{{Strategy: "p1", Detected: true, Score: 0.9}},
	})
	_ = pm.Register(&mockPlugin{
		info:       StrategyInfo{Name: "p2"},
		detections: []Detection{{Strategy: "p2", Detected: false, Score: 0.1}},
	})

	detections, err := pm.RunAll(context.Background(), &SourceData{Type: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(detections) != 2 {
		t.Fatalf("got %d detections, want 2", len(detections))
	}
}

func TestPluginManager_RunAll_Disabled(t *testing.T) {
	pm := NewPluginManager()
	_ = pm.Register(&mockPlugin{
		info:       StrategyInfo{Name: "enabled"},
		detections: []Detection{{Strategy: "enabled", Detected: true}},
	})
	_ = pm.Register(&mockPlugin{
		info:       StrategyInfo{Name: "disabled"},
		detections: []Detection{{Strategy: "disabled", Detected: true}},
	})

	pm.SetEnabled(map[string]bool{"enabled": true})

	detections, err := pm.RunAll(context.Background(), &SourceData{Type: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(detections) != 1 {
		t.Fatalf("got %d detections, want 1", len(detections))
	}
	if detections[0].Strategy != "enabled" {
		t.Fatalf("got strategy %q, want %q", detections[0].Strategy, "enabled")
	}
}

func TestPluginManager_RunAll_PanicRecovery(t *testing.T) {
	pm := NewPluginManager()
	_ = pm.Register(&mockPlugin{
		info:     StrategyInfo{Name: "panicker"},
		panicMsg: "boom!",
	})
	_ = pm.Register(&mockPlugin{
		info:       StrategyInfo{Name: "safe"},
		detections: []Detection{{Strategy: "safe", Detected: true}},
	})

	detections, err := pm.RunAll(context.Background(), &SourceData{Type: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still get safe plugin's detection
	if len(detections) != 1 {
		t.Fatalf("got %d detections, want 1", len(detections))
	}
}

func TestPluginManager_RunAll_AllFail(t *testing.T) {
	pm := NewPluginManager()
	_ = pm.Register(&mockPlugin{
		info: StrategyInfo{Name: "fail1"},
		err:  errors.New("fail1"),
	})
	_ = pm.Register(&mockPlugin{
		info: StrategyInfo{Name: "fail2"},
		err:  errors.New("fail2"),
	})

	_, err := pm.RunAll(context.Background(), &SourceData{Type: "test"})
	if err == nil {
		t.Fatal("expected error when all plugins fail")
	}
}

func TestPluginManager_RunAll_ContextCancelled(t *testing.T) {
	pm := NewPluginManager()
	_ = pm.Register(&mockPlugin{
		info:       StrategyInfo{Name: "slow"},
		detections: []Detection{{Strategy: "slow"}},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := pm.RunAll(ctx, &SourceData{Type: "test"})
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestPluginDetector(t *testing.T) {
	pm := NewPluginManager()
	_ = pm.Register(&mockPlugin{
		info:       StrategyInfo{Name: "d1"},
		detections: []Detection{{Strategy: "d1", Detected: true}},
	})

	det := pm.Detector()
	detections, err := det.Detect(context.Background(), &SourceData{Type: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(detections) != 1 {
		t.Fatalf("got %d detections, want 1", len(detections))
	}
}

func TestPluginManager_MergeIntoRegistry(t *testing.T) {
	pm := NewPluginManager()
	_ = pm.Register(&mockPlugin{
		info: StrategyInfo{Name: "plugin_strat", Category: CategoryPattern, Confidence: 0.9},
	})

	reg := NewStrategyRegistry()
	pm.MergeIntoRegistry(reg)

	info, ok := reg.Get("plugin_strat")
	if !ok {
		t.Fatal("expected strategy to be in registry")
	}
	if info.Confidence != 0.9 {
		t.Fatalf("got confidence %f, want 0.9", info.Confidence)
	}
}

func TestPluginManager_ReplacePlugin(t *testing.T) {
	pm := NewPluginManager()

	_ = pm.Register(&mockPlugin{
		info:       StrategyInfo{Name: "replaceable"},
		detections: []Detection{{Strategy: "v1"}},
	})

	_ = pm.Register(&mockPlugin{
		info:       StrategyInfo{Name: "replaceable"},
		detections: []Detection{{Strategy: "v2"}},
	})

	if pm.Count() != 1 {
		t.Fatalf("got count %d, want 1 (replace, not add)", pm.Count())
	}

	detections, _ := pm.RunAll(context.Background(), &SourceData{Type: "test"})
	if len(detections) != 1 || detections[0].Strategy != "v2" {
		t.Fatal("expected replaced plugin to produce v2 detections")
	}
}
