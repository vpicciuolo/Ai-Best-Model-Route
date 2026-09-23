# Complete setup guide

This guide takes AI Best Model Route from a clean machine to a working multi-model router with Codex, OpenAI-compatible clients, optional NVIDIA Build, virtual `auto:*` routes, route preview, and secure provider credentials.

## 1. Understand the architecture first

AI Best Model Route has three layers:

```text
Client
  |
  |  OpenAI-compatible Responses / Chat
  v
AI Best Model Route
  |
  |  OYYO decision plane
  |  capability filters + policy + scoring + health
  v
Bifrost
  |
  |  provider execution / credentials / protocol adaptation
  v
Model providers
```

The OYYO routing layer decides **which configured model should run**.

Bifrost handles **how the request reaches that provider**.

Keeping those layers separate is important: semantic model choice should not be mixed with provider credential/key/load-balancing mechanics.

---

## 2. Prerequisites

For the local Docker quickstart:

- Linux or macOS;
- Docker with a running daemon;
- `curl`;
- `jq`;
- `openssl`;
- normal shell utilities;
- Codex Desktop or Codex CLI if you want the Codex integration.

For development/testing:

- Nix;
- Go version declared by `go.mod`;
- `just`.

Check the basics:

```bash
docker info
curl --version
jq --version
```

---

## 3. Clone and inspect

```bash
git clone https://github.com/vpicciuolo/Ai-Best-Model-Route.git
cd Ai-Best-Model-Route
```

Useful files:

```text
README.md
AGENTS.md
config/quickstart.json
config/nvidia-build.json
config/router.example.yaml
docs/routing-engine.md
docs/nvidia-build.md
scripts/setup-local.sh
scripts/setup-nvidia-build.sh
```

---

## 4. Choose the setup path

### Path A — easiest NVIDIA Build setup

Run:

```bash
./scripts/setup-nvidia-build.sh
```

Paste the NVIDIA API key when prompted.

Then fully restart Codex and create a new task.

See [NVIDIA Build integration](nvidia-build.md).

### Path B — agent-managed multi-provider setup

Open the repository in Codex and ask:

```text
Set AI Best Model Route up for me.
```

The repository's `AGENTS.md` instructs Codex to:

- inspect existing configuration;
- preserve current providers;
- determine which provider accounts/plans you want;
- verify provider endpoints and account-visible models;
- create credential placeholders outside the repository;
- avoid asking you to paste provider secrets into chat;
- generate/merge Bifrost configuration;
- verify the installed catalog with the exact local virtual key;
- keep the OYYO `auto` route as the default for new tasks.

### Path C — manual local configuration

Start from:

```text
config/quickstart.json
```

Copy it outside the repository:

```bash
mkdir -p ~/.config/ai-best-model-route
chmod 700 ~/.config/ai-best-model-route

cp config/quickstart.json ~/.config/ai-best-model-route/config.json
chmod 600 ~/.config/ai-best-model-route/config.json
```

Then add providers and environment credential references manually.

---

## 5. Provider credentials

Never store live provider keys directly in tracked JSON.

Use:

```text
~/.config/ai-best-model-route/providers.env
```

Example:

```dotenv
NVIDIA_API_KEY=
ANOTHER_PROVIDER_API_KEY=
```

Protect it:

```bash
chmod 600 ~/.config/ai-best-model-route/providers.env
```

Bifrost provider config should reference:

```json
"value": "env.NVIDIA_API_KEY"
```

not the literal key.

The setup executor validates that provider credentials in the JSON use `env.NAME` references.

---

## 6. Configure an OpenAI-compatible provider

For a provider exposing OpenAI-compatible endpoints, Bifrost can use a custom provider based on `openai`.

Conceptual structure:

```json
{
  "providers": {
    "my-provider": {
      "keys": [
        {
          "name": "my-provider",
          "value": "env.MY_PROVIDER_API_KEY",
          "models": ["*"],
          "weight": 1
        }
      ],
      "network_config": {
        "base_url": "https://provider.example.com"
      },
      "custom_provider_config": {
        "base_provider_type": "openai"
      }
    }
  }
}
```

Pay attention to the URL boundary. For standard OpenAI-compatible providers, Bifrost generally appends operation paths such as `/v1/models` and `/v1/chat/completions`. Do not accidentally configure a base that produces `/v1/v1/...`.

---

## 7. Tell AI Best Model Route about the provider

The Bifrost provider enables execution.

The plugin provider profile enables model resolution, discovery, Codex metadata, and Responses compatibility.

Example:

```yaml
providers:
  my-provider:
    display_name: My Provider
    credential_mode: bifrost
    responses_mode: chat_polyfill
    adapter: openai-chat
    discover_models: true
    codex_defaults:
      context_window: 131072
      input_modalities: [text]
```

Use `discover_models: true` only when the provider's model endpoint represents the models actually available to the supplied account/key.

If the provider model endpoint is only a global marketing catalog, use an explicit verified list instead.

---

## 8. Model discovery and canonical IDs

AI Best Model Route uses canonical slugs:

```text
provider/upstream-model-id
```

If an upstream NVIDIA model is:

```text
nvidia/example-model
```

the ABMR canonical slug is:

```text
nvidia-build/nvidia/example-model
```

Nested upstream IDs are preserved.

Dynamic discovery lets the provider catalog change without hard-coding every ID into application configuration.

---

## 9. Enable OYYO automatic routing

The default virtual routes are:

```text
auto
auto:quality
auto:fast
auto:cheap
auto:code
auto:private
```

The router first applies hard requirements, then scores the remaining candidates.

A model becomes a reviewed routing candidate through an explicit model entry:

```yaml
models:
  provider/model:
    provider: provider
    upstream_model: model
    codex:
      context_window: 131072
      input_modalities: [text]
    route:
      quality: 0.85
      input_cost_per_m: 1.00
      output_cost_per_m: 4.00
      latency_ms: 600
      reliability: 0.995
      strengths: [reasoning, code, tools]
      privacy_tier: 2
      local: false
```

### Routing metadata rules

Use:

- `quality`: normalized 0–1 from your evaluation workload;
- `input_cost_per_m`: current cost per million input tokens;
- `output_cost_per_m`: current cost per million output tokens;
- `latency_ms`: measured representative latency;
- `reliability`: normalized 0–1 operational baseline;
- `strengths`: verified/evaluated task affinities;
- `privacy_tier`: deployment policy signal from 0–3;
- `local`: true only for genuinely local/private execution.

Do not invent values to make the config look complete.

---

## 10. Configure routing profiles

Example:

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

The router normalizes weights internally.

Custom profile names can be added, for example:

```text
auto:research
auto:enterprise
auto:offline
auto:vision
```

---

## 11. Start the local router

For a generated configuration:

```bash
./scripts/setup-local.sh \
  --config ~/.config/ai-best-model-route/config.json \
  --env-file ~/.config/ai-best-model-route/providers.env \
  --model auto \
  --reasoning-effort medium \
  --accept-plaintext-key \
  --accept-new-threads-only
```

If the project's existing local containers need to be replaced:

```bash
./scripts/setup-local.sh \
  --config ~/.config/ai-best-model-route/config.json \
  --env-file ~/.config/ai-best-model-route/providers.env \
  --model auto \
  --reasoning-effort medium \
  --accept-plaintext-key \
  --accept-new-threads-only \
  --replace
```

The setup starts two local containers:

```text
ai-best-model-route-core
ai-best-model-route-gateway
```

and uses a persistent Docker volume:

```text
ai-best-model-route-data
```

---

## 12. Codex configuration

The setup executor manages a custom provider in:

```text
~/.codex/config.toml
```

Conceptually:

```toml
model = "auto"
model_provider = "ai-best-route"
model_reasoning_effort = "medium"

[model_providers.ai-best-route]
name = "AI Best Model Route"
base_url = "http://127.0.0.1/v1"
wire_api = "responses"
requires_openai_auth = true
http_headers = { "x-bf-vk" = "LOCAL_VIRTUAL_KEY" }
```

The exact virtual key is generated locally.

After configuration:

1. fully quit Codex;
2. reopen it;
3. create a **new** task.

Do not use an existing/resumed task to verify the new provider.

---

## 13. Verify health

```bash
docker ps --filter name=ai-best-model-route
curl --fail http://127.0.0.1/health
```

If health fails:

```bash
docker logs ai-best-model-route-core
docker logs ai-best-model-route-gateway
```

---

## 14. Preview an auto-routing decision

Before spending inference tokens:

```bash
curl http://127.0.0.1/v1/route/preview \
  -H "content-type: application/json" \
  -H "x-bf-vk: YOUR_LOCAL_VIRTUAL_KEY" \
  -d '{
    "model": "auto:code",
    "input": "Review this Go service and identify the concurrency bug."
  }'
```

The response explains:

- selected profile;
- detected signals;
- complexity;
- estimated input size;
- selected model;
- every eligible candidate and score;
- every excluded candidate and exclusion reason.

---

## 15. Add per-request policy

Example:

```json
{
  "model": "auto",
  "input": "Analyze this confidential codebase.",
  "ai_route": {
    "profile": "quality",
    "require_local": true,
    "min_quality": 0.80,
    "require_strengths": ["code"],
    "exclude_models": ["provider/model"]
  }
}
```

Supported constraints:

- `profile`
- `min_quality`
- `max_input_cost_per_m`
- `max_output_cost_per_m`
- `require_local`
- `require_strengths`
- `exclude_models`

The `ai_route` object is stripped before the request reaches the upstream model provider.

---

## 16. Use from the OpenAI Python SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1/v1",
    api_key="YOUR_LOCAL_VIRTUAL_KEY",
)

response = client.responses.create(
    model="auto",
    input="Design an API migration plan with rollback safety.",
)

print(response.output_text)
```

Use the authentication/header arrangement required by your deployment. The Codex quickstart additionally uses the Bifrost virtual-key header configured in Codex.

---

## 17. Use Chat Completions

```python
response = client.chat.completions.create(
    model="auto:fast",
    messages=[
        {"role": "user", "content": "Summarize this incident in five bullets."}
    ],
)
```

The concrete route is selected before Bifrost provider execution.

---

## 18. Hosted tools and Chat-only models

A Chat Completions provider cannot automatically execute OpenAI-hosted/server-side tools such as hosted web search.

For compatible Codex deployments, configure:

```yaml
hosted_tool_fallback_model: openai/<native-responses-model>
```

When a Chat-polyfilled request contains a hosted tool, the whole request moves to that native Responses model.

The router does not silently drop the tool just to keep the original model.

---

## 19. Security checklist

For local development:

- bind only to loopback;
- keep provider env files mode `0600`;
- never commit secrets;
- never put provider keys into Codex config;
- never use the Codex OpenAI bearer as a generic provider key;
- keep content logging disabled;
- rotate leaked virtual/provider keys.

For remote/production deployment:

- terminate TLS;
- add a proper authentication boundary;
- use managed secret storage;
- use virtual-key model policy;
- separate users/teams/projects;
- add rate limits and budgets;
- collect redacted operational telemetry;
- review every model before granting auto-routing eligibility.

---

## 20. Development and tests

```bash
nix develop
just test
just check-config
nix flake check -L
```

CI covers the inherited compatibility/security tests plus the OYYO routing engine tests.

When changing routing behavior, add tests for both:

- the model that should be selected;
- the model that must be excluded.

---

## 21. Upgrade

Pull the latest source:

```bash
git pull
```

Run the test suite, then rebuild/reinstall.

For local published images, the setup executor pulls the configured image tag before replacing project containers.

Preserve:

```text
~/.config/ai-best-model-route/providers.env
~/.config/ai-best-model-route/config.json
~/.codex/config.toml
```

The setup script creates a Codex config backup when it modifies an existing file.

---

## 22. Remove local containers

Stop:

```bash
docker stop ai-best-model-route-gateway ai-best-model-route-core
```

Remove containers:

```bash
docker rm -f ai-best-model-route-gateway ai-best-model-route-core
```

Remove the project network if desired:

```bash
docker network rm ai-best-model-route
```

The persistent volume is intentionally left in place.

Delete it only if you also want to discard local Bifrost state:

```bash
docker volume rm ai-best-model-route-data
```

---

## 23. Recommended first validation matrix

Before calling an installation ready, preview and execute representative requests:

| Test | Expected behavior |
| --- | --- |
| simple short text | balanced/fast-capable route |
| deep architecture analysis | stronger reasoning candidate |
| coding/refactor request | code-affinity candidate |
| image input | image-capable model only |
| hosted search/tool | native Responses route or configured fallback |
| long context | model with sufficient effective context |
| `require_local: true` | local model only |
| strict cost cap | expensive candidates excluded |
| repeated upstream 5xx | temporary circuit opens |
| unknown model | fail closed |

This matrix catches the most important routing mistakes before production traffic does.
