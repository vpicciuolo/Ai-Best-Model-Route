# Integrations

AI Best Model Route follows one integration rule:

> If a client can point at an OpenAI-compatible endpoint, it can normally use `auto` without learning the provider topology underneath it.

## Codex

Codex is a first-class client because the inherited router supports the Responses API, catalog hydration, request-scoped OpenAI authentication, and managed-provider credential isolation.

Recommended flow:

1. Open this repository in Codex.
2. Ask: `Set AI Best Model Route up for me.`
3. Codex reads `AGENTS.md` and inspects existing local router state.
4. It asks which provider/API accounts should be available.
5. Provider secrets go only into the generated local `providers.env` file.
6. The setup executor installs the local Codex provider.
7. Fully quit and restart Codex.
8. Start a new task.
9. Use `auto`, `auto:quality`, `auto:fast`, `auto:cheap`, `auto:code`, or `auto:private`.

Virtual routes are injected into the Codex model catalog.

## OpenAI Python SDK

Point the client at ABMR and use a virtual model:

```python
from pathlib import Path
from openai import OpenAI

virtual_key = Path("~/.config/ai-best-model-route/virtual-key").expanduser().read_text().strip()

client = OpenAI(
    base_url="http://127.0.0.1/v1",
    api_key="abmr-managed-provider",
    default_headers={"x-bf-vk": virtual_key},
)

response = client.responses.create(
    model="nvidia-build/<model-id-from-your-catalog>",
    input="Analyze the incident timeline and identify the most likely root cause.",
)
```

The placeholder `api_key` satisfies SDK validation and is stripped before a Bifrost-managed provider call. The local gateway authorization is the `x-bf-vk` header.

Use `auto:*` from a generic SDK when the eligible routing candidates are compatible with that client. The built-in OpenAI **Codex login passthrough** is intentionally Codex-specific. For standalone OpenAI API access, configure a separate Bifrost-managed OpenAI API provider with its own provider secret.

## JavaScript / TypeScript SDKs

The same pattern applies: change the OpenAI-compatible base URL and use a virtual model. Keep provider keys in the gateway/runtime, not in browser JavaScript.

## LangChain / LangGraph

Use an OpenAI-compatible chat or Responses client pointed at ABMR and set the model to `auto` or a profile alias. ABMR performs model selection; the graph can remain provider-agnostic.

For nodes that require strict reproducibility, pin a concrete model rather than an auto route.

## OpenWebUI / LibreChat and similar clients

Add ABMR as a custom OpenAI-compatible endpoint. If the client reads the model catalog, the `auto:*` entries can appear alongside concrete models.

## Internal agents and services

Use either:

- `POST /v1/responses`
- `POST /v1/chat/completions`

Set `model` to a virtual route. For service-specific policy, add `ai_route`. ABMR strips it before provider execution.

## CI and evaluations

Use `POST /v1/route/preview` to test routing policy without consuming model inference.

Useful assertions include:

- image input never selects a text-only route;
- `require_local` never selects a remote route;
- cost caps exclude models above the configured price;
- code requests prefer evaluated code-oriented candidates;
- an open circuit stops receiving new traffic;
- unknown required capabilities fail closed.

## IDE integrations

IDE configuration changes frequently. Use a custom OpenAI-compatible endpoint only when the current IDE/version supports one. Verify current product documentation rather than copying stale setup snippets.

## Provider-specific wire protocols

A client that can only speak a provider-specific wire protocol may need an adapter. Anthropic Messages, Gemini-native APIs, and OpenAI Responses are not interchangeable solely because they can reach similar models.

Bifrost can support many provider transports below ABMR; client protocol compatibility remains a separate concern.
