# Roadmap

## v0.1 — deterministic best-model routing

- OYYO `auto:*` virtual models
- capability and context hard filters
- quality/cost/latency/reliability/affinity/privacy scoring
- per-request policy constraints
- local health memory and circuit breaking
- Codex catalog injection
- route preview and explanations
- inherited Bifrost credential isolation and Responses compatibility
- reproducible unit tests

## v0.2 — measurable routing

- OpenTelemetry routing spans
- stable route decision IDs
- actual token/cost reconciliation after responses
- latency EWMA from real traffic
- per-model and provider SLO windows
- privacy-safe decision logs with content redaction
- benchmark/evaluation CLI
- replayable routing fixtures and golden decisions

## v0.3 — dynamic intelligence

- current pricing feeds
- refreshed context/capability metadata
- provider performance feeds
- data-residency policy
- shadow routing
- canary routing
- A/B model experiments
- budget windows per key, team, and project

## v0.4 — learned routing

- pluggable learned-router interface
- small local classifier
- embedding/similarity router
- pairwise preference router
- workload-specific offline calibration
- quality estimators and confidence thresholds
- deterministic fallback when confidence is low

## v0.5 — OYYO agent routing fabric

- agent/tool capability registry
- MCP-aware routing policy
- route based on tool availability as well as model capability
- separate planning and execution models
- memory-aware routing
- local/cloud hybrid policy
- OYYO orchestration hooks
- policy packs for enterprise deployments

## Design rules

1. Hard policy always wins over learned preference.
2. Routing decisions must be inspectable.
3. Unknown required capability fails closed.
4. Provider credentials remain isolated.
5. A route decision should not require sending the prompt to a second remote model by default.
6. Every optimization must be measurable against a deterministic baseline.
7. Semantic model choice and provider/key choice remain separate layers.
8. Dynamic discovery never automatically grants a new model production-routing eligibility.
