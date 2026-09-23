# NVIDIA Build / NVIDIA API Catalog integration

AI Best Model Route includes a ready-to-use integration for the hosted NVIDIA API Catalog at **build.nvidia.com**.

The integration is deliberately account-aware. It does not freeze a static NVIDIA model list into this repository. Instead, it validates the user's NVIDIA API key against the live OpenAI-compatible model endpoint and lets Bifrost discover the models that key can actually access.

Official NVIDIA endpoint:

```text
https://integrate.api.nvidia.com/v1
```

Official model catalog:

```text
https://build.nvidia.com/models
```

## What you get

After setup:

- one local AI Best Model Route endpoint;
- your NVIDIA API key stored outside the repository in a mode-`0600` env file;
- dynamic discovery of NVIDIA models visible to that key;
- all discovered NVIDIA model IDs available through the router catalog;
- Codex access through the existing AI Best Model Route provider;
- conservative Responses-to-Chat adaptation for broad compatibility;
- direct manual model selection when you want a specific NVIDIA model;
- OYYO `auto:*` routing for models you explicitly approve as routing candidates;
- hosted-tool fallback to the configured OpenAI passthrough route when Codex asks a Chat-only route to execute a server-side OpenAI tool.

## Important distinction: available vs auto-routed

**Available** means the NVIDIA model is returned by the authenticated NVIDIA model catalog and can be selected directly.

**Auto-routed** means the model is also eligible for OYYO's best-model decision engine.

AI Best Model Route intentionally does **not** place every newly discovered model into automatic production routing. A new catalog entry may be an embedding model, a specialized model, a multimodal model with a different contract, or simply a model whose cost/quality/latency profile has not been evaluated.

This separation gives you immediate catalog access without silently granting new upstream models automatic traffic.

---

## Fast path: one command

From the repository root:

```bash
./scripts/setup-nvidia-build.sh
```

The script asks:

```text
Paste your NVIDIA Build API key (input is hidden):
```

Paste the key and press Enter.

The installer then:

1. calls the live NVIDIA `/v1/models` endpoint;
2. rejects an invalid key before touching the running router;
3. counts the model IDs returned for that account;
4. stores `NVIDIA_API_KEY` in:
   ```text
   ~/.config/ai-best-model-route/providers.env
   ```
5. sets the file to mode `0600`;
6. copies the ready NVIDIA configuration to:
   ```text
   ~/.config/ai-best-model-route/config.json
   ```
7. stores a private catalog snapshot at:
   ```text
   ~/.config/ai-best-model-route/nvidia-models.json
   ```
8. starts/replaces only the AI Best Model Route local containers;
9. installs the Codex provider;
10. sets new Codex tasks to the virtual `auto` route.

Then **fully quit and reopen Codex** and create a **new task**.

Existing Codex tasks retain the provider/session state they had when they were created.

---

## Non-interactive setup

If `NVIDIA_API_KEY` is already set in your terminal:

```bash
export NVIDIA_API_KEY="YOUR_KEY"
./scripts/setup-nvidia-build.sh
```

The installer still writes the key into the protected local provider env file because the Docker runtime needs it after the shell exits.

To validate and write configuration without starting Docker or modifying Codex:

```bash
./scripts/setup-nvidia-build.sh --configure-only
```

---

## Where to get the API key

1. Open:
   ```text
   https://build.nvidia.com
   ```
2. Sign in to your NVIDIA account.
3. Open a model/API page.
4. Use NVIDIA's **Generate API Key** flow.
5. Copy the generated key.
6. Run:
   ```bash
   ./scripts/setup-nvidia-build.sh
   ```
7. Paste the key only into the hidden terminal prompt.

Do not commit the key, paste it into an issue, add it to `config.json`, or put it in this repository.

---

## How the NVIDIA provider is configured

The ready template is:

```text
config/nvidia-build.json
```

At the Bifrost execution layer the provider uses:

```json
{
  "nvidia-build": {
    "keys": [
      {
        "name": "nvidia-build",
        "value": "env.NVIDIA_API_KEY",
        "models": ["*"],
        "weight": 1
      }
    ],
    "network_config": {
      "base_url": "https://integrate.api.nvidia.com"
    },
    "custom_provider_config": {
      "base_provider_type": "openai"
    }
  }
}
```

NVIDIA documents the public API root with `/v1`. Bifrost's OpenAI-compatible provider appends operation paths such as `/v1/models` and `/v1/chat/completions`, so its configured network base is intentionally the origin **without** a terminal `/v1`.

At the AI Best Model Route plugin layer:

```json
{
  "nvidia-build": {
    "display_name": "NVIDIA Build",
    "credential_mode": "bifrost",
    "responses_mode": "chat_polyfill",
    "adapter": "openai-chat",
    "discover_models": true
  }
}
```

The conservative Chat polyfill is intentional. NVIDIA's catalog contains models with different feature sets; the router should not advertise native Responses/tool/image capabilities for every discovered model unless they are verified.

---

## Verify NVIDIA discovery directly

After the key is installed locally:

```bash
set -a
. ~/.config/ai-best-model-route/providers.env
set +a

curl --fail --silent \
  https://integrate.api.nvidia.com/v1/models \
  -H "Authorization: Bearer $NVIDIA_API_KEY" \
  -H "Accept: application/json" |
  jq -r '.data[].id'
```

Be careful with `set -a` and shell history on shared machines. The one-command installer avoids printing the key.

---

## Verify through AI Best Model Route

Check the local router:

```bash
curl --fail http://127.0.0.1/health
```

The exact virtual key used by Codex is stored in the managed `ai-best-route` provider block in:

```text
~/.codex/config.toml
```

Do not paste that key into public logs.

A catalog request through the router should include NVIDIA Build entries after the provider has hydrated its authenticated model catalog.

In Codex, fully restart the application and create a new task. The model picker should expose the configured virtual routes and available provider models.

---

## Use a specific NVIDIA model

Once an NVIDIA model appears in the router catalog, use its canonical ABMR slug:

```text
nvidia-build/<nvidia-model-id>
```

For example, if NVIDIA returns:

```text
nvidia/nemotron-3.5-lightning-30b-a3b
```

the ABMR slug is:

```text
nvidia-build/nvidia/nemotron-3.5-lightning-30b-a3b
```

Do not assume a model in this example is enabled for every account forever. The authenticated catalog is the source of availability.

---

## Make an NVIDIA model eligible for auto routing

Discovery gives access. To let `auto`, `auto:quality`, `auto:fast`, etc. choose a model, add an explicit model entry under the plugin config after verifying its capabilities and operational metadata.

Example structure:

```yaml
models:
  nvidia-build/nvidia/example-model:
    provider: nvidia-build
    upstream_model: nvidia/example-model
    codex:
      context_window: 131072
      max_context_window: 131072
      input_modalities: [text]
    route:
      quality: 0.82
      input_cost_per_m: 0.50
      output_cost_per_m: 1.50
      latency_ms: 450
      reliability: 0.995
      strengths: [reasoning, code, tools]
      privacy_tier: 2
```

Those numbers are examples of the **shape**, not recommended values. Populate them from current NVIDIA pricing, your own measurements, and your own evaluation workload.

If you do not have verified data for a field, leave it unset instead of inventing it.

---

## Why every NVIDIA model is not automatically tagged as vision/tools/reasoning

The live NVIDIA catalog can contain:

- general chat LLMs;
- reasoning models;
- VLMs;
- embedding models;
- rerankers;
- task-specific inference models;
- models with different tool-calling behavior.

A plain OpenAI-style `/v1/models` response does not necessarily contain enough metadata to prove every capability.

AI Best Model Route therefore follows this rule:

> **availability can be discovered dynamically; capabilities used for hard routing decisions must be verified.**

This prevents a text-only route from silently receiving an image request or a Chat-only model from being presented as a native hosted-tool route.

---

## Existing AI Best Model Route installation

The one-key installer is optimized for a clean NVIDIA-first setup. If you already maintain a multi-provider configuration, do not overwrite it casually.

Recommended existing-install flow:

1. Open this repository in Codex.
2. Ask:
   ```text
   Add NVIDIA Build to my existing AI Best Model Route setup.
   Preserve every existing provider and routing candidate.
   ```
3. Codex follows `AGENTS.md`, merges the provider, preserves the existing env file and creates only the missing `NVIDIA_API_KEY=` placeholder.
4. Put the key into the local env file.
5. Ask Codex to verify discovery and apply the configuration.
6. Restart Codex and create a new task.

---

## Troubleshooting

### NVIDIA returns 401 or 403

Regenerate/check the API key in build.nvidia.com. Make sure you are using a key intended for the NVIDIA hosted API endpoint.

### The key is valid but a model is missing

The router intentionally follows what the authenticated NVIDIA catalog returns for your account. Model availability can differ over time or by account/entitlement.

### A model appears but Codex cannot use an image/tool feature

Catalog availability is not the same as verified Codex capability metadata. Add an explicit model profile only after confirming that model's protocol and features.

### Hosted web search fails on an NVIDIA Chat route

NVIDIA routes use the conservative Chat polyfill. OpenAI-hosted/server-side tools cannot be executed by an arbitrary Chat Completions provider. The NVIDIA template configures an OpenAI passthrough fallback for those requests when Codex OpenAI authentication is available.

### Docker setup succeeds but Codex still shows the old catalog

Fully quit Codex, reopen it, and create a **new task**. Existing tasks keep their original provider/session state.

### Port 80 is already occupied

The local quickstart publishes the router on loopback port 80. Stop the conflicting local service or adapt the local setup and Codex `base_url` together.

---

## Security checklist

- never commit `NVIDIA_API_KEY`;
- keep `providers.env` mode `0600`;
- do not log authorization headers;
- keep the local development router on loopback;
- use TLS and a real authentication boundary for any remote deployment;
- use virtual-key policy to limit models in multi-user environments;
- review models before adding them to automatic routing;
- treat current pricing/capability metadata as changeable operational data.
