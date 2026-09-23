package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	CurrentVersion                 = 1
	DefaultHostedToolFallbackModel = "openai/gpt-6-luna"
)

type CredentialMode string

const (
	CredentialBifrost            CredentialMode = "bifrost"
	CredentialRequestPassthrough CredentialMode = "request_passthrough"
)

type ResponsesMode string

const (
	ResponsesNative       ResponsesMode = "native"
	ResponsesChatPolyfill ResponsesMode = "chat_polyfill"
	ResponsesUnsupported  ResponsesMode = "unsupported"
)

type Config struct {
	Version                 int                        `json:"version" yaml:"version"`
	Instructions            string                     `json:"instructions_template,omitempty" yaml:"instructions_template,omitempty"`
	HostedToolFallbackModel string                     `json:"hosted_tool_fallback_model,omitempty" yaml:"hosted_tool_fallback_model,omitempty"`
	Providers               map[string]ProviderProfile `json:"providers" yaml:"providers"`
	Models                  map[string]ModelProfile    `json:"models" yaml:"models"`
	AutoRoute               AutoRouteConfig             `json:"auto_route,omitempty" yaml:"auto_route,omitempty"`
	resolutionIndex         map[string]ResolvedModel
	resolvedModels          []ResolvedModel
}

type ProviderProfile struct {
	DisplayName        string            `json:"display_name,omitempty" yaml:"display_name,omitempty"`
	CredentialMode     CredentialMode    `json:"credential_mode" yaml:"credential_mode"`
	ResponsesMode      ResponsesMode     `json:"responses_mode" yaml:"responses_mode"`
	Adapter            string            `json:"adapter,omitempty" yaml:"adapter,omitempty"`
	DiscoverModels     bool              `json:"discover_models,omitempty" yaml:"discover_models,omitempty"`
	ModelNameOverrides map[string]string `json:"model_name_overrides,omitempty" yaml:"model_name_overrides,omitempty"`
	CodexDefaults      CodexProfile      `json:"codex_defaults,omitempty" yaml:"codex_defaults,omitempty"`
}

func (c Config) ProviderDisplayName(name string) string {
	if displayName := strings.TrimSpace(c.Providers[name].DisplayName); displayName != "" {
		return displayName
	}
	parts := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, part := range parts {
		switch strings.ToLower(part) {
		case "openai":
			parts[i] = "OpenAI"
		case "openrouter":
			parts[i] = "OpenRouter"
		case "nvidia":
			parts[i] = "NVIDIA"
		default:
			runes := []rune(part)
			runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
			parts[i] = string(runes)
		}
	}
	return strings.Join(parts, " ")
}

type ReasoningLevel struct {
	Effort      string `json:"effort" yaml:"effort"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type CodexProfile struct {
	DisplayName                   string           `json:"display_name,omitempty" yaml:"display_name,omitempty"`
	Description                   string           `json:"description,omitempty" yaml:"description,omitempty"`
	ContextWindow                 int64            `json:"context_window,omitempty" yaml:"context_window,omitempty"`
	MaxContextWindow              int64            `json:"max_context_window,omitempty" yaml:"max_context_window,omitempty"`
	EffectiveContextWindowPercent int64            `json:"effective_context_window_percent,omitempty" yaml:"effective_context_window_percent,omitempty"`
	DefaultReasoningLevel         string           `json:"default_reasoning_level,omitempty" yaml:"default_reasoning_level,omitempty"`
	SupportedReasoningLevels      []ReasoningLevel `json:"supported_reasoning_levels,omitempty" yaml:"supported_reasoning_levels,omitempty"`
	InputModalities               []string         `json:"input_modalities,omitempty" yaml:"input_modalities,omitempty"`
	SupportsReasoningSummaries    bool             `json:"supports_reasoning_summaries,omitempty" yaml:"supports_reasoning_summaries,omitempty"`
	SupportsVerbosity             bool             `json:"supports_verbosity,omitempty" yaml:"supports_verbosity,omitempty"`
	SupportsSearch                bool             `json:"supports_search,omitempty" yaml:"supports_search,omitempty"`
	SupportsImageDetailOriginal   bool             `json:"supports_image_detail_original,omitempty" yaml:"supports_image_detail_original,omitempty"`
	TruncationLimit               int64            `json:"truncation_limit,omitempty" yaml:"truncation_limit,omitempty"`
}

type ContextVariant struct {
	ContextWindow                 int64 `json:"context_window" yaml:"context_window"`
	EffectiveContextWindowPercent int64 `json:"effective_context_window_percent,omitempty" yaml:"effective_context_window_percent,omitempty"`
}

type ModelProfile struct {
	Aliases         []string          `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	Provider        string            `json:"provider,omitempty" yaml:"provider,omitempty"`
	UpstreamModel   string            `json:"upstream_model,omitempty" yaml:"upstream_model,omitempty"`
	ResponsesMode   ResponsesMode     `json:"responses_mode,omitempty" yaml:"responses_mode,omitempty"`
	Adapter         string            `json:"adapter,omitempty" yaml:"adapter,omitempty"`
	Codex           CodexProfile      `json:"codex" yaml:"codex"`
	ContextVariants []ContextVariant  `json:"context_variants,omitempty" yaml:"context_variants,omitempty"`
	Route           RouteModelProfile `json:"route,omitempty" yaml:"route,omitempty"`
}

type AutoRouteConfig struct {
	Enabled        bool                    `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	DefaultProfile string                  `json:"default_profile,omitempty" yaml:"default_profile,omitempty"`
	Aliases        map[string]string       `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	Profiles       map[string]RouteProfile `json:"profiles,omitempty" yaml:"profiles,omitempty"`
}

type RouteProfile struct {
	QualityWeight     float64 `json:"quality_weight,omitempty" yaml:"quality_weight,omitempty"`
	CostWeight        float64 `json:"cost_weight,omitempty" yaml:"cost_weight,omitempty"`
	LatencyWeight     float64 `json:"latency_weight,omitempty" yaml:"latency_weight,omitempty"`
	ReliabilityWeight float64 `json:"reliability_weight,omitempty" yaml:"reliability_weight,omitempty"`
	AffinityWeight    float64 `json:"affinity_weight,omitempty" yaml:"affinity_weight,omitempty"`
	PrivacyWeight     float64 `json:"privacy_weight,omitempty" yaml:"privacy_weight,omitempty"`
	MinQuality        float64 `json:"min_quality,omitempty" yaml:"min_quality,omitempty"`
	MaxInputCostPerM  float64 `json:"max_input_cost_per_m,omitempty" yaml:"max_input_cost_per_m,omitempty"`
	MaxOutputCostPerM float64 `json:"max_output_cost_per_m,omitempty" yaml:"max_output_cost_per_m,omitempty"`
	PreferLocal       bool    `json:"prefer_local,omitempty" yaml:"prefer_local,omitempty"`
}

type RouteModelProfile struct {
	Disabled       bool     `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Quality        float64  `json:"quality,omitempty" yaml:"quality,omitempty"`
	InputCostPerM  float64  `json:"input_cost_per_m,omitempty" yaml:"input_cost_per_m,omitempty"`
	OutputCostPerM float64  `json:"output_cost_per_m,omitempty" yaml:"output_cost_per_m,omitempty"`
	LatencyMS      float64  `json:"latency_ms,omitempty" yaml:"latency_ms,omitempty"`
	Reliability    float64  `json:"reliability,omitempty" yaml:"reliability,omitempty"`
	Strengths      []string `json:"strengths,omitempty" yaml:"strengths,omitempty"`
	PrivacyTier    int      `json:"privacy_tier,omitempty" yaml:"privacy_tier,omitempty"`
	Local          bool     `json:"local,omitempty" yaml:"local,omitempty"`
}

func Decode(r io.Reader) (Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := cfg.ApplyDefaultsAndValidate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func FromAny(raw any) (Config, error) {
	if raw == nil {
		return Config{}, errors.New("plugin config is required")
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return Config{}, fmt.Errorf("encode plugin config: %w", err)
	}
	var cfg Config
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode plugin config: %w", err)
	}
	if err := cfg.ApplyDefaultsAndValidate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) ApplyDefaultsAndValidate() error {
	if c.Version == 0 {
		c.Version = CurrentVersion
	}
	if c.Version != CurrentVersion {
		return fmt.Errorf("unsupported config version %d", c.Version)
	}
	applyAutoRouteDefaults(&c.AutoRoute)
	if len(c.Providers) == 0 {
		return errors.New("at least one provider is required")
	}
	if len(c.Models) == 0 {
		return errors.New("at least one model is required")
	}

	for name, provider := range c.Providers {
		if strings.TrimSpace(name) == "" {
			return errors.New("provider name cannot be empty")
		}
		if provider.CredentialMode == "" {
			provider.CredentialMode = CredentialBifrost
		}
		if provider.ResponsesMode == "" {
			provider.ResponsesMode = ResponsesNative
		}
		if provider.Adapter == "" {
			if provider.ResponsesMode == ResponsesChatPolyfill {
				provider.Adapter = "openai-chat"
			} else {
				provider.Adapter = "native"
			}
		}
		if err := validateProvider(name, provider); err != nil {
			return err
		}
		if provider.DiscoverModels {
			defaults := ModelProfile{Provider: name, ResponsesMode: provider.ResponsesMode, Adapter: provider.Adapter, UpstreamModel: "discovered", Codex: provider.CodexDefaults}
			applyCodexDefaults(name+"/discovered", &defaults.Codex)
			if err := validateModel(name+"/discovered", defaults); err != nil {
				return fmt.Errorf("provider %q codex_defaults: %w", name, err)
			}
		}
		c.Providers[name] = provider
	}

	owned := make(map[string]string)
	for slug, model := range c.Models {
		if strings.TrimSpace(slug) == "" || !strings.Contains(slug, "/") {
			return fmt.Errorf("model %q must use provider/model form", slug)
		}
		providerName := model.Provider
		if providerName == "" {
			providerName = strings.SplitN(slug, "/", 2)[0]
			model.Provider = providerName
		}
		provider, ok := c.Providers[providerName]
		if !ok {
			return fmt.Errorf("model %q references unknown provider %q", slug, providerName)
		}
		if model.ResponsesMode == "" {
			model.ResponsesMode = provider.ResponsesMode
		}
		if model.Adapter == "" {
			model.Adapter = provider.Adapter
		}
		if model.UpstreamModel == "" {
			model.UpstreamModel = strings.SplitN(slug, "/", 2)[1]
		}
		applyCodexDefaults(slug, &model.Codex)
		if err := validateModel(slug, model); err != nil {
			return err
		}
		for _, name := range modelNames(slug, model) {
			if previous, exists := owned[name]; exists && previous != slug {
				return fmt.Errorf("model name or alias %q is owned by both %q and %q", name, previous, slug)
			}
			owned[name] = slug
		}
		for i := range model.ContextVariants {
			variant := &model.ContextVariants[i]
			if variant.EffectiveContextWindowPercent == 0 {
				variant.EffectiveContextWindowPercent = model.Codex.EffectiveContextWindowPercent
			}
			if err := validateVariant(slug, model, *variant); err != nil {
				return err
			}
			for _, name := range variantNames(slug, model, variant.ContextWindow) {
				if previous, exists := owned[name]; exists {
					return fmt.Errorf("model name or alias %q collides with %q", name, previous)
				}
				owned[name] = slug + "#variant"
			}
		}
		c.Models[slug] = model
	}
	c.buildResolutionIndex()
	if c.HostedToolFallbackModel == "" {
		if fallback, ok := c.ResolveModel(DefaultHostedToolFallbackModel); ok &&
			fallback.Slug == DefaultHostedToolFallbackModel &&
			fallback.Provider.CredentialMode == CredentialRequestPassthrough &&
			fallback.Model.ResponsesMode == ResponsesNative {
			c.HostedToolFallbackModel = DefaultHostedToolFallbackModel
		}
	}

	if c.HostedToolFallbackModel != "" {
		fallback, ok := c.ResolveModel(c.HostedToolFallbackModel)
		if !ok {
			return fmt.Errorf("hosted_tool_fallback_model %q is not present in the router catalog", c.HostedToolFallbackModel)
		}
		if fallback.Provider.CredentialMode != CredentialRequestPassthrough || fallback.Model.ResponsesMode != ResponsesNative {
			return fmt.Errorf("hosted_tool_fallback_model %q must use native Responses with request_passthrough credentials", c.HostedToolFallbackModel)
		}
		c.HostedToolFallbackModel = fallback.Slug
	}
	if err := validateAutoRoute(*c); err != nil {
		return err
	}
	return nil
}

func validateProvider(name string, p ProviderProfile) error {
	if p.DisplayName != "" && strings.TrimSpace(p.DisplayName) == "" {
		return fmt.Errorf("provider %q display_name cannot be blank", name)
	}
	switch p.CredentialMode {
	case CredentialBifrost:
	case CredentialRequestPassthrough:
		if name != "openai" {
			return fmt.Errorf("provider %q cannot use request_passthrough; only openai is allowed", name)
		}
	default:
		return fmt.Errorf("provider %q has invalid credential_mode %q", name, p.CredentialMode)
	}
	if !validResponsesMode(p.ResponsesMode) {
		return fmt.Errorf("provider %q has invalid responses_mode %q", name, p.ResponsesMode)
	}
	if !validAdapter(p.Adapter) {
		return fmt.Errorf("provider %q has unknown adapter %q", name, p.Adapter)
	}
	for model, displayName := range p.ModelNameOverrides {
		if strings.TrimSpace(model) == "" || strings.TrimSpace(displayName) == "" {
			return fmt.Errorf("provider %q model_name_overrides must use non-empty model IDs and names", name)
		}
	}
	return nil
}

func validateModel(slug string, m ModelProfile) error {
	if strings.TrimSpace(m.UpstreamModel) == "" {
		return fmt.Errorf("model %q upstream_model cannot be empty", slug)
	}
	if !validResponsesMode(m.ResponsesMode) {
		return fmt.Errorf("model %q has invalid responses_mode %q", slug, m.ResponsesMode)
	}
	if !validAdapter(m.Adapter) {
		return fmt.Errorf("model %q has unknown adapter %q", slug, m.Adapter)
	}
	if m.ResponsesMode == ResponsesChatPolyfill {
		if m.Codex.SupportsSearch || m.Codex.SupportsImageDetailOriginal {
			return fmt.Errorf("model %q advertises native-only capabilities while using chat_polyfill", slug)
		}
		for _, modality := range m.Codex.InputModalities {
			if modality != "text" {
				return fmt.Errorf("model %q advertises modality %q while using the text-only chat_polyfill", slug, modality)
			}
		}
	}
	if m.Codex.ContextWindow <= 0 {
		return fmt.Errorf("model %q context_window must be positive", slug)
	}
	if m.Codex.MaxContextWindow < m.Codex.ContextWindow {
		return fmt.Errorf("model %q max_context_window must be at least context_window", slug)
	}
	if m.Codex.EffectiveContextWindowPercent < 1 || m.Codex.EffectiveContextWindowPercent > 100 {
		return fmt.Errorf("model %q effective_context_window_percent must be between 1 and 100", slug)
	}
	efforts := make(map[string]bool)
	for _, level := range m.Codex.SupportedReasoningLevels {
		switch level.Effort {
		case "minimal", "low", "medium", "high", "xhigh":
		default:
			return fmt.Errorf("model %q has invalid reasoning effort %q", slug, level.Effort)
		}
		efforts[level.Effort] = true
	}
	if !efforts[m.Codex.DefaultReasoningLevel] {
		return fmt.Errorf("model %q default reasoning level %q is not supported", slug, m.Codex.DefaultReasoningLevel)
	}
	return nil
}

func validateVariant(slug string, model ModelProfile, variant ContextVariant) error {
	if variant.ContextWindow <= 0 || variant.ContextWindow%1000 != 0 {
		return fmt.Errorf("model %q context variant must be a positive multiple of 1000", slug)
	}
	if variant.ContextWindow > model.Codex.MaxContextWindow {
		return fmt.Errorf("model %q context variant %d exceeds max_context_window %d", slug, variant.ContextWindow, model.Codex.MaxContextWindow)
	}
	if variant.EffectiveContextWindowPercent < 1 || variant.EffectiveContextWindowPercent > 100 {
		return fmt.Errorf("model %q context variant effective_context_window_percent must be between 1 and 100", slug)
	}
	return nil
}

func applyCodexDefaults(slug string, p *CodexProfile) {
	if p.DisplayName == "" {
		p.DisplayName = slug
	}
	if p.Description == "" {
		p.Description = p.DisplayName
	}
	if p.ContextWindow == 0 {
		p.ContextWindow = 131072
	}
	if p.MaxContextWindow == 0 {
		p.MaxContextWindow = p.ContextWindow
	}
	if p.EffectiveContextWindowPercent == 0 {
		p.EffectiveContextWindowPercent = 95
	}
	if p.DefaultReasoningLevel == "" {
		p.DefaultReasoningLevel = "medium"
	}
	if len(p.SupportedReasoningLevels) == 0 {
		p.SupportedReasoningLevels = []ReasoningLevel{{Effort: "low", Description: "Fast reasoning"}, {Effort: "medium", Description: "Balanced reasoning"}, {Effort: "high", Description: "Deep reasoning"}}
	}
	if len(p.InputModalities) == 0 {
		p.InputModalities = []string{"text"}
	}
	if p.TruncationLimit == 0 {
		p.TruncationLimit = 10000
	}
}

func validResponsesMode(mode ResponsesMode) bool {
	return mode == ResponsesNative || mode == ResponsesChatPolyfill || mode == ResponsesUnsupported
}

func validAdapter(adapter string) bool {
	switch adapter {
	case "native", "openai-chat", "single-system-message", "strict-text-only":
		return true
	default:
		return false
	}
}

func (c Config) ModelNames() []string {
	names := make([]string, 0, len(c.Models))
	for name := range c.Models {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}


func applyAutoRouteDefaults(r *AutoRouteConfig) {
	if !r.Enabled {
		return
	}
	if r.DefaultProfile == "" {
		r.DefaultProfile = "balanced"
	}
	defaults := defaultRouteProfiles()
	if r.Profiles == nil {
		r.Profiles = make(map[string]RouteProfile, len(defaults))
	}
	for name, profile := range defaults {
		if _, exists := r.Profiles[name]; !exists {
			r.Profiles[name] = profile
		}
	}
	if r.Aliases == nil {
		r.Aliases = map[string]string{}
	}
	aliases := map[string]string{
		"auto": "balanced",
		"auto:quality": "quality",
		"auto:fast": "fast",
		"auto:cheap": "cheap",
		"auto:code": "code",
		"auto:private": "private",
	}
	for alias, profile := range aliases {
		if _, exists := r.Aliases[alias]; !exists {
			r.Aliases[alias] = profile
		}
	}
}

func defaultRouteProfiles() map[string]RouteProfile {
	return map[string]RouteProfile{
		"balanced": {QualityWeight: 0.34, CostWeight: 0.16, LatencyWeight: 0.16, ReliabilityWeight: 0.18, AffinityWeight: 0.12, PrivacyWeight: 0.04},
		"quality":  {QualityWeight: 0.62, CostWeight: 0.04, LatencyWeight: 0.06, ReliabilityWeight: 0.16, AffinityWeight: 0.12},
		"fast":     {QualityWeight: 0.18, CostWeight: 0.10, LatencyWeight: 0.44, ReliabilityWeight: 0.20, AffinityWeight: 0.08},
		"cheap":    {QualityWeight: 0.18, CostWeight: 0.48, LatencyWeight: 0.10, ReliabilityWeight: 0.16, AffinityWeight: 0.08},
		"code":     {QualityWeight: 0.38, CostWeight: 0.08, LatencyWeight: 0.10, ReliabilityWeight: 0.16, AffinityWeight: 0.28},
		"private":  {QualityWeight: 0.20, CostWeight: 0.08, LatencyWeight: 0.08, ReliabilityWeight: 0.16, AffinityWeight: 0.08, PrivacyWeight: 0.40, PreferLocal: true},
	}
}

func (c Config) RoutingProfile(name string) (RouteProfile, bool) {
	profile, ok := c.AutoRoute.Profiles[name]
	return profile, ok
}

func validateAutoRoute(c Config) error {
	if !c.AutoRoute.Enabled {
		return nil
	}
	if _, ok := c.AutoRoute.Profiles[c.AutoRoute.DefaultProfile]; !ok {
		return fmt.Errorf("auto_route default_profile %q is not defined", c.AutoRoute.DefaultProfile)
	}
	for alias, profileName := range c.AutoRoute.Aliases {
		if strings.TrimSpace(alias) == "" || strings.TrimSpace(profileName) == "" {
			return errors.New("auto_route aliases must use non-empty alias and profile names")
		}
		if _, ok := c.AutoRoute.Profiles[profileName]; !ok {
			return fmt.Errorf("auto_route alias %q references unknown profile %q", alias, profileName)
		}
	}
	for name, p := range c.AutoRoute.Profiles {
		if strings.TrimSpace(name) == "" {
			return errors.New("auto_route profile name cannot be empty")
		}
		weights := []float64{p.QualityWeight, p.CostWeight, p.LatencyWeight, p.ReliabilityWeight, p.AffinityWeight, p.PrivacyWeight}
		total := 0.0
		for _, w := range weights {
			if w < 0 {
				return fmt.Errorf("auto_route profile %q has a negative weight", name)
			}
			total += w
		}
		if total <= 0 {
			return fmt.Errorf("auto_route profile %q must have at least one positive weight", name)
		}
		if p.MinQuality < 0 || p.MinQuality > 1 || p.MaxInputCostPerM < 0 || p.MaxOutputCostPerM < 0 {
			return fmt.Errorf("auto_route profile %q has invalid quality or cost limits", name)
		}
	}
	for slug, model := range c.Models {
		r := model.Route
		if r.Quality < 0 || r.Quality > 1 || r.Reliability < 0 || r.Reliability > 1 {
			return fmt.Errorf("model %q route quality/reliability must be between 0 and 1", slug)
		}
		if r.InputCostPerM < 0 || r.OutputCostPerM < 0 || r.LatencyMS < 0 {
			return fmt.Errorf("model %q route cost/latency values cannot be negative", slug)
		}
		if r.PrivacyTier < 0 || r.PrivacyTier > 3 {
			return fmt.Errorf("model %q route privacy_tier must be between 0 and 3", slug)
		}
	}
	return nil
}
