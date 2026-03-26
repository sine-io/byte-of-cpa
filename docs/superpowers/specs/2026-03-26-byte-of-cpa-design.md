# Byte Of CPA Design

## Overview

`byte-of-cpa` is a tutorial-first repository for programmers who want to build their own CPA.
The goal is not to recreate the full `CLIProxyAPI` feature set. The goal is to teach the minimum architecture that makes a CPA real: protocol compatibility, config-driven upstream definition, routing, translation, and testability.

The tutorial will use the existing [`nanocpa`](../../../../nanocpa/README.md) code as the reference answer base, but the teaching flow will be organized as a progressive build, not as a code dump.

## Audience

The target reader already knows how to program and can read Go code.
They do not need to be taught basic syntax, HTTP, or JSON from first principles.
They do need to understand why a CPA needs specific layers, where those layers begin and end, and how the layers fit together.

## Problem Statement

The full `CLIProxyAPI` project is a broad production proxy system. It supports multiple downstream compatibility surfaces, multiple upstream providers, OAuth flows, management APIs, more advanced routing, and many operational features.

That scope is useful as a reference, but it is too large for a first-principles tutorial. If `byte-of-cpa` starts from the full surface area, readers will copy code without understanding why the code exists.

The tutorial therefore needs a constrained target:

- small enough to finish
- real enough to feel like a CPA instead of a toy relay
- structured enough that each layer can be explained in isolation

## Design Goals

- Teach readers how to build a minimal but real CPA.
- Prefer a thin-slice progression where every chapter adds one meaningful capability.
- Keep the first downstream protocol to OpenAI-compatible HTTP.
- Build a provider-agnostic runtime skeleton before introducing the first concrete provider.
- Use Claude as the first concrete provider because it makes protocol translation explicit.
- Keep the tutorial codebase small, readable, and testable.
- Let readers check out chapter snapshots with git tags and follow along incrementally.

## Non-Goals

The first edition of the tutorial will not cover:

- OAuth login flows
- remote management APIs
- control panel UI
- WebSocket APIs
- Responses API
- streaming SSE
- tool calling
- multimodal payloads
- multiple downstream protocol surfaces
- complex retry, quota cooling, and fallback systems
- production operations topics such as persistence, metrics, log pipelines, or hot reload

These are valid later topics, but they are outside the first teaching loop.

## Recommended Product Shape

The tutorial endpoint is a minimal CPA with these properties:

- downstream protocol: OpenAI-compatible HTTP
- concrete endpoints: `POST /v1/chat/completions` and `GET /v1/models`
- downstream access control: Bearer API key
- upstream definition: YAML config
- runtime shape: generic manager plus provider executors
- first provider: Claude
- model visibility: driven by configured provider/model registrations
- routing strategy: round-robin for multiple matching upstreams
- error contract: stable OpenAI-style JSON error responses
- verification: unit tests around config, translator, handler, routing, and server behavior

This shape is large enough to teach the core mechanics of a CPA and small enough to remain teachable.

## Minimum Capabilities

The tutorial project should, at minimum, teach and produce these capabilities:

1. Start an HTTP server.
2. Load and validate YAML config.
3. Protect downstream routes with Bearer API key authentication.
4. Expose `POST /v1/chat/completions`.
5. Expose `GET /v1/models`.
6. Register model availability from config and reject unsupported models.
7. Organize runtime execution around provider-agnostic interfaces.
8. Translate OpenAI chat requests to Claude message requests and Claude responses back to OpenAI chat responses.
9. Route across multiple upstream instances of the same provider with round-robin selection.

If the tutorial produces less than this, the result is too close to a single-upstream relay.
If it tries to do much more than this in the first edition, the tutorial loses focus.

## Why This Scope Maps To CLIProxyAPI

The complete `CLIProxyAPI` project demonstrates that a CPA is more than raw request forwarding. Its useful core ideas are:

- downstream compatibility matters
- upstreams should be configured rather than hard-coded
- models are part of routing, not just request text
- provider-specific translation belongs behind stable interfaces
- multiple upstream credentials and simple routing add real value

`byte-of-cpa` should therefore teach the smallest architecture that still reflects those ideas.
The tutorial should not mirror production breadth. It should mirror production shape.

## Tutorial Approach

The tutorial will use a thin-slice progression.

Each chapter introduces one new capability layer, explains why the previous stage is insufficient, and leaves the repository in a runnable state.
This is preferred over a bottom-up build because readers should see visible results early.
It is also preferred over a one-file demo because the tutorial is trying to teach system boundaries, not just prove that a request can be proxied.

## Repository Structure

The repository will be organized around two things:

- the working tutorial code in [`nanocpa/`](../../../../nanocpa/)
- tutorial documentation and planning documents in `docs/`

`main` will represent the latest complete tutorial state.
Each chapter will end with a git tag.
The tutorial will not use per-chapter code copies and will not maintain long-lived chapter branches.

## Git Tag Strategy

Each chapter ends with a stable tag so readers can move chapter by chapter without duplicating source trees.

Recommended tag names:

- `chapter-01-bootstrap`
- `chapter-02-config`
- `chapter-03-access`
- `chapter-04-openai-surface`
- `chapter-05-model-registry`
- `chapter-06-runtime-skeleton`
- `chapter-07-claude-provider`
- `chapter-08-routing-and-hardening`

Each chapter document should clearly state:

- start tag
- end tag
- what changed in this chapter
- how to verify the chapter result

## Chapter Map

### Chapter 1: Bootstrap

**Teaching goal:** show the smallest runnable service boundary.

**New capability:**
- `main`
- server construction
- listen address and safe HTTP timeouts

**Reader learns:**
- a CPA is first a network service
- even the smallest service should have explicit construction boundaries

**Result:** a service starts successfully, even if it has no useful API yet.

### Chapter 2: Config

**Teaching goal:** explain why upstreams and access rules belong in config, not in hard-coded constants.

**New capability:**
- config struct
- YAML loading
- normalization
- validation

**Reader learns:**
- configuration is part of the architecture
- validation makes the rest of the code simpler and safer

**Result:** the service becomes configurable and rejects invalid startup state early.

### Chapter 3: Access

**Teaching goal:** show that a CPA needs its own downstream authentication layer.

**New capability:**
- Bearer API key middleware

**Reader learns:**
- downstream access control is distinct from upstream provider credentials
- middleware is the right place for cross-cutting request policy

**Result:** only authorized callers can use the proxy.

### Chapter 4: OpenAI Surface

**Teaching goal:** establish the downstream contract before connecting to real upstream behavior.

**New capability:**
- `POST /v1/chat/completions`
- `GET /v1/models`
- stable JSON error shape

**Reader learns:**
- compatibility begins at the boundary
- handler-level validation should happen before runtime dispatch

**Result:** the service now looks like an OpenAI-compatible API surface.

### Chapter 5: Model Registry

**Teaching goal:** explain why models are a routing concern.

**New capability:**
- model registration from config
- model listing
- model support checks

**Reader learns:**
- model names define which upstreams are eligible
- `/v1/models` should describe real availability, not placeholders

**Result:** model discovery and validation become data-driven.

### Chapter 6: Runtime Skeleton

**Teaching goal:** create provider-agnostic execution boundaries before adding provider logic.

**New capability:**
- runtime auth records
- manager
- executor interface
- selector interface

**Reader learns:**
- provider-specific code should sit behind stable execution interfaces
- clean runtime boundaries make future providers possible

**Result:** the system is now structured as a generic CPA skeleton instead of endpoint-specific glue.

### Chapter 7: Claude Provider

**Teaching goal:** make protocol translation concrete.

**New capability:**
- OpenAI request translation to Claude
- Claude upstream executor
- Claude response translation back to OpenAI

**Reader learns:**
- the core of a CPA is translation, not blind forwarding
- provider adapters should remain narrow and testable

**Result:** the service can complete an end-to-end OpenAI-compatible request through Claude.

### Chapter 8: Routing And Hardening

**Teaching goal:** move from a single-upstream demo to a minimal real CPA.

**New capability:**
- multiple configured upstream instances
- round-robin selection
- stronger tests around routing and server safety
- clearer upstream/downstream error handling

**Reader learns:**
- one concrete reason to build a CPA is routing across multiple upstream instances
- stable behavior comes from explicit tests, not assumptions

**Result:** the tutorial reaches its first complete milestone: a minimal, explainable, multi-instance CPA.

## Documentation Structure

Each chapter document should use the same teaching template:

1. What problem this chapter solves.
2. Why the previous chapter is insufficient.
3. The new concepts introduced in this chapter.
4. The concrete implementation work.
5. How to run or test the result.
6. What capabilities the system has at the end of the chapter.
7. What limitation remains and why the next chapter exists.

This repetition is intentional. It keeps the tutorial focused on causality instead of only on code changes.

## Reference Code Relationship

The existing `nanocpa` code should be treated as the reference answer set for the first edition.
It already contains the main architectural pieces the tutorial wants to teach:

- bootstrap entrypoint
- config loader and validator
- API key middleware
- OpenAI-compatible handlers
- model registry
- runtime manager and selector
- Claude executor
- OpenAI/Claude translator
- tests around the most important behavior

The tutorial should not simply point readers at those files.
Instead, it should reorganize the learning flow so those files become the destination of a sequence of understandable steps.

## Success Criteria

The design is successful if a reader can finish the tutorial and confidently explain:

- why a CPA needs a downstream compatibility surface
- why config validation comes early
- why downstream auth and upstream auth are different concerns
- why model registration belongs outside the handler
- why a manager/executor boundary exists
- why a provider adapter is mainly about translation
- why multiple upstream instances need selection logic

The design is not successful if the reader only ends up with working code and cannot explain the architecture.

## Implementation Constraint

The tutorial will optimize for clarity over breadth.
Any new feature proposed during implementation should be judged against one question:

Does this make the first-edition tutorial clearer, or does it just make the project more production-like?

If it only increases production resemblance, it should be deferred.
