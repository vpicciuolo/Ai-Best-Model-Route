package autoroute

import (
	"encoding/json"
	"testing"

	"github.com/vpicciuolo/ai-best-model-route/internal/config"
)

func testConfig(t *testing.T) config.Config {
	t.Helper()
	cfg := config.Config{
		Version: 1,
		Providers: map[string]config.ProviderProfile{
			"p": {
				CredentialMode: config.CredentialBifrost,
				ResponsesMode:  config.ResponsesNative,
				Adapter:        "native",
			},
		},
		Models: map[string]config.ModelProfile{
			"p/strong": {
				Provider: "p",
				UpstreamModel: "strong",
				ResponsesMode: config.ResponsesNative,
				Adapter: "native",
				Codex: config.CodexProfile{
					ContextWindow: 200000,
					MaxContextWindow: 200000,
					EffectiveContextWindowPercent: 95,
					InputModalities: []string{"text", "image"},
					SupportsSearch: true,
				},
				Route: config.RouteModelProfile{
					Quality: .96,
					InputCostPerM: 8,
					OutputCostPerM: 20,
					LatencyMS: 1600,
					Reliability: .995,
					Strengths: []string{"reasoning", "code", "tools", "vision"},
					PrivacyTier: 2,
				},
			},
			"p/fast": {
				Provider: "p",
				UpstreamModel: "fast",
				ResponsesMode: config.ResponsesNative,
				Adapter: "native",
				Codex: config.CodexProfile{
					ContextWindow: 128000,
					MaxContextWindow: 128000,
					EffectiveContextWindowPercent: 95,
					InputModalities: []string{"text"},
					SupportsSearch: true,
				},
				Route: config.RouteModelProfile{
					Quality: .72,
					InputCostPerM: .2,
					OutputCostPerM: .6,
					LatencyMS: 180,
					Reliability: .995,
					Strengths: []string{"fast", "general"},
					PrivacyTier: 1,
				},
			},
			"p/local": {
				Provider: "p",
				UpstreamModel: "local",
				ResponsesMode: config.ResponsesNative,
				Adapter: "native",
				Codex: config.CodexProfile{
					ContextWindow: 128000,
					MaxContextWindow: 128000,
					EffectiveContextWindowPercent: 95,
					InputModalities: []string{"text"},
				},
				Route: config.RouteModelProfile{
					Quality: .80,
					LatencyMS: 400,
					Reliability: .99,
					Strengths: []string{"code", "general"},
					PrivacyTier: 3,
					Local: true,
				},
			},
		},
		AutoRoute: config.AutoRouteConfig{Enabled: true},
	}
	if err := cfg.ApplyDefaultsAndValidate(); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func selected(t *testing.T, engine *Engine, body string) string {
	t.Helper()
	routed, decision, err := engine.Rewrite([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if decision == nil {
		t.Fatal("expected auto-routing decision")
	}
	var envelope struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(routed, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Model != decision.SelectedModel {
		t.Fatalf("rewritten model %q differs from decision %q", envelope.Model, decision.SelectedModel)
	}
	return envelope.Model
}

func TestFastProfilePrefersFastModel(t *testing.T) {
	engine := New(testConfig(t))
	got := selected(t, engine, `{"model":"auto:fast","input":"hello, summarize this sentence"}`)
	if got != "p/fast" {
		t.Fatalf("got %s", got)
	}
}

func TestQualityProfilePrefersStrongModel(t *testing.T) {
	engine := New(testConfig(t))
	got := selected(t, engine, `{"model":"auto:quality","input":"deeply analyze this architecture and reason about failure modes"}`)
	if got != "p/strong" {
		t.Fatalf("got %s", got)
	}
}

func TestVisionHardFilter(t *testing.T) {
	engine := New(testConfig(t))
	got := selected(t, engine, `{"model":"auto","input":[{"role":"user","content":[{"type":"input_image","image_url":"data:image/png;base64,AA"}]}]}`)
	if got != "p/strong" {
		t.Fatalf("vision request routed to %s", got)
	}
}

func TestLocalConstraint(t *testing.T) {
	engine := New(testConfig(t))
	got := selected(t, engine, `{"model":"auto","input":"review this code","ai_route":{"require_local":true}}`)
	if got != "p/local" {
		t.Fatalf("local request routed to %s", got)
	}
}

func TestAIRouteRemovedBeforeForward(t *testing.T) {
	engine := New(testConfig(t))
	routed, _, err := engine.Rewrite([]byte(`{"model":"auto","input":"hello","ai_route":{"profile":"cheap"}}`))
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(routed, &object); err != nil {
		t.Fatal(err)
	}
	if _, ok := object["ai_route"]; ok {
		t.Fatal("ai_route leaked upstream")
	}
}

func TestCircuitBreakerRemovesRepeatedFailure(t *testing.T) {
	engine := New(testConfig(t))
	body := `{"model":"auto:quality","input":"deeply analyze this architecture and reason about failure modes"}`
	first := selected(t, engine, body)
	if first != "p/strong" {
		t.Fatalf("expected strong first, got %s", first)
	}
	engine.Observe("p/strong", 500)
	engine.Observe("p/strong", 500)
	engine.Observe("p/strong", 500)
	second := selected(t, engine, body)
	if second == "p/strong" {
		t.Fatal("circuit-open model should not remain selected")
	}
}

func TestHostedSearchRequiresNativeSearchModel(t *testing.T) {
	cfg := testConfig(t)
	model := cfg.Models["p/strong"]
	model.ResponsesMode = config.ResponsesChatPolyfill
	model.Codex.SupportsSearch = false
	cfg.Models["p/strong"] = model
	if err := cfg.ApplyDefaultsAndValidate(); err == nil {
		t.Fatal("expected invalid polyfill search capability")
	}
}

func TestInjectVirtualModels(t *testing.T) {
	cfg := testConfig(t)
	out, err := InjectVirtualModels([]byte(`{"models":[]}`), cfg)
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Models []map[string]any `json:"models"`
	}
	if err := json.Unmarshal(out, &envelope); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, model := range envelope.Models {
		if model["id"] == "auto:code" {
			found = true
		}
	}
	if !found {
		t.Fatal("auto:code missing from catalog")
	}
}

func TestPreviewDoesNotExecuteAndExplainsCandidates(t *testing.T) {
	engine := New(testConfig(t))
	decision, err := engine.Preview([]byte(`{"model":"auto:code","input":"refactor this Go API and fix its race condition"}`))
	if err != nil {
		t.Fatal(err)
	}
	if decision.SelectedModel == "" || len(decision.Candidates) != 3 {
		t.Fatalf("unexpected decision: %#v", decision)
	}
	if decision.Profile != "code" {
		t.Fatalf("expected code profile, got %q", decision.Profile)
	}
}
