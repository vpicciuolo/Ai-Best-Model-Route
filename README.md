# AI Best Model Route

<h2 align="center">Open-Source LLM Router, AI Model Router & OpenAI-Compatible Multi-Provider Gateway</h2>

<p align="center">
  <strong>One request. The right model. Automatically.</strong><br/>
  Automatically choose the best AI model for every request based on
  <strong>capability, quality, cost, latency, reliability, privacy and runtime health</strong>.
</p>

<p align="center">
  <sub>An open-source <a href="https://oyyo.one">OYYO AI Intelligence</a> spin-off by <a href="https://hrn.ae">HRN Innovation Technologies LTD</a>.</sub>
</p>

<p align="center">
  <a href="https://oyyo.one"><img alt="OYYO AI Intelligence" src="https://img.shields.io/badge/OYYO-AI%20Intelligence-111111?style=for-the-badge"></a>
  <a href="https://hrn.ae"><img alt="HRN Innovation Technologies" src="https://img.shields.io/badge/HRN-Innovation%20Technologies-202020?style=for-the-badge"></a>
  <a href="https://build.nvidia.com"><img alt="NVIDIA Build and NVIDIA NIM Ready" src="https://img.shields.io/badge/NVIDIA%20Build%20%2F%20NIM-Ready-76B900?style=for-the-badge&logo=nvidia&logoColor=white"></a>
  <a href="https://github.com/vpicciuolo/Ai-Best-Model-Route/actions/workflows/ci.yml"><img alt="AI Best Model Route CI" src="https://img.shields.io/github/actions/workflow/status/vpicciuolo/Ai-Best-Model-Route/ci.yml?branch=main&style=for-the-badge&label=CI"></a>
</p>

<p align="center">
  <a href="https://github.com/vpicciuolo/Ai-Best-Model-Route/blob/main/LICENSE"><img alt="Apache 2.0 License" src="https://img.shields.io/github/license/vpicciuolo/Ai-Best-Model-Route?style=flat-square"></a>
  <a href="https://go.dev"><img alt="Go 1.27+" src="https://img.shields.io/badge/Go-1.27+-00ADD8?style=flat-square&logo=go&logoColor=white"></a>
  <a href="https://www.getbifrost.ai"><img alt="Bifrost AI Gateway Execution Plane" src="https://img.shields.io/badge/Bifrost-Execution%20Plane-5B5BD6?style=flat-square"></a>
  <a href="https://developers.openai.com/codex/"><img alt="OpenAI Codex Ready" src="https://img.shields.io/badge/OpenAI%20Codex-Ready-000000?style=flat-square&logo=openai&logoColor=white"></a>
  <a href="https://www.docker.com"><img alt="Docker Ready" src="https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker&logoColor=white"></a>
  <a href="https://github.com/vpicciuolo/Ai-Best-Model-Route/pkgs/container/ai-best-model-route"><img alt="GHCR Container" src="https://img.shields.io/badge/GHCR-Container-2496ED?style=flat-square&logo=docker&logoColor=white"></a>
  <a href="https://github.com/vpicciuolo/Ai-Best-Model-Route/stargazers"><img alt="GitHub stars" src="https://img.shields.io/github/stars/vpicciuolo/Ai-Best-Model-Route?style=flat-square"></a>
</p>

<p align="center">
  <strong>AI Model Router</strong> ·
  <strong>LLM Router</strong> ·
  <strong>AI Gateway</strong> ·
  <strong>Automatic Model Selection</strong> ·
  <strong>Multi-LLM Routing</strong> ·
  <strong>OpenAI-Compatible API</strong> ·
  <strong>NVIDIA NIM</strong> ·
  <strong>Codex Router</strong> ·
  <strong>Cost-Aware Routing</strong> ·
  <strong>Latency-Aware Routing</strong> ·
  <strong>Privacy-Aware Routing</strong> ·
  <strong>Self-Hosted</strong>
</p>

<p align="center">
  <a href="#-30-second-quickstart"><strong>Quickstart</strong></a> ·
  <a href="#-nvidia-build--nvidia-nim-one-key-setup"><strong>NVIDIA Build / NIM</strong></a> ·
  <a href="#-codex-setup"><strong>Codex</strong></a> ·
  <a href="docs/complete-setup.md"><strong>Complete Setup</strong></a> ·
  <a href="docs/routing-engine.md"><strong>Routing Engine</strong></a> ·
  <a href="docs/integrations.md"><strong>Integrations</strong></a> ·
  <a href="#-faq"><strong>FAQ</strong></a>
</p>

---

## What is AI Best Model Route?

**AI Best Model Route (ABMR)** is an open-source **LLM router**, **AI model router**, and **OpenAI-compatible multi-provider routing gateway** for applications that use more than one AI model.

Instead of hard-coding one LLM for every request, your application can call a virtual model such as `auto`, `auto:quality`, `auto:fast`, `auto:cheap`, `auto:code`, or `auto:private`. ABMR evaluates the request, removes models that cannot safely handle it, scores the remaining candidates, and sends the request to the best eligible model.

ABMR is designed for **multi-LLM applications, AI agents, coding agents, local AI stacks, NVIDIA NIM / NVIDIA Build users, OpenAI-compatible applications, Codex workflows, private AI deployments, and self-hosted AI infrastructure**.

> **The application asks once. ABMR decides which intelligence should answer.**

### In one sentence

**AI Best Model Route is a local-first automatic LLM model selector that sits in front of multiple AI providers and chooses the best model per request using capability, quality, cost, latency, reliability, privacy and health signals.**

---

## Why use an LLM router?

Modern AI applications increasingly use multiple models:

- one model may be best for **coding**;
- another may be faster for **simple chat**;
- another may be cheaper for **high-volume workloads**;
- another may be required for **vision or tools**;
- a local model may be preferred for **privacy**;
- a provider may be temporarily **rate-limited or unhealthy**;
- a long-context request may require a model with a larger **context window**.

Without a model router, the application becomes responsible for all of those decisions.

ABMR moves that logic into infrastructure.

### Common use cases

| Use case | What ABMR does |
| --- | --- |
| **Best LLM per request** | Selects the strongest eligible model for the task |
| **Reduce AI cost** | Routes cost-sensitive requests toward cheaper eligible models |
| **Lower latency** | Prefers lower-latency candidates with `auto:fast` |
| **Coding agents** | Biases routing toward models evaluated/tagged for software engineering |
| **Private AI / local LLMs** | Enforces local-only routing and privacy preferences |
| **NVIDIA NIM / Build** | Dynamically discovers models available to your NVIDIA API key |
| **OpenAI-compatible apps** | Provides one endpoint for multi-provider model access |
| **Codex** | Exposes virtual `auto:*` routes directly in the Codex model catalog |
| **Reliability / failover** | Tracks route health and temporarily avoids repeatedly failing models |
| **Long context** | Filters out models that cannot safely fit the estimated request |
| **Vision / tools** | Rejects candidates missing required modalities or hosted-tool support |
| **CI / evaluation** | Previews routing decisions without running inference |

---

## ✨ Features at a glance

- **Automatic model selection** with virtual `auto:*` models
- **Multi-provider LLM routing** behind one OpenAI-compatible API
- **Cost-aware routing** using configurable model pricing metadata
- **Latency-aware routing** using model latency metadata
- **Quality-aware routing** using deployment-specific evaluation scores
- **Reliability-aware routing** with runtime health feedback
- **Privacy-aware routing** including local/private model preference
- **Capability-first filtering** for vision, tools, context and protocol support
- **Task-aware routing** for coding, reasoning, long context, tool use and fast/simple work
- **Circuit breaking** after repeated upstream failures
- **Explainable routing** with candidate scores and exclusion reasons
- **Route preview API** before spending inference tokens
- **Per-request routing constraints** through `ai_route`
- **OpenAI Responses API + Chat Completions compatibility**
- **OpenAI SDK compatible**
- **OpenAI Codex integration**
- **NVIDIA Build / NVIDIA NIM one-key integration**
- **Dynamic authenticated model discovery**
- **Docker / GHCR distribution**
- **Local-first and self-hosted**
- **Provider credential isolation**
- **Apache 2.0 open source**

---

## ⚡ 30-second quickstart

### NVIDIA Build / NVIDIA NIM

The fastest path if you have an NVIDIA API key:

```bash
git clone https://github.com/vpicciuolo/Ai-Best-Model-Route.git
cd Ai-Best-Model-Route

./scripts/setup-nvidia-build.sh
```

Paste your NVIDIA API key into the hidden prompt.

The installer verifies the key, discovers the models available to your NVIDIA account, configures the provider, starts ABMR, and configures Codex.

Full guide: **[docs/nvidia-build.md](docs/nvidia-build.md)**

### OpenAI Codex

Open the repository in Codex and ask:

```text
Set AI Best Model Route up for me.
```

Codex reads `AGENTS.md`, discovers your provider requirements, prepares secure credential placeholders, installs the router, and configures new tasks to use `auto`.

Full guide: **[docs/codex-setup.md](docs/codex-setup.md)**

### General multi-provider setup

Use the complete installation guide for OpenAI-compatible providers, Bifrost providers, custom model metadata, routing profiles and SDK integration:

**[Complete setup guide →](docs/complete-setup.md)**

---

## 🧠 Automatic model routing

ABMR exposes virtual models instead of forcing every client to choose a concrete LLM.

| Virtual model | Routing goal |
| --- | --- |
| `auto` | Balanced quality + cost + latency + reliability |
| `auto:quality` | Prefer the strongest eligible model |
| `auto:fast` | Prefer low-latency eligible models |
| `auto:cheap` | Minimize cost after capability constraints |
| `auto:code` | Prefer models tagged/evaluated for coding |
| `auto:private` | Prefer local/private models and stronger privacy tiers |

Custom profiles can also be created, for example:

```text
auto:research
auto:enterprise
auto:vision
auto:offline
```

Profiles are configuration—not hard-coded provider marketing.

---

## How the routing engine works

ABMR uses a deterministic, inspectable routing pipeline.

```text
Request
  │
  ▼
Request inspection
  │
  ├── task signals
  ├── context size
  ├── tools
  ├── image input
  └── request policy
  │
  ▼
Hard capability & policy filters
  │
  ├── context window
  ├── modality
  ├── hosted tools
  ├── local-only
  ├── model exclusions
  └── cost / quality limits
  │
  ▼
Candidate scoring
  │
  ├── quality
  ├── cost
  ├── latency
  ├── reliability
  ├── task affinity
  └── privacy
  │
  ▼
Runtime health / circuit state
  │
  ▼
Best eligible model
  │
  ▼
Bifrost execution plane
  │
  ▼
OpenAI / NVIDIA / Anthropic / Gemini / Azure / Bedrock /
Groq / OpenRouter / Ollama / local models / other configured providers
```

Hard requirements are **filters**, not preferences. A cheap model cannot win if it cannot safely execute the request.

The built-in v1 classifier is deliberately lightweight and local. No second remote LLM call is required merely to decide which LLM to call.

---

## 🏗️ Architecture

ABMR separates **model intelligence** from **provider execution**.

```text
Codex / OpenAI SDK / LangChain / LangGraph /
OpenWebUI / LibreChat / internal agents / apps
                         │
                         ▼
              AI Best Model Route
              OYYO decision plane
                         │
        ┌────────────────┴────────────────┐
        │ classify + filter + score       │
        │ policy + health + explain       │
        └────────────────┬────────────────┘
                         │
                         ▼
                      Bifrost
                execution/provider plane
                         │
     ┌───────────────────┼────────────────────┐
     ▼                   ▼                    ▼
  OpenAI             NVIDIA NIM          Other providers
  Anthropic          NVIDIA Build        Local models
  Gemini             OpenRouter          Private models
```

**ABMR decides what should run. Bifrost decides how to reach it safely and efficiently.**

This separation prevents ABMR from becoming another monolithic provider gateway.

---

## 🔌 Providers, models and ecosystems

ABMR is provider-agnostic at the decision layer. Provider execution is handled through Bifrost and configured OpenAI-compatible endpoints.

Typical deployments can include models from:

- OpenAI
- Anthropic
- Google Gemini
- NVIDIA Build / NVIDIA NIM
- Microsoft Azure
- AWS Bedrock
- Groq
- OpenRouter
- Ollama
- local/self-hosted OpenAI-compatible servers
- other Bifrost-supported or custom OpenAI-compatible providers

Exact availability depends on your provider credentials and configuration.

### Client and framework compatibility

| Client / framework | ABMR integration |
| --- | --- |
| **OpenAI SDK** | Point `base_url` to ABMR and pass the local virtual key |
| **OpenAI Codex** | First-class integration; `auto:*` routes appear in the model catalog |
| **LangChain** | Use an OpenAI-compatible client targeting ABMR |
| **LangGraph** | Keep graph nodes provider-agnostic and route through ABMR |
| **OpenWebUI** | Add ABMR as a custom OpenAI-compatible endpoint |
| **LibreChat** | Add ABMR as an OpenAI-compatible endpoint |
| **AI agents** | Call Responses API or Chat Completions with `auto:*` |
| **CI / eval systems** | Use `/v1/route/preview` without inference |

See **[Integration guide →](docs/integrations.md)**.

---

## 🟢 NVIDIA Build / NVIDIA NIM one-key setup

ABMR includes a ready-to-use integration for NVIDIA's hosted API catalog.

```bash
./scripts/setup-nvidia-build.sh
```

The installer:

1. accepts `NVIDIA_API_KEY` from the environment, an existing protected env file, or a hidden terminal prompt;
2. validates the API key against NVIDIA's live model endpoint;
3. discovers models visible to that NVIDIA account;
4. configures `https://integrate.api.nvidia.com`;
5. securely stores the API key outside the repository;
6. merges NVIDIA into an existing standard ABMR setup without deleting other providers;
7. installs/updates the local Docker router;
8. configures Codex;
9. exposes account-visible NVIDIA models dynamically.

> **Dynamic discovery means availability—not automatic quality endorsement.** Newly discovered NVIDIA models can be selected directly, but only explicitly reviewed models should participate in automatic `auto:*` routing.

Full guide: **[NVIDIA Build / NIM integration →](docs/nvidia-build.md)**

---

## 🤖 Codex setup

AI Best Model Route is designed so Codex can understand and configure the project itself.

Open the repository in Codex and ask:

```text
Set AI Best Model Route up for me.
```

Codex follows `AGENTS.md` and can:

- inspect an existing router installation;
- preserve existing providers;
- determine provider/account scope;
- verify endpoints and model discovery;
- create secure credential placeholders;
- avoid putting provider secrets into chat or Git;
- configure model metadata;
- install the Docker router;
- configure the Codex provider;
- verify the resulting model catalog.

After setup, fully restart Codex and create a **new task**.

---

## 🐍 OpenAI SDK / OpenAI-compatible API

ABMR can sit behind any client that supports an OpenAI-compatible base URL.

```python
from pathlib import Path
from openai import OpenAI

virtual_key = Path(
    "~/.config/ai-best-model-route/virtual-key"
).expanduser().read_text().strip()

client = OpenAI(
    base_url="http://127.0.0.1/v1",
    api_key="abmr-managed-provider",
    default_headers={"x-bf-vk": virtual_key},
)

response = client.responses.create(
    model="nvidia-build/<model-id-from-your-catalog>",
    input="Review this architecture and find the highest-risk design flaw.",
)

print(response.output_text)
```

For Bifrost-managed routes, the placeholder SDK Authorization value is stripped before upstream execution. Local ABMR access is authorized by `x-bf-vk`.

For Chat Completions-compatible clients:

```python
response = client.chat.completions.create(
    model="auto:fast",
    messages=[
        {"role": "user", "content": "Summarize this in five bullets."}
    ],
)
```

> Codex's OpenAI login passthrough is Codex-specific. For standalone OpenAI Platform API access, configure OpenAI as a normal Bifrost-managed provider.

---

## 🔎 Explainable routing: preview before inference

ABMR can explain which model it would select **without running the model**.

```bash
curl http://127.0.0.1/v1/route/preview \
  -H "content-type: application/json" \
  -H "x-bf-vk: $BIFROST_API_KEY" \
  -d '{
    "model": "auto:code",
    "input": "Refactor a distributed Go service with race conditions."
  }'
```

The decision includes:

- selected model;
- routing profile;
- detected task signals;
- estimated input size;
- complexity;
- candidate models;
- candidate scores;
- exclusion reasons.

Useful for:

- routing-policy CI;
- privacy validation;
- cost-policy validation;
- troubleshooting;
- model-profile tuning;
- regression testing.

---

## 🎛️ Per-request routing policy

Applications can add an `ai_route` object. ABMR consumes it locally and removes it before forwarding the request upstream.

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

Supported constraints include:

- routing profile;
- minimum model quality;
- maximum input/output token price;
- local-only execution;
- required model strengths;
- explicit model exclusions.

Impossible policies fail closed rather than silently weakening the requirement.

---

## 📊 Model routing metadata

Automatic routing candidates can carry deployment-specific metadata:

```yaml
models:
  provider/example-model:
    provider: provider
    upstream_model: example-model

    codex:
      context_window: 200000
      input_modalities: [text, image]

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

Routing metadata should come from real provider documentation, pricing and your own measurements—not guessed marketing claims.

### Routing profiles

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
```

Weights are normalized by the router.

---

## 🛡️ Reliability, failover and circuit breaking

ABMR records route health from model responses.

Repeated HTTP `429` and `5xx` failures reduce effective reliability. After the configured failure streak, a model can temporarily enter an open-circuit state and stop receiving new automatic traffic.

Successful responses recover the route.

This complements Bifrost's provider/key execution and health behavior without duplicating the provider gateway.

---

## 🔐 Local-first security

ABMR is designed to keep routing intelligence and credentials under your control.

Security principles:

- fail closed on unknown or ambiguous models;
- do not silently downgrade required capabilities;
- keep provider API keys outside the repository;
- reference provider credentials through runtime environment variables;
- never reuse the caller's Codex/OpenAI bearer as a generic provider key;
- strip caller provider credentials from Bifrost-managed routes;
- keep the local development gateway on loopback;
- disable prompt/response content logging by default;
- keep ABMR virtual-key authorization separate from provider credentials;
- perform deterministic routing locally without sending the prompt to a separate routing SaaS.

Local files:

```text
~/.config/ai-best-model-route/providers.env
~/.config/ai-best-model-route/virtual-key
~/.config/ai-best-model-route/config.json
```

See **[Security model →](docs/security.md)**.

---

## 🐳 Docker / GHCR

Validated container images are automatically published to GitHub Container Registry.

```text
ghcr.io/vpicciuolo/ai-best-model-route:main
```

Each validated main build also receives an immutable SHA tag:

```text
ghcr.io/vpicciuolo/ai-best-model-route:sha-<commit>
```

The publish workflow runs the full validation suite before publishing the image.

---

## AI model routing vs a traditional AI gateway

A traditional **AI gateway / LLM gateway** primarily answers:

> How do I reach this model/provider reliably?

AI Best Model Route primarily answers:

> Which model should execute this request in the first place?

ABMR deliberately separates those concerns.

| Layer | Responsibility |
| --- | --- |
| **AI Best Model Route** | model selection, task affinity, policy, quality/cost/latency/privacy scoring, explainability |
| **Bifrost** | provider transport, credentials, protocol adaptation, provider/key routing, governance |

That means ABMR can be used as the **model intelligence layer** in front of a multi-provider AI gateway rather than rebuilding every provider integration from scratch.

---

## Model routing vs provider routing

**Model routing** asks:

> Which LLM is the best fit for this task?

**Provider routing** asks:

> Where should that model run right now?

Example:

```text
request
  → auto:code
  → ABMR chooses model family X
  → Bifrost chooses the configured provider/key/path
  → inference
```

This keeps semantic/task selection separate from infrastructure execution.

---

## Is ABMR a semantic router?

**v1 is a deterministic task-aware model router, not a learned embedding-based semantic router.**

ABMR v1 classifies useful request signals locally and combines those signals with hard capability filters and explicit routing metadata. This makes decisions deterministic, explainable and dependency-light.

The architecture is intentionally compatible with future learned or semantic routing layers, but those are not part of the v1 contract.

---

## ✅ Production validation

The v1 repository is validated through GitHub Actions and Nix.

The test/build pipeline includes:

- Go unit tests;
- routing-engine tests;
- gateway routing tests;
- configuration/schema validation;
- credential-isolation tests;
- Responses compatibility tests;
- catalog hydration tests;
- streaming/non-streaming integration tests;
- race checks;
- plugin-load checks;
- macOS setup tests;
- vulnerability scanning;
- container build validation;
- GHCR publishing validation.

---

## 📚 Documentation

| Guide | Purpose |
| --- | --- |
| **[Complete Setup](docs/complete-setup.md)** | End-to-end installation and configuration |
| **[NVIDIA Build / NIM](docs/nvidia-build.md)** | One-key NVIDIA provider integration |
| **[Codex Setup](docs/codex-setup.md)** | Configure ABMR for OpenAI Codex |
| **[Routing Engine](docs/routing-engine.md)** | Decision pipeline, scoring and policy |
| **[Integrations](docs/integrations.md)** | OpenAI SDK, LangChain, LangGraph and clients |
| **[Architecture](docs/architecture.md)** | Internal architecture |
| **[Configuration](docs/configuration.md)** | Router configuration reference |
| **[Security](docs/security.md)** | Credential and deployment security |
| **[Operations](docs/operations.md)** | Health, deployment and rollback |
| **[Roadmap](docs/roadmap.md)** | Future directions beyond v1 |

---

## ❓ FAQ

### What is an LLM router?

An **LLM router** selects which large language model should handle a request. Instead of hard-coding one model, a router can choose based on task type, capabilities, cost, latency, reliability, privacy or policy.

AI Best Model Route is an open-source LLM router that performs that selection locally and exposes the result through an OpenAI-compatible gateway.

### What is an AI model router?

An **AI model router** is infrastructure that decides which AI model should receive a request. ABMR supports task-aware model selection across configured providers and can distinguish requirements such as coding, vision, hosted tools, context size, privacy and cost.

### Is AI Best Model Route an AI gateway?

ABMR exposes an **OpenAI-compatible AI gateway interface**, but its primary purpose is model intelligence and automatic model selection. Bifrost provides the underlying provider execution plane.

### Does ABMR support OpenAI-compatible APIs?

Yes. ABMR exposes OpenAI-compatible model catalog, Responses and Chat Completions surfaces for supported configurations. It can be used with the OpenAI SDK and other software that accepts a custom OpenAI-compatible endpoint.

### Does ABMR support NVIDIA NIM and NVIDIA Build?

Yes. ABMR v1 includes a dedicated NVIDIA Build / NIM installer. It validates `NVIDIA_API_KEY`, queries NVIDIA's authenticated model catalog, configures the hosted endpoint, and dynamically exposes models available to that account.

### Can ABMR automatically choose the cheapest LLM?

Yes. `auto:cheap` biases toward lower-cost models after hard capability and policy requirements have been satisfied. Cost metadata must be configured from current provider pricing.

### Can ABMR choose the fastest LLM?

Yes. `auto:fast` weights latency more heavily while still enforcing capability and policy filters.

### Can ABMR route coding requests?

Yes. `auto:code` increases task affinity for models explicitly tagged/evaluated for software engineering.

### Can ABMR route only to local or private models?

Yes. Per-request policy can require local execution, and `auto:private` prefers local/private routes and stronger privacy tiers.

### Does ABMR support model failover?

ABMR tracks model-route health and uses circuit breaking to avoid repeatedly failing automatic routes. Bifrost handles the lower provider/key execution layer.

### Does ABMR work with LangChain and LangGraph?

Yes. Use an OpenAI-compatible client pointed at ABMR. LangChain and LangGraph can remain model/provider agnostic while ABMR chooses a route.

### Does ABMR work with OpenWebUI or LibreChat?

Yes, where the client supports a custom OpenAI-compatible endpoint.

### Is ABMR self-hosted?

Yes. The v1 project is local-first and ships Docker/GHCR deployment support. The default development setup binds the gateway to loopback.

### Is ABMR open source?

Yes. AI Best Model Route is released under the **Apache License 2.0** and preserves the required attribution of its upstream open-source foundations.

---

## 🌐 OYYO AI Intelligence spin-off

AI Best Model Route is a standalone open-source project and also part of the OYYO AI Intelligence architecture.

**OYYO principle:**

> One intelligence. One workspace. One place for everything.

OYYO is designed around the idea that a user should describe the goal while the system selects the appropriate model, agent, tool, memory, connector and execution path.

ABMR extracts the **model decision layer** from that philosophy into an independent project developers can use without an OYYO account.

### OYYO open-source ecosystem

| Project | Purpose |
| --- | --- |
| [OYYO Models](https://github.com/vpicciuolo/oyyo-models) | OYYO model assets and model-related work |
| [OYYO SDK](https://github.com/vpicciuolo/oyyo-sdk) | Developer-facing OYYO integration surface |
| [OYYO Benchmark](https://github.com/vpicciuolo/oyyo-benchmark) | Model and routing evaluation / benchmarking |
| [URL Intelligence Agent](https://github.com/vpicciuolo/url-intelligence-agent) | Open-source URL analysis and intelligence agent |

<p>
  <a href="https://github.com/vpicciuolo/oyyo-models"><img alt="OYYO Models GitHub repository" src="https://img.shields.io/badge/GitHub-OYYO%20Models-181717?style=flat-square&logo=github"></a>
  <a href="https://github.com/vpicciuolo/oyyo-sdk"><img alt="OYYO SDK GitHub repository" src="https://img.shields.io/badge/GitHub-OYYO%20SDK-181717?style=flat-square&logo=github"></a>
  <a href="https://github.com/vpicciuolo/oyyo-benchmark"><img alt="OYYO Benchmark GitHub repository" src="https://img.shields.io/badge/GitHub-OYYO%20Benchmark-181717?style=flat-square&logo=github"></a>
  <a href="https://github.com/vpicciuolo/url-intelligence-agent"><img alt="URL Intelligence Agent GitHub repository" src="https://img.shields.io/badge/GitHub-URL%20Intelligence%20Agent-181717?style=flat-square&logo=github"></a>
</p>

**OYYO AI Intelligence:** https://oyyo.one  
**HRN Innovation Technologies LTD:** https://hrn.ae  
**GitHub:** https://github.com/vpicciuolo

---

## Development

The decision engine is written in Go and is designed to remain deterministic and dependency-light.

```bash
nix develop
just test
just check-config
nix flake check -L
```

See **[CONTRIBUTING.md](CONTRIBUTING.md)** before submitting changes.

---

## Upstream and attribution

AI Best Model Route is a derivative work of:

- **Bifrost Model Router**, copyright 2026 Apply Innovations, Apache License 2.0.
- **Bifrost**, copyright 2025 H3 Labs Inc., Apache License 2.0.

The OYYO routing decision layer, automatic model-selection logic, NVIDIA integration, routing policy, explainability surfaces and subsequent modifications in this repository are maintained by HRN Innovation Technologies LTD / OYYO AI Intelligence.

See **[NOTICE](NOTICE)** for the complete attribution chain.

---

## License

Apache License 2.0. See **[LICENSE](LICENSE)**.

---

<p align="center">
  <strong>AI Best Model Route</strong><br/>
  <strong>One request. The right model. Automatically.</strong><br/><br/>
  If this project is useful, consider giving it a ⭐ so more developers can discover it.
</p>
