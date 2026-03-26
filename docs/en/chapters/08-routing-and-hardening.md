# Chapter 08: Routing and Hardening

This chapter is where the tutorial finally becomes a real CPA instead of a thin provider adapter. Once multiple upstream instances can serve the same model and the proxy routes between them predictably, the project starts adding value that a single direct SDK integration does not.

## What Problem This Chapter Solves

It shows how a CPA balances work across multiple configured upstreams for the same model and guards the downstream API surface with stable validation and upstream failure contracts.

## Why The Previous Chapter Is Not Enough

Chapter 7 proved that the runtime could translate and execute a Claude request, but it still behaved like a single-connection bridge. That is useful for smoke tests, not for a proxy layer. Multi-instance routing is the first real CPA value add because it lets one logical model fan out across multiple auths without changing the caller.

## New Concepts

- Round-robin selection per model
- Deterministic candidate ordering across multiple auths for the same model
- Normalized validation and upstream error responses
- Hardening tests that cover routing boundaries

## Implementation

- Start Tag: `chapter-07-claude-provider`
- End Tag: `chapter-08-routing-and-hardening`
- Implement deterministic per-model round-robin selector behavior so repeated requests alternate cleanly across eligible auths.
- Keep routing intentionally simple: no weighted routing, no persistence, and no cooldown scheduler. Disabled or cooldown auths are simply skipped.
- Improve downstream error handling so validation errors return `invalid_request_error` objects, upstream failures collapse into stable API errors, and JSON responses keep the correct default content type.

## Verification

`cd nanocpa && go test ./internal/auth ./internal/api -run 'TestManager_|TestServer_'` exercises round-robin selection, per-model isolation, unsupported model handling, and server-level routing boundaries.

`cd nanocpa && go test ./internal/api/handlers -run 'TestChatCompletions'` verifies the hardened downstream error contract and default JSON response semantics.

## What You Have Now

- A CPA that routes across multiple upstream auths per model with deterministic round-robin behavior.
- A downstream API surface that returns stable validation errors and normalized upstream failures.

## What Comes Next

The tutorial stops here for the first edition because this milestone captures the smallest complete CPA worth teaching: request translation, provider execution, per-model routing, and a hardened public boundary. Anything beyond this point starts pulling the tutorial toward platform engineering instead of a compact learning project.

Production features are intentionally deferred:

- Weighted or latency-aware routing
- Cooldown scheduling and active health management
- Persistent runtime state across restarts
- Provider-specific retry policies and circuit breakers
- Metrics, tracing, dashboards, and operational controls
- Streaming support, tool support, and broader provider coverage
