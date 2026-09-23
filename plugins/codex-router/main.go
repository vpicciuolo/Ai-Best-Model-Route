package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/maximhq/bifrost/core/schemas"
	"github.com/vpicciuolo/ai-best-model-route/internal/autoroute"
	"github.com/vpicciuolo/ai-best-model-route/internal/catalog"
	"github.com/vpicciuolo/ai-best-model-route/internal/config"
	"github.com/vpicciuolo/ai-best-model-route/internal/credentials"
	"github.com/vpicciuolo/ai-best-model-route/internal/editorial"
	responsescompat "github.com/vpicciuolo/ai-best-model-route/internal/responses"
)

var state struct {
	sync.RWMutex
	config config.Config
}

var metadataLookup catalog.MetadataLookup = editorial.NewOpenRouterResolver().LookupMetadata
var routeEngine *autoroute.Engine

func Init(raw any) error {
	cfg, err := config.FromAny(raw)
	if err != nil {
		return err
	}
	state.Lock()
	state.config = cfg
	state.Unlock()
	routeEngine = autoroute.New(cfg)
	return nil
}

func GetName() string { return "ai-best-model-route" }

func Cleanup() error { return nil }

func currentConfig() config.Config {
	state.RLock()
	defer state.RUnlock()
	return state.config
}

func HTTPTransportPreAuthHook(_ *schemas.BifrostContext, req *schemas.HTTPRequest) (*schemas.HTTPResponse, error) {
	if isRoutePreviewRequest(req) {
		if !credentials.VirtualKeyPresent(req.Headers) {
			return errorResponse(401, "missing_bifrost_auth", "a Bifrost virtual key is required in the x-bf-vk header"), nil
		}
		if routeEngine == nil {
			return errorResponse(503, "router_not_ready", "AI Best Model Route is not initialized"), nil
		}
		decision, err := routeEngine.Preview(req.Body)
		if err != nil {
			return errorResponse(400, "route_unavailable", err.Error()), nil
		}
		body, _ := json.Marshal(decision)
		return &schemas.HTTPResponse{StatusCode: 200, Headers: map[string]string{"Content-Type": "application/json", "Cache-Control": "no-store"}, Body: body}, nil
	}
	if !isInferenceRequest(req) {
		return nil, nil
	}
	cfg := currentConfig()
	if routeEngine != nil {
		routed, _, err := routeEngine.Rewrite(req.Body)
		if err != nil {
			return errorResponse(400, "route_unavailable", err.Error()), nil
		}
		req.Body = routed
	}
	if isResponsesRequest(req) {
		routedBody, _, err := responsescompat.ApplyHostedToolFallback(req.Body, cfg)
		if err != nil {
			return errorResponse(400, "invalid_request", "request body must be valid JSON"), nil
		}
		req.Body = routedBody
	}
	result := credentials.Decide(req.Body, cfg)
	switch result.Decision {
	case credentials.UseOpenAIPassthrough:
		if !credentials.VirtualKeyPresent(req.Headers) {
			return errorResponse(401, "missing_bifrost_auth", "a Bifrost virtual key is required in the x-bf-vk header"), nil
		}
		if !credentials.BearerPresent(req.Headers) {
			return errorResponse(401, "missing_openai_auth", "OpenAI models require Codex OpenAI authentication"), nil
		}
		credentials.SetHeader(req.Headers, credentials.DirectKeyHeader, "true")
		credentials.DeleteHeader(req.Headers, "x-api-key")
		credentials.DeleteHeader(req.Headers, "x-goog-api-key")
		return nil, nil
	case credentials.UseBifrostCredential:
		credentials.StripProviderCredentials(req.Headers)
		return nil, nil
	case credentials.Reject:
		credentials.StripProviderCredentials(req.Headers)
		return errorResponse(400, "unresolved_model", result.Message), nil
	default:
		return nil, nil
	}
}

func HTTPTransportPreHook(_ *schemas.BifrostContext, _ *schemas.HTTPRequest) (*schemas.HTTPResponse, error) {
	return nil, nil
}

func HTTPTransportPostHook(_ *schemas.BifrostContext, req *schemas.HTTPRequest, resp *schemas.HTTPResponse) error {
	if isCodexModelsRequest(req) && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, err := catalog.HydrateWithMetadata(resp.Body, currentConfig(), metadataLookup)
		if err != nil {
			return err
		}
		body, err = autoroute.InjectVirtualModels(body, currentConfig())
		if err != nil {
			return err
		}
		resp.Body = body
		if resp.Headers == nil {
			resp.Headers = make(map[string]string)
		}
		resp.Headers["Content-Type"] = "application/json"
		resp.Headers["Cache-Control"] = "no-store"
	}
	if isInferenceRequest(req) {
		model := autoroute.ModelFromBody(req.Body)
		if routeEngine != nil && model != "" {
			routeEngine.Observe(model, resp.StatusCode)
		}
		if model != "" {
			if resp.Headers == nil {
				resp.Headers = make(map[string]string)
			}
			resp.Headers["X-AI-Best-Model"] = model
			resp.Headers["X-AI-Route-Engine"] = "OYYO-AI-Best-Model-Route"
		}
	}
	return nil
}

func PreRequestHook(_ *schemas.BifrostContext, _ *schemas.BifrostRequest) error { return nil }

func PreLLMHook(ctx *schemas.BifrostContext, req *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.LLMPluginShortCircuit, error) {
	if req == nil || (req.RequestType != schemas.ResponsesRequest && req.RequestType != schemas.ResponsesStreamRequest) {
		return req, nil, nil
	}
	provider, model, _ := req.GetRequestFields()
	identity := model
	cfg := currentConfig()
	if provider != "" {
		if _, configured := cfg.Providers[string(provider)]; !configured {
			return req, shortCircuit(400, "unresolved_model", "model provider is not present in the router configuration"), nil
		}
		identity = string(provider) + "/" + model
	}
	resolved, ok := cfg.ResolveModel(identity)
	if !ok {
		return req, shortCircuit(400, "unresolved_model", "model is not present in the router catalog"), nil
	}
	// Bifrost separates the configured provider from the complete upstream model
	// ID. Reconstruct the canonical router slug before resolving it; nested model
	// IDs must retain every upstream path component.
	req.SetProvider(schemas.ModelProvider(resolved.Model.Provider))
	req.SetModel(resolved.UpstreamModel)
	switch resolved.Model.ResponsesMode {
	case config.ResponsesNative:
		return req, nil, nil
	case config.ResponsesChatPolyfill:
		adapter, ok := responsescompat.Get(resolved.Model.Adapter)
		if !ok {
			return req, shortCircuit(500, "invalid_router_config", "configured Responses adapter is not registered"), nil
		}
		if compatErr := adapter.Normalize(req.ResponsesRequest); compatErr != nil {
			return req, shortCircuit(400, compatErr.Code, compatErr.Message), nil
		}
		ctx.SetValue(schemas.BifrostContextKeyChangeRequestType, schemas.ChatCompletionRequest)
		return req, nil, nil
	case config.ResponsesUnsupported:
		return req, shortCircuit(400, "responses_unsupported", fmt.Sprintf("model %q does not support the Responses API", resolved.Slug)), nil
	default:
		return req, shortCircuit(500, "invalid_router_config", "model has no valid Responses mode"), nil
	}
}

func PostLLMHook(_ *schemas.BifrostContext, resp *schemas.BifrostResponse, bifrostErr *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) {
	return resp, bifrostErr, nil
}

func isResponsesRequest(req *schemas.HTTPRequest) bool {
	return req != nil && strings.EqualFold(req.Method, "POST") && (strings.HasSuffix(req.Path, "/v1/responses") ||
		strings.HasSuffix(req.Path, "/chatgpt_passthrough/backend-api/codex/responses"))
}

func isInferenceRequest(req *schemas.HTTPRequest) bool {
	if req == nil || !strings.EqualFold(req.Method, "POST") {
		return false
	}
	return strings.HasSuffix(req.Path, "/v1/responses") ||
		strings.HasSuffix(req.Path, "/chatgpt_passthrough/backend-api/codex/responses") ||
		strings.HasSuffix(req.Path, "/v1/chat/completions") ||
		strings.HasSuffix(req.Path, "/v1/completions")
}

func isRoutePreviewRequest(req *schemas.HTTPRequest) bool {
	return req != nil && strings.EqualFold(req.Method, "POST") && strings.HasSuffix(req.Path, "/v1/route/preview")
}

func isCodexModelsRequest(req *schemas.HTTPRequest) bool {
	return req != nil && strings.EqualFold(req.Method, "GET") &&
		strings.HasSuffix(req.Path, "/v1/models") &&
		req.CaseInsensitiveQueryLookup("client_version") != ""
}

func errorResponse(status int, code, message string) *schemas.HTTPResponse {
	body, _ := json.Marshal(map[string]any{"error": map[string]any{
		"type": "router_error", "code": code, "message": message,
	}})
	return &schemas.HTTPResponse{StatusCode: status, Headers: map[string]string{"Content-Type": "application/json"}, Body: body}
}

func shortCircuit(status int, code, message string) *schemas.LLMPluginShortCircuit {
	errorType := "router_error"
	allowFallbacks := false
	return &schemas.LLMPluginShortCircuit{Error: &schemas.BifrostError{
		IsBifrostError: true,
		StatusCode:     schemas.Ptr(status),
		AllowFallbacks: &allowFallbacks,
		Error: &schemas.ErrorField{
			Type: &errorType, Code: &code, Message: message,
		},
	}}
}
