# AI Best Model Route

<p align="center">
  <strong>One request. The right model. Automatically.</strong><br/>
  <sub>An open-source OYYO AI Intelligence spin-off by 
    HRN Innovation Technologies LTD.</sub>
</p>

<p align="center">
  <a href="https://oyyo.one"><img alt="OYYO AI Intelligence" src="https://img.shields.io/badge/OYYO-AI%20Intelligence-111111?style=for-the-badge"></a>
  <a href="https://hrn.ae"><img alt="HRN Innovation Technologies" src="https://img.shields.io/badge/HRN-Innovation%20Technologies-202020?style=for-the-badge"></a>
  <a href="https://build.nvidia.com"><img alt="NVIDIA Build Ready" src="https://img.shields.io/badge/NVIDIA%20Build-Ready-76B900?style=for-the-badge&logo=nvidia&logoColor=white"></a>
  <a href="https://github.com/vpicciuolo/Ai-Best-Model-Route/actions/workflows/ci.yml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/vpicciuolo/Ai-Best-Model-Route/ci.yml?branch=main&style=for-the-badge&label=CI"></a>
</p>

<p align="center">
  <a href="https://github.com/vpicciuolo/Ai-Best-Model-Route/blob/main/LICENSE"><img alt="Apache-2.0" src="https://img.shields.io/github/license/vpicciuolo/Ai-Best-Model-Route?style=flat-square"></a>
  <a href="https://go.dev"><img alt="Go" src="https://img.shields.io/badge/Go-1.27+-00ADD8?style=flat-square&logo=go&logoColor=white"></a>
  <a href="https://www.getbifrost.ai"><img alt="Bifrost" src="https://img.shields.io/badge/Bifrost-Execution%20Plane-5B5BD6?style=flat-square"></a>
  <a href="https://developers.openai.com/codex/"><img alt="Codex Ready" src="https://img.shields.io/badge/Codex-Ready-000000?style=flat-square&logo=openai&logoColor=white"></a>
  <a href="https://www.docker.com"><img alt="Docker" src="https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker&logoColor=white"></a>
  <a href="https://github.com/vpicciuolo/Ai-Best-Model-Route/stargazers"><img alt="GitHub stars" src="https://img.shields.io/github/stars/vpicciuolo/Ai-Best-Model-Route?style=flat-square"></a>
</p>

<p align="center">
  <a href="#nvidia-build--one-key-setup">NVIDIA Build</a> ·
  <a href="#fastest-codex-setup">Codex Setup</a> ·
  <a href="docs/complete-setup.md">Complete Setup Guide</a> ·
  <a href="docs/routing-engine.md">Routing Engine</a> ·
  <a href="docs/integrations.md">Integrations</a> ·
  <a href="docs/roadmap.md">Roadmap</a>
</p>

> **One request. The right model. Automatically.**

**AI Best Model Route** (ABMR) is an open-source, local-first AI model routing layer created as a spin-off of **OYYO AI Intelligence** by HRN Innovation Technologies.

Instead of forcing a developer or an AI agent to choose a provider and model before every task, ABMR exposes virtual models such as `auto`, `auto:quality`, `auto:fast`, `auto:cheap`, `auto:code`, and `auto:private`. The router inspects the request, applies hard capability and policy constraints, scores the eligible models, and forwards the request through Bifrost to the selected model/provider.

The design goal is simple:

**the human asks for the result; the routing layer decides which intelligence should do the work.**

This repository starts from the excellent Apache-2.0 `applyinnovations/bifrost-model-router` project and extends the concept from a Codex-oriented multi-provider bridge into an explainable best-model decision fabric. Bifrost remains the high-performance execution plane. ABMR adds the OYYO-derived decision layer.

---

## NVIDIA Build — one-key setup

Want the NVIDIA API Catalog available inside Codex without maintaining a model list by hand?

```bash
./scripts/setup-nvidia-build.sh
```

Paste your **NVIDIA API key** when prompted. The installer validates the key against NVIDIA's current model endpoint, configures the OpenAI-compatible NVIDIA Build provider at `integrate.api.nvidia.com`, enables account-aware model discovery, and installs AI Best Model Route locally.

After Codex restarts, the NVIDIA models returned for **your API key** are exposed dynamically. You do not need to hard-code the NVIDIA catalog, and future account-visible models can appear without editing the router.

> Automatic discovery means **access**, not an automatic quality endorsement. Newly discovered NVIDIA models are available for direct selection; only reviewed models with explicit routing metadata should participate in the `auto:*` decision engine.

Full instructions: **[NVIDIA Build integration](docs/nvidia-build.md)**.

## Why this exists

Multi-model AI is powerful, but the user experience is still backwards.

A developer normally has to know:

- which provider owns the model;
- whether the model supports Responses or only Chat Completions;
- whether it supports tools, images, structured output, long context, or reasoning;
- which provider is currently healthy;
- which model is fast enough;
- which model is cheap enough;
- which model is strong enough;
- whether a request must remain local/private;
- what to do when a provider fails or rate-limits;
- how to expose all of this cleanly to Codex and OpenAI-compatible clients.

OYYO's philosophy is that the user should not become the middleware between models.

ABMR moves that decision into infrastructure.

---

## What is different

The upstream router already provides important foundations: Codex/OpenAI credential isolation, account-aware model discovery, Responses-to-Chat compatibility, model catalog hydration, hosted-tool fallback, local Docker setup, and a Bifrost-native plugin architecture.

ABMR adds a separate **decision plane** before execution:

1. **Virtual auto models** – use `auto` instead of choosing a concrete model.
2. **Task-aware classification** – coding, reasoning, long-context, vision, tool-use, fast/simple work.
3. **Hard capability filters** – a model is removed before scoring if it cannot safely satisfy the request.
4. **Multi-objective scoring** – quality, cost, latency, reliability, task affinity, and privacy.
5. **Routing profiles** – balanced, quality, fast, cheap, code, and private.
6. **Per-request constraints** – optional local-only, minimum quality, maximum cost, required strengths, model exclusions.
7. **Health memory** – repeated upstream failures reduce eligibility and can temporarily open a circuit.
8. **Explainability** – preview a route before spending tokens and inspect the selected model.
9. **Codex-native discovery** – virtual `auto:*` routes appear in the model catalog.
10. **Fail-closed behavior** – if no configured model satisfies the request, the router returns an explicit error instead of silently degrading capability.
11. **Local-first policy** – routing can happen locally without a second LLM call and without sending the prompt to a separate routing SaaS.
12. **Bifrost execution** – ABMR does not reimplement every provider protocol. Bifrost remains the transport, provider, credential, and compatibility layer.

---

## Architecture

```text
Codex / OpenAI SDK / LangChain / compatible client
                     |
                     v
          AI Best Model Route
          OYYO decision plane
                     |
      +--------------+--------------+
      | classify request             |
      | derive hard requirements     |
      | apply policy / constraints   |
      | score eligible models        |
      | health / circuit state       |
      | select concrete route        |
      +--------------+--------------+
                     |
                     v
                Bifrost
          execution / provider plane
                     |
     +---------------+-------------------+
     | OpenAI | Anthropic | Gemini | ... |
     | Azure  | Bedrock   | Groq   | ... |
     | OpenRouter | Ollama | local | ... |
     +-----------------------------------+
```

The separation is intentional. **Selection** and **execution** are different problems.

ABMR decides *what should run*. Bifrost decides *how to reach it safely and efficiently*.

---

## Auto routing modes

| Virtual model | Intent |
| --- | --- |
| `auto` | Balanced default: quality + cost + latency + reliability |
| `auto:quality` | Prefer the strongest eligible model |
| `auto:fast` | Prefer low-latency models while preserving capability |
| `auto:cheap` | Minimize cost after capability constraints |
| `auto:code` | Bias toward models tagged for software engineering |
| `auto:private` | Prefer local/private routes and stronger privacy tiers |

Profiles are configuration, not hard-coded provider marketing. You can change the weights or add your own profile.

Examples:

```text
auto:research
auto:enterprise
auto:vision
auto:agent
auto:offline
```

---

## Decision pipeline

For an `auto:*` request, ABMR follows this order:

### 1. Inspect

The router reads only the request needed for routing:

- requested virtual model/profile;
- textual request content;
- approximate input size;
- tools;
- image content;
- reasoning hints;
- optional `ai_route` constraints.

### 2. Derive hard requirements

Examples:

- image input -> model must advertise image input;
- hosted/server-side tool -> route must be native Responses capable;
- long input -> model must have sufficient context;
- local-only constraint -> remote routes are rejected;
- excluded model -> removed from candidates.

Hard requirements are **filters**, not soft preferences.

### 3. Classify task affinity

The built-in v0.1 classifier is intentionally lightweight and local. It detects signals for:

- coding/debugging/architecture;
- deep reasoning/analysis;
- tool-use;
- vision;
- long-context;
- simple/fast conversational work.

No second model call is required just to decide which model to use.

### 4. Score eligible candidates

Each candidate can carry normalized routing metadata:

- quality score;
- input/output cost;
- expected latency;
- reliability baseline;
- strengths/tags;
- privacy tier;
- local execution flag.

The selected profile supplies the weights.

### 5. Apply health state

Repeated failures reduce a route's effective reliability. After a configurable failure streak the route can temporarily enter an open-circuit state so traffic moves elsewhere.

### 6. Select deterministically

The highest scoring eligible route wins. Ties are resolved deterministically by canonical model slug, which keeps routing reproducible.

---

## Fastest Codex setup

This repository is designed to be understandable by Codex itself.

1. Open/clone this repository in Codex.
2. Ask:

```text
Set AI Best Model Route up for me.
```

3. Codex reads `AGENTS.md`, inspects your existing configuration, asks which provider/API accounts you want to make available, discovers account-visible models where possible, and prepares the local config.
4. Provider secrets are placed only in a local mode-`0600` env file. Do **not** paste provider API keys into chat or commit them.
5. The setup installs the local provider in Codex.
6. Fully restart Codex and create a **new task**.
7. Choose `auto` or one of the `auto:*` profiles.

The router preserves direct OpenAI-through-Codex authentication for declared OpenAI passthrough routes and strips that credential from managed-provider routes.

See **[docs/codex-setup.md](docs/codex-setup.md)** for the full flow.

---

## Direct OpenAI-compatible use

ABMR can also sit behind software that lets you set an OpenAI-compatible base URL.

Conceptually:

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
    input="Review this architecture and find the highest-risk design flaw."
)
```

For generic SDKs, the `api_key` placeholder only satisfies the client library; ABMR strips it before managed-provider execution. The actual local authorization is `x-bf-vk`. Codex's special OpenAI login passthrough is Codex-specific; if a standalone client needs OpenAI API models, configure OpenAI as a normal Bifrost-managed provider instead of relying on the Codex passthrough route.

For Chat Completions-compatible clients:

```python
response = client.chat.completions.create(
    model="auto:fast",
    messages=[{"role": "user", "content": "Summarize this in five bullets."}],
)
```

The concrete upstream provider credential remains owned by the local Bifrost layer.

---

## cURL

```bash
curl http://127.0.0.1/v1/responses \
  -H "content-type: application/json" \
  -H "x-bf-vk: $BIFROST_API_KEY" \
  -d '{
    "model": "auto:code",
    "input": "Find the concurrency bug in this Go service and propose a minimal fix."
  }'
```

---

## Per-request routing constraints

Applications can add an `ai_route` object. ABMR consumes and removes it before the request is sent upstream.

```json
{
  "model": "auto",
  "input": "Analyze this contract",
  "ai_route": {
    "profile": "quality",
    "min_quality": 0.85,
    "max_input_cost_per_m": 5,
    "max_output_cost_per_m": 20,
    "require_local": false,
    "require_strengths": ["reasoning"],
    "exclude_models": ["provider/model-to-avoid"]
  }
}
```

This makes routing policy a request-level concern without leaking ABMR-specific fields to the model provider.

---

## Preview a route without running the model

Use the local preview endpoint with the same request body:

```bash
curl http://127.0.0.1/v1/route/preview \
  -H "content-type: application/json" \
  -H "x-bf-vk: $BIFROST_API_KEY" \
  -d '{
    "model": "auto:code",
    "input": "Refactor a distributed Go service with race conditions."
  }'
```

The response explains the selected route and ranked candidates. This is useful for:

- tuning profile weights;
- verifying privacy constraints;
- debugging why a model was excluded;
- CI policy tests;
- measuring how a configuration change affects routing before spending tokens.

---

## Model routing metadata

Concrete model entries can add an optional `route` section:

```yaml
models:
  openai/example-strong-model:
    aliases: [example-strong-model]
    codex:
      context_window: 200000
      input_modalities: [text, image]
      supports_search: true
    route:
      quality: 0.95
      input_cost_per_m: 2.50
      output_cost_per_m: 15.00
      latency_ms: 1200
      reliability: 0.995
      strengths: [reasoning, code, tools]
      privacy_tier: 2
      local: false
```

Routing metadata should be maintained from real measurements and provider pricing, not guessed marketing claims.

---

## Routing profiles

```yaml
auto_route:
  enabled: true
  default_profile: balanced

  aliases:
    auto: balanced
    "auto:quality": quality
    "auto:fast": fast
    "auto:cheap": cheap
    "auto:code": code
    "auto:private": private

  profiles:
    balanced:
      quality_weight: 0.34
      cost_weight: 0.16
      latency_weight: 0.16
      reliability_weight: 0.18
      affinity_weight: 0.12
      privacy_weight: 0.04

    quality:
      quality_weight: 0.62
      cost_weight: 0.04
      latency_weight: 0.06
      reliability_weight: 0.16
      affinity_weight: 0.12

    fast:
      quality_weight: 0.18
      cost_weight: 0.10
      latency_weight: 0.44
      reliability_weight: 0.20
      affinity_weight: 0.08
```

Weights are normalized by the router, so they do not need to sum to exactly 1.0.

---

## Using it with other tools

Any product that can target an OpenAI-compatible endpoint can usually use ABMR by changing the base URL and model name.

Typical integration pattern:

| Client/framework | Integration |
| --- | --- |
| OpenAI SDK | Set `base_url` + `x-bf-vk`; use managed models or `auto` with an SDK-compatible candidate set |
| Codex | Install the custom Responses provider; virtual routes appear in catalog |
| LangChain / LangGraph | Use an OpenAI-compatible chat/Responses client pointing at ABMR |
| OpenWebUI / LibreChat | Add ABMR as a custom OpenAI-compatible endpoint |
| Internal agents | Call `/v1/responses` or `/v1/chat/completions` with `auto:*` |
| CI/evaluation systems | Call `/v1/route/preview` to test policy without inference |

A client with a provider-specific wire protocol that cannot target an OpenAI-compatible endpoint requires an adapter. Do not assume protocol compatibility merely because the underlying model exists.

See **[docs/integrations.md](docs/integrations.md)**.

---

## Provider routing vs model routing

These are intentionally separate layers.

**Model routing** asks:

> Which model is the best fit for this task?

**Provider routing** asks:

> Where should this model run right now?

ABMR focuses on the first question and delegates the second to Bifrost's provider catalog, governance, load-balancing, key selection, retries, and provider-level routing capabilities where configured.

That means a future route can look like:

```text
request
  -> auto:code
  -> choose model family X
  -> Bifrost chooses healthy provider/region/key for X
  -> execute
```

This avoids mixing semantic model choice with infrastructure failover.

---

## Security principles

- Fail closed on unknown or ambiguous models.
- Never treat the caller's Codex/OpenAI bearer as a generic provider key.
- Strip caller provider credentials from managed-provider routes.
- Keep provider secrets outside the repository.
- Use environment references in Bifrost config.
- Bind the local quickstart to loopback.
- Do not log prompt/response content by default.
- Do not silently downgrade capabilities to make a request “work.”
- Keep virtual-key authorization separate from upstream-provider credentials.
- Use local deterministic routing when possible so a routing decision does not require exposing the prompt to another service.

See **[docs/security.md](docs/security.md)**.

---

## Observability and explainability

The router should make decisions inspectable.

The first implementation exposes:

- route preview;
- selected concrete model;
- profile;
- task signals;
- candidate scores;
- exclusion reasons;
- circuit/health state.

The roadmap adds OpenTelemetry route spans, decision IDs, cost reconciliation, shadow routing, offline evaluation, and learned routing models.

See **[docs/routing-engine.md](docs/routing-engine.md)** and **[docs/roadmap.md](docs/roadmap.md)**.

---

## OYYO AI Intelligence spin-off

AI Best Model Route is intentionally usable as a standalone open-source component.

It is also a direct expression of the OYYO AI architecture:

**One intelligence. One workspace. One place for everything.**

OYYO's intelligence fabric is designed around the idea that a user should describe the goal, while the system selects the right model, agent, tool, memory, connector, and execution path.

ABMR extracts one foundational part of that idea—the **model decision layer**—into a project that other developers can adopt independently.

No OYYO account is required to use the open-source router.

---

## Development

The inherited execution layer is built with Go and Bifrost. The routing layer is kept dependency-light and deterministic.

```bash
nix develop
just test
just check-config
nix flake check -L
```

The CI suite should cover:

- config validation;
- routing classification;
- hard capability filters;
- scoring determinism;
- per-request policy;
- circuit breaking;
- credential isolation;
- Responses compatibility;
- catalog hydration;
- streaming and non-streaming integration;
- setup-script behavior;
- vulnerability scanning.

---

## Project status

This repository is an active OYYO AI Intelligence spin-off. The first milestone is a robust local/open-source decision layer on top of Bifrost, with Codex as a first-class client and OpenAI-compatible clients as the general integration surface.

Planned next layers include learned routing, online evaluation, dynamic price/performance feeds, provider-level SLOs, shadow traffic, model experiments, and richer enterprise policy.

---

## OYYO open-source ecosystem

AI Best Model Route is one part of the public OYYO AI Intelligence stack. Related repositories from **@vpicciuolo**:

| Project | Purpose |
| --- | --- |
| [OYYO Models](https://github.com/vpicciuolo/oyyo-models) | OYYO model assets and model-related work |
| [OYYO SDK](https://github.com/vpicciuolo/oyyo-sdk) | Developer-facing OYYO integration surface |
| [OYYO Benchmark](https://github.com/vpicciuolo/oyyo-benchmark) | Evaluation and benchmarking work for models and routing |
| [URL Intelligence Agent](https://github.com/vpicciuolo/url-intelligence-agent) | Open-source URL analysis / intelligence agent |

<p>
  <a href="https://github.com/vpicciuolo/oyyo-models"><img alt="OYYO Models" src="https://img.shields.io/badge/GitHub-OYYO%20Models-181717?style=flat-square&logo=github"></a>
  <a href="https://github.com/vpicciuolo/oyyo-sdk"><img alt="OYYO SDK" src="https://img.shields.io/badge/GitHub-OYYO%20SDK-181717?style=flat-square&logo=github"></a>
  <a href="https://github.com/vpicciuolo/oyyo-benchmark"><img alt="OYYO Benchmark" src="https://img.shields.io/badge/GitHub-OYYO%20Benchmark-181717?style=flat-square&logo=github"></a>
  <a href="https://github.com/vpicciuolo/url-intelligence-agent"><img alt="URL Intelligence Agent" src="https://img.shields.io/badge/GitHub-URL%20Intelligence%20Agent-181717?style=flat-square&logo=github"></a>
</p>

**OYYO AI Intelligence:** https://oyyo.one  
**HRN Innovation Technologies:** https://hrn.ae  
**GitHub:** https://github.com/vpicciuolo

## Upstream and attribution

This project is a derivative work of:

- **Bifrost Model Router**, copyright 2026 Apply Innovations, Apache License 2.0.
- **Bifrost**, copyright 2025 H3 Labs Inc., Apache License 2.0.

The derivative routing and OYYO integration work is maintained under this repository by HRN Innovation Technologies / OYYO AI Intelligence.

See [NOTICE](NOTICE) for the complete attribution chain.

---

## License

Apache License 2.0. See [LICENSE](LICENSE).

---

**AI Best Model Route**  
An **OYYO AI Intelligence** spin-off.  
Built so the user can ask once—and the system can decide which model should answer.
