# Routing engine

AI Best Model Route separates the **decision plane** from the **execution plane**.

The OYYO-derived decision plane chooses a concrete model. Bifrost remains responsible for provider transport, credentials, protocol adaptation, key selection, and provider-level routing.

## Evaluation order

1. Parse the virtual model/profile.
2. Consume local-only `ai_route` constraints.
3. Inspect the request locally.
4. Derive hard requirements.
5. Remove ineligible models.
6. Apply temporary circuit state.
7. Score remaining candidates.
8. Deterministically choose the highest score.
9. Rewrite `model` to the concrete canonical slug.
10. Run normal credential isolation and Bifrost execution.

Hard filters happen before soft scoring. A low-cost model cannot win if it lacks an input modality, context window, hosted-tool contract, local-only requirement, required strength, or policy minimum.

## Current signals

The v0.1 classifier is intentionally local and deterministic. It detects request characteristics associated with coding, reasoning, tools, vision, long context, and lightweight conversational work. It does not make a second LLM request merely to choose the first LLM.

This provides three useful properties:

- no routing-token bill;
- no extra prompt disclosure to a separate routing service;
- reproducible behavior that can be unit-tested.

This heuristic classifier is a safe baseline, not the final word in learned routing.

## Scoring

Profiles combine six normalized terms:

```text
score =
  quality_weight     * quality
+ cost_weight        * cost_efficiency
+ latency_weight     * latency_efficiency
+ reliability_weight * observed_reliability
+ affinity_weight    * task_affinity
+ privacy_weight     * privacy
```

Weights are normalized at evaluation time.

Routing metadata belongs to deployment configuration because price, latency, and measured quality change. Do not bake provider marketing claims into the engine.

## Candidate scope

Provider discovery and automatic routing are deliberately separate.

A discovered provider model can appear in the normal catalog without automatically becoming an `auto` candidate. Automatic routing considers explicit model entries that carry reviewed routing metadata. This prevents a newly announced model from entering production traffic solely because it appeared in an upstream catalog.

## Health and circuit breaking

HTTP 429 and 5xx results count as route-health failures. Three consecutive failures temporarily open the local circuit for 30 seconds. Successful responses reset the failure streak and update an EWMA-style success signal.

Other client 4xx errors do not degrade model health because they normally describe the request rather than the upstream service.

This circuit is intentionally small and complementary to Bifrost provider/key health logic.

## Per-request policy

`ai_route` is consumed by ABMR and removed before forwarding.

Supported fields:

- `profile`
- `min_quality`
- `max_input_cost_per_m`
- `max_output_cost_per_m`
- `require_local`
- `require_strengths`
- `exclude_models`

An impossible policy fails closed with `route_unavailable`.

## Explainability

`POST /v1/route/preview` accepts the same JSON shape as inference and returns the routing decision without running inference.

The result includes:

- selected profile;
- task classification and signals;
- estimated input tokens;
- complexity;
- selected concrete model;
- every candidate considered;
- score components for eligible candidates;
- exclusion reasons for ineligible candidates.

Use preview in CI, regression tests, configuration reviews, cost-policy tests, and routing-debug workflows.

## Learned-routing roadmap

A learned router should be additive, not a replacement for hard capability and governance filters.

The intended future pipeline is:

```text
hard policy/capability filters
  -> deterministic features
  -> optional learned quality predictor
  -> online health/cost/latency signals
  -> constrained selection
```

Possible learned strategies include pairwise preference routers, small local classifiers, embedding/similarity routing, and task-specific evaluators. Every learned strategy should be benchmarked against this deterministic baseline and retain a usable explanation surface.
