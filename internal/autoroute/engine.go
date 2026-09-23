package autoroute

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vpicciuolo/ai-best-model-route/internal/config"
)

type Engine struct {
	cfg    config.Config
	mu     sync.RWMutex
	health map[string]*healthState
	now    func() time.Time
}

type healthState struct {
	SuccessEWMA float64
	Failures    int
	OpenUntil   time.Time
}

type Decision struct {
	RequestedModel       string      `json:"requested_model"`
	Profile              string      `json:"profile"`
	SelectedModel        string      `json:"selected_model"`
	Task                 string      `json:"task"`
	Signals              []string    `json:"signals"`
	EstimatedInputTokens int         `json:"estimated_input_tokens"`
	Complexity           float64     `json:"complexity"`
	Candidates           []Candidate `json:"candidates"`
}

type Candidate struct {
	Model     string   `json:"model"`
	Eligible  bool     `json:"eligible"`
	Score     float64  `json:"score,omitempty"`
	Quality   float64  `json:"quality,omitempty"`
	Cost      float64  `json:"cost_efficiency,omitempty"`
	Latency   float64  `json:"latency_efficiency,omitempty"`
	Reliable  float64  `json:"reliability,omitempty"`
	Affinity  float64  `json:"affinity,omitempty"`
	Privacy   float64  `json:"privacy,omitempty"`
	Reasons   []string `json:"reasons,omitempty"`
	Exclusion string   `json:"exclusion,omitempty"`
}

type Constraints struct {
	Profile           string   `json:"profile,omitempty"`
	MinQuality        float64  `json:"min_quality,omitempty"`
	MaxInputCostPerM  float64  `json:"max_input_cost_per_m,omitempty"`
	MaxOutputCostPerM float64  `json:"max_output_cost_per_m,omitempty"`
	RequireLocal      bool     `json:"require_local,omitempty"`
	RequireStrengths  []string `json:"require_strengths,omitempty"`
	ExcludeModels     []string `json:"exclude_models,omitempty"`
}

type requestSignals struct {
	EstimatedTokens int
	NeedsVision     bool
	HasTools        bool
	HasHostedTools  bool
	NeedsSearch     bool
	Task            string
	Signals         []string
	Complexity      float64
}

func New(cfg config.Config) *Engine {
	return &Engine{cfg: cfg, health: map[string]*healthState{}, now: time.Now}
}

func ModelFromBody(body []byte) string {
	var envelope struct {
		Model string `json:"model"`
	}
	if json.Unmarshal(body, &envelope) != nil {
		return ""
	}
	return envelope.Model
}

func (e *Engine) Rewrite(body []byte) ([]byte, *Decision, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		return body, nil, fmt.Errorf("request body must be valid JSON")
	}
	var model string
	_ = json.Unmarshal(fields["model"], &model)
	constraints, err := constraintsFrom(fields["ai_route"])
	if err != nil {
		return body, nil, err
	}
	_, hadConstraints := fields["ai_route"]
	delete(fields, "ai_route")

	profileName, auto := e.resolveProfile(model)
	if !auto {
		if !hadConstraints {
			return body, nil, nil
		}
		clean, err := json.Marshal(fields)
		return clean, nil, err
	}
	if constraints.Profile != "" {
		if _, ok := e.cfg.RoutingProfile(constraints.Profile); !ok {
			return body, nil, fmt.Errorf("unknown routing profile %q", constraints.Profile)
		}
		profileName = constraints.Profile
	}

	decision, err := e.decide(fields, model, profileName, constraints)
	if err != nil {
		return body, nil, err
	}
	encodedModel, _ := json.Marshal(decision.SelectedModel)
	fields["model"] = encodedModel
	routed, err := json.Marshal(fields)
	if err != nil {
		return body, nil, fmt.Errorf("encode routed request: %w", err)
	}
	return routed, decision, nil
}

func (e *Engine) Preview(body []byte) (*Decision, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil, fmt.Errorf("request body must be valid JSON")
	}
	var model string
	_ = json.Unmarshal(fields["model"], &model)
	constraints, err := constraintsFrom(fields["ai_route"])
	if err != nil {
		return nil, err
	}
	delete(fields, "ai_route")

	profileName, auto := e.resolveProfile(model)
	if !auto {
		if constraints.Profile != "" {
			profileName = constraints.Profile
		} else {
			profileName = e.cfg.AutoRoute.DefaultProfile
		}
		model = "auto"
	}
	if _, ok := e.cfg.RoutingProfile(profileName); !ok {
		return nil, fmt.Errorf("unknown routing profile %q", profileName)
	}
	return e.decide(fields, model, profileName, constraints)
}

func (e *Engine) resolveProfile(model string) (string, bool) {
	if !e.cfg.AutoRoute.Enabled {
		return "", false
	}
	if profileName, ok := e.cfg.AutoRoute.Aliases[model]; ok {
		return profileName, true
	}
	if model == "auto" {
		return e.cfg.AutoRoute.DefaultProfile, true
	}
	if strings.HasPrefix(model, "auto:") {
		profileName := strings.TrimPrefix(model, "auto:")
		if _, ok := e.cfg.RoutingProfile(profileName); ok {
			return profileName, true
		}
	}
	return "", false
}

func constraintsFrom(raw json.RawMessage) (Constraints, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return Constraints{}, nil
	}
	var c Constraints
	if err := json.Unmarshal(raw, &c); err != nil {
		return Constraints{}, fmt.Errorf("ai_route must be a valid object")
	}
	if c.MinQuality < 0 || c.MinQuality > 1 || c.MaxInputCostPerM < 0 || c.MaxOutputCostPerM < 0 {
		return Constraints{}, fmt.Errorf("ai_route contains invalid quality or cost limits")
	}
	return c, nil
}

func (e *Engine) decide(fields map[string]json.RawMessage, requested, profileName string, constraints Constraints) (*Decision, error) {
	profile, ok := e.cfg.RoutingProfile(profileName)
	if !ok {
		return nil, fmt.Errorf("unknown routing profile %q", profileName)
	}
	signals := inspect(fields)
	excluded := exactSet(constraints.ExcludeModels)
	requiredStrengths := normalizedSet(constraints.RequireStrengths)
	candidates := make([]Candidate, 0, len(e.cfg.Models))

	for _, slug := range e.cfg.ModelNames() {
		resolved, ok := e.cfg.ResolveModel(slug)
		if !ok || resolved.Variant != nil {
			continue
		}
		routing := resolved.Model.Route
		candidate := Candidate{Model: slug}

		if routing.Disabled {
			candidate.Exclusion = "routing disabled for model"
			candidates = append(candidates, candidate)
			continue
		}
		if excluded[slug] || excluded[resolved.UpstreamModel] {
			candidate.Exclusion = "excluded by request policy"
			candidates = append(candidates, candidate)
			continue
		}
		if resolved.Model.ResponsesMode == config.ResponsesUnsupported {
			candidate.Exclusion = "Responses API unsupported"
			candidates = append(candidates, candidate)
			continue
		}
		if e.circuitOpen(slug) {
			candidate.Exclusion = "temporary circuit open after repeated upstream failures"
			candidates = append(candidates, candidate)
			continue
		}

		effectiveContext := resolved.Model.Codex.MaxContextWindow
		if effectiveContext <= 0 {
			effectiveContext = resolved.Model.Codex.ContextWindow
		}
		if pct := resolved.Model.Codex.EffectiveContextWindowPercent; pct > 0 && pct < 100 {
			effectiveContext = effectiveContext * pct / 100
		}
		if effectiveContext > 0 && int64(signals.EstimatedTokens) > effectiveContext {
			candidate.Exclusion = fmt.Sprintf("estimated input exceeds effective context window %d", effectiveContext)
			candidates = append(candidates, candidate)
			continue
		}
		if signals.NeedsVision && !containsFold(resolved.Model.Codex.InputModalities, "image") {
			candidate.Exclusion = "request contains image input but model is text-only"
			candidates = append(candidates, candidate)
			continue
		}
		if signals.HasHostedTools && resolved.Model.ResponsesMode != config.ResponsesNative {
			candidate.Exclusion = "hosted/server-side tools require a native Responses route"
			candidates = append(candidates, candidate)
			continue
		}
		if signals.NeedsSearch && !resolved.Model.Codex.SupportsSearch {
			candidate.Exclusion = "request needs hosted search but model does not advertise search"
			candidates = append(candidates, candidate)
			continue
		}
		if constraints.RequireLocal && !routing.Local {
			candidate.Exclusion = "request requires local execution"
			candidates = append(candidates, candidate)
			continue
		}

		quality := defaultIfZero(routing.Quality, 0.60)
		reliability := defaultIfZero(routing.Reliability, 0.99) * e.healthFactor(slug)
		minQuality := maxFloat(profile.MinQuality, constraints.MinQuality)
		if minQuality > 0 && quality < minQuality {
			candidate.Exclusion = fmt.Sprintf("quality %.3f below required %.3f", quality, minQuality)
			candidates = append(candidates, candidate)
			continue
		}
		maxInput := positiveMin(profile.MaxInputCostPerM, constraints.MaxInputCostPerM)
		if maxInput > 0 && routing.InputCostPerM > maxInput {
			candidate.Exclusion = fmt.Sprintf("input cost %.4f exceeds cap %.4f", routing.InputCostPerM, maxInput)
			candidates = append(candidates, candidate)
			continue
		}
		maxOutput := positiveMin(profile.MaxOutputCostPerM, constraints.MaxOutputCostPerM)
		if maxOutput > 0 && routing.OutputCostPerM > maxOutput {
			candidate.Exclusion = fmt.Sprintf("output cost %.4f exceeds cap %.4f", routing.OutputCostPerM, maxOutput)
			candidates = append(candidates, candidate)
			continue
		}

		strengths := normalizedSet(routing.Strengths)
		missingStrength := ""
		for need := range requiredStrengths {
			if !strengths[need] {
				missingStrength = need
				break
			}
		}
		if missingStrength != "" {
			candidate.Exclusion = "missing required strength: " + missingStrength
			candidates = append(candidates, candidate)
			continue
		}

		cost := costEfficiency(routing)
		latency := latencyEfficiency(routing.LatencyMS)
		affinity := affinityScore(strengths, signals.Signals, profileName)
		privacy := float64(routing.PrivacyTier) / 3.0
		if routing.Local {
			privacy = 1
		}
		if profile.PreferLocal && routing.Local {
			affinity = clamp(affinity+0.15, 0, 1)
		}
		qualityMetric := clamp(quality*(0.70+0.30*signals.Complexity)+(1-signals.Complexity)*0.08, 0, 1)
		score := weighted(profile, qualityMetric, cost, latency, reliability, affinity, privacy)

		candidate.Eligible = true
		candidate.Score = round6(score)
		candidate.Quality = round6(quality)
		candidate.Cost = round6(cost)
		candidate.Latency = round6(latency)
		candidate.Reliable = round6(reliability)
		candidate.Affinity = round6(affinity)
		candidate.Privacy = round6(privacy)
		candidate.Reasons = []string{
			"eligible after capability and policy filters",
			"profile=" + profileName,
			"request=" + strings.Join(signals.Signals, ","),
		}
		candidates = append(candidates, candidate)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Eligible != candidates[j].Eligible {
			return candidates[i].Eligible
		}
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Model < candidates[j].Model
		}
		return candidates[i].Score > candidates[j].Score
	})
	if len(candidates) == 0 || !candidates[0].Eligible {
		return nil, fmt.Errorf("no configured model satisfies the request and routing policy")
	}

	return &Decision{
		RequestedModel: requested,
		Profile: profileName,
		SelectedModel: candidates[0].Model,
		Task: signals.Task,
		Signals: signals.Signals,
		EstimatedInputTokens: signals.EstimatedTokens,
		Complexity: round6(signals.Complexity),
		Candidates: candidates,
	}, nil
}

func inspect(fields map[string]json.RawMessage) requestSignals {
	raw, _ := json.Marshal(fields)
	var root map[string]any
	_ = json.Unmarshal(raw, &root)

	var text strings.Builder
	needsVision := false
	walk(root, &text, &needsVision)
	lower := strings.ToLower(text.String())
	chars := text.Len()
	estimatedTokens := chars / 4
	if estimatedTokens < 1 && chars > 0 {
		estimatedTokens = 1
	}

	toolTypes := extractToolTypes(root["tools"])
	hasTools := len(toolTypes) > 0
	hostedTools := false
	needsSearch := false
	for _, toolType := range toolTypes {
		if isHostedTool(toolType) {
			hostedTools = true
		}
		if strings.HasPrefix(toolType, "web_search") || toolType == "x_search" {
			needsSearch = true
		}
	}

	codeTask := hasAny(lower,
		"debug", "refactor", "code", "function", "class ", "compile", "runtime",
		"typescript", "javascript", "python", "golang", " rust", " sql", " api",
		"repository", "pull request", "unit test", "stack trace", "exception",
		"docker", "kubernetes", "architecture",
	) || strings.Contains(lower, "'''")
	reasoningTask := hasAny(lower,
		"analyze", "analyse", "reason", "prove", "derive", "compare", "trade-off",
		"tradeoff", "root cause", "step by step", "deep", "strategy", "design",
		"evaluate", "why",
	)
	longContext := estimatedTokens >= 32000

	signals := []string{}
	if codeTask {
		signals = append(signals, "code")
	}
	if reasoningTask {
		signals = append(signals, "reasoning")
	}
	if needsVision {
		signals = append(signals, "vision")
	}
	if hasTools {
		signals = append(signals, "tools")
	}
	if longContext {
		signals = append(signals, "long-context")
	}
	if estimatedTokens < 1200 && !reasoningTask && !codeTask {
		signals = append(signals, "fast")
	}
	if len(signals) == 0 {
		signals = append(signals, "general")
	}

	complexity := 0.18
	if codeTask {
		complexity += 0.22
	}
	if reasoningTask {
		complexity += 0.30
	}
	if hasTools {
		complexity += 0.12
	}
	if needsVision {
		complexity += 0.10
	}
	switch {
	case estimatedTokens >= 64000:
		complexity += 0.22
	case estimatedTokens >= 16000:
		complexity += 0.12
	case estimatedTokens >= 4000:
		complexity += 0.05
	}

	signals = uniqueSorted(signals)
	return requestSignals{
		EstimatedTokens: estimatedTokens,
		NeedsVision: needsVision,
		HasTools: hasTools,
		HasHostedTools: hostedTools,
		NeedsSearch: needsSearch,
		Task: signals[0],
		Signals: signals,
		Complexity: clamp(complexity, 0, 1),
	}
}

func walk(value any, text *strings.Builder, vision *bool) {
	switch x := value.(type) {
	case map[string]any:
		for key, item := range x {
			lowerKey := strings.ToLower(key)
			if lowerKey == "model" || lowerKey == "ai_route" {
				continue
			}
			if lowerKey == "image_url" {
				*vision = true
			}
			if lowerKey == "type" {
				if value, ok := item.(string); ok && strings.Contains(strings.ToLower(value), "image") {
					*vision = true
				}
			}
			walk(item, text, vision)
		}
	case []any:
		for _, item := range x {
			walk(item, text, vision)
		}
	case string:
		if x != "" {
			text.WriteString(x)
			text.WriteByte(' ')
		}
	}
}

func extractToolTypes(value any) []string {
	items, _ := value.([]any)
	out := []string{}
	var visit func([]any)
	visit = func(list []any) {
		for _, raw := range list {
			item, _ := raw.(map[string]any)
			if item == nil {
				continue
			}
			if toolType, ok := item["type"].(string); ok && toolType != "" {
				out = append(out, strings.ToLower(toolType))
			}
			if nested, ok := item["tools"].([]any); ok {
				visit(nested)
			}
		}
	}
	visit(items)
	return uniqueSorted(out)
}

func isHostedTool(toolType string) bool {
	toolType = strings.ToLower(strings.TrimSpace(toolType))
	switch toolType {
	case "file_search", "computer_use_preview", "web_fetch", "mcp", "code_interpreter",
		"image_generation", "memory", "tool_search", "x_search", "advisor":
		return true
	}
	return strings.HasPrefix(toolType, "web_search") ||
		strings.HasPrefix(toolType, "computer_") ||
		strings.HasPrefix(toolType, "code_execution") ||
		strings.HasPrefix(toolType, "tool_search_tool_") ||
		strings.HasPrefix(toolType, "memory_") ||
		strings.HasPrefix(toolType, "advisor_")
}

func weighted(profile config.RouteProfile, quality, cost, latency, reliability, affinity, privacy float64) float64 {
	total := profile.QualityWeight + profile.CostWeight + profile.LatencyWeight +
		profile.ReliabilityWeight + profile.AffinityWeight + profile.PrivacyWeight
	if total <= 0 {
		return 0
	}
	return (profile.QualityWeight*quality +
		profile.CostWeight*cost +
		profile.LatencyWeight*latency +
		profile.ReliabilityWeight*reliability +
		profile.AffinityWeight*affinity +
		profile.PrivacyWeight*privacy) / total
}

func costEfficiency(route config.RouteModelProfile) float64 {
	if route.InputCostPerM == 0 && route.OutputCostPerM == 0 {
		if route.Local {
			return 1
		}
		return 0.5
	}
	average := (route.InputCostPerM + route.OutputCostPerM) / 2
	return clamp(1/(1+average/5), 0, 1)
}

func latencyEfficiency(ms float64) float64 {
	if ms <= 0 {
		return 0.5
	}
	return clamp(1/(1+ms/1000), 0, 1)
}

func affinityScore(strengths map[string]bool, signals []string, profileName string) float64 {
	if len(signals) == 0 {
		return 0.5
	}
	hits := 0.0
	for _, signal := range signals {
		if strengths[normalize(signal)] {
			hits++
		}
	}
	score := hits / float64(len(signals))
	if strengths["general"] {
		score += 0.10
	}
	if profileName == "code" && strengths["code"] {
		score += 0.25
	}
	return clamp(score, 0, 1)
}

func (e *Engine) Observe(model string, status int) {
	if status == 0 || model == "" {
		return
	}
	failure := status == 429 || status >= 500
	success := status >= 200 && status < 400
	if !failure && !success {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	state := e.health[model]
	if state == nil {
		state = &healthState{SuccessEWMA: 1}
		e.health[model] = state
	}
	if success {
		state.SuccessEWMA = 0.80*state.SuccessEWMA + 0.20
		state.Failures = 0
		state.OpenUntil = time.Time{}
		return
	}
	state.SuccessEWMA = 0.80 * state.SuccessEWMA
	state.Failures++
	if state.Failures >= 3 {
		state.OpenUntil = e.now().Add(30 * time.Second)
	}
}

func (e *Engine) circuitOpen(model string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	state := e.health[model]
	return state != nil && !state.OpenUntil.IsZero() && e.now().Before(state.OpenUntil)
}

func (e *Engine) healthFactor(model string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	state := e.health[model]
	if state == nil || state.SuccessEWMA <= 0 {
		return 1
	}
	return clamp(state.SuccessEWMA, 0.05, 1)
}

func InjectVirtualModels(body []byte, cfg config.Config) ([]byte, error) {
	if !cfg.AutoRoute.Enabled {
		return body, nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode model catalog for virtual routes: %w", err)
	}
	field := raw["models"]
	if len(field) == 0 {
		field = raw["data"]
	}
	var models []map[string]any
	if len(field) > 0 {
		if err := json.Unmarshal(field, &models); err != nil {
			return nil, fmt.Errorf("decode catalog models for virtual routes: %w", err)
		}
	}

	seen := map[string]bool{}
	for _, model := range models {
		if slug, _ := model["slug"].(string); slug != "" {
			seen[slug] = true
		}
		if id, _ := model["id"].(string); id != "" {
			seen[id] = true
		}
	}

	var maxContext int64 = 131072
	modalities := map[string]bool{"text": true}
	search, summaries, verbosity, imageDetail := false, false, false, false
	for _, slug := range cfg.ModelNames() {
		resolved, ok := cfg.ResolveModel(slug)
		if !ok || resolved.Model.Route.Disabled {
			continue
		}
		if resolved.Model.Codex.MaxContextWindow > maxContext {
			maxContext = resolved.Model.Codex.MaxContextWindow
		}
		for _, modality := range resolved.Model.Codex.InputModalities {
			modalities[modality] = true
		}
		search = search || resolved.Model.Codex.SupportsSearch
		summaries = summaries || resolved.Model.Codex.SupportsReasoningSummaries
		verbosity = verbosity || resolved.Model.Codex.SupportsVerbosity
		imageDetail = imageDetail || resolved.Model.Codex.SupportsImageDetailOriginal
	}
	inputModalities := make([]string, 0, len(modalities))
	for modality := range modalities {
		inputModalities = append(inputModalities, modality)
	}
	sort.Strings(inputModalities)

	aliases := make([]string, 0, len(cfg.AutoRoute.Aliases))
	for alias := range cfg.AutoRoute.Aliases {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)

	for _, alias := range aliases {
		if seen[alias] {
			continue
		}
		profileName := cfg.AutoRoute.Aliases[alias]
		models = append(models, map[string]any{
			"slug": alias,
			"id": alias,
			"display_name": "Auto · " + title(profileName) + " — OYYO Intelligence",
			"description": "AI Best Model Route automatically selects the best eligible configured model for the " + profileName + " profile.",
			"default_reasoning_level": "medium",
			"supported_reasoning_levels": []map[string]string{
				{"effort": "low", "description": "Fast reasoning"},
				{"effort": "medium", "description": "Balanced reasoning"},
				{"effort": "high", "description": "Deep reasoning"},
			},
			"shell_type": "unified_exec",
			"visibility": "list",
			"supported_in_api": true,
			"priority": 0,
			"additional_speed_tiers": []string{},
			"service_tiers": []any{},
			"base_instructions": cfg.Instructions,
			"model_messages": map[string]any{"instructions_template": cfg.Instructions},
			"supports_reasoning_summary_parameter": summaries,
			"default_reasoning_summary": "auto",
			"support_verbosity": verbosity,
			"supports_image_detail_original": imageDetail,
			"context_window": maxContext,
			"max_context_window": maxContext,
			"effective_context_window_percent": 95,
			"experimental_supported_tools": []string{},
			"input_modalities": inputModalities,
			"supports_search_tool": search,
			"responses_mode": "native",
		})
	}
	encoded, err := json.Marshal(models)
	if err != nil {
		return nil, err
	}
	raw["models"] = encoded
	delete(raw, "data")
	return json.Marshal(raw)
}

func normalizedSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		if normalized := normalize(value); normalized != "" {
			out[normalized] = true
		}
	}
	return out
}

func exactSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(strings.ReplaceAll(value, "_", "-")))
}

func containsFold(values []string, wanted string) bool {
	for _, value := range values {
		if strings.EqualFold(value, wanted) {
			return true
		}
	}
	return false
}

func hasAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func defaultIfZero(value, fallback float64) float64 {
	if value == 0 {
		return fallback
	}
	return value
}

func positiveMin(a, b float64) float64 {
	if a <= 0 {
		return b
	}
	if b <= 0 {
		return a
	}
	return math.Min(a, b)
}

func maxFloat(a, b float64) float64 {
	return math.Max(a, b)
}

func clamp(value, low, high float64) float64 {
	return math.Max(low, math.Min(high, value))
}

func round6(value float64) float64 {
	return math.Round(value*1e6) / 1e6
}

func uniqueSorted(values []string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	out := make([]string, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func title(value string) string {
	if value == "" {
		return value
	}
	runes := []rune(strings.ToLower(value))
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}
