# Security policy

AI Best Model Route sits between clients and model providers and may handle credentials, prompts, tool declarations, and model-selection policy.

Report security issues privately rather than opening a public issue containing exploit details, customer data, or secrets.

## Security invariants

- A Codex/OpenAI bearer is request-scoped OpenAI material, not a generic provider credential.
- Managed-provider routes strip caller `Authorization`, `x-api-key`, `x-goog-api-key`, and caller-supplied direct-key headers.
- Provider credentials belong in runtime environment/secret stores, never tracked configuration.
- `ai_route` is consumed locally and removed before provider forwarding.
- Unknown or ambiguous models and impossible routing policies fail closed.
- The supplied local quickstart binds to loopback.
- Prompt/response content logging is disabled by default in the supplied Bifrost configuration.
- Route preview performs no inference and must never expose provider secrets.
- A model being discoverable does not automatically make it eligible for `auto` routing.

See `docs/security.md` for the inherited credential and deployment model.
