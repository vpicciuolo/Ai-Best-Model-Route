# Contributing

Contributions should preserve three architectural boundaries:

1. **Decision plane** — task, capability, policy, and quality selection in `internal/autoroute`.
2. **Execution plane** — Bifrost provider transport, credentials, compatibility, and provider/key routing.
3. **Client adapters** — Codex/OpenAI-compatible integration without leaking provider-specific behavior into decision logic.

Before a pull request:

```bash
nix develop
just test
just check-config
nix flake check -L
```

New routing strategies need tests showing both positive selection and hard-exclusion behavior. New capabilities must fail closed when their support cannot be verified.

Do not commit live provider credentials, captured authorization headers, real customer prompts, or production response data.

When changing code derived from the upstream Apache-2.0 project, preserve required copyright and NOTICE attribution and mark material modifications appropriately.
