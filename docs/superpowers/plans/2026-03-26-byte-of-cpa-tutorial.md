# Byte Of CPA Tutorial Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the current `nanocpa` codebase into a chapter-driven tutorial repository with runnable chapter milestones, stable chapter tags, and documentation that teaches both implementation and rationale.

**Architecture:** Keep `nanocpa/` as the working codebase and rebuild the tutorial as a linear sequence of chapter commits on `main`, using the current code as the chapter-08 reference answer. Add a docs layer that explains each chapter’s problem, implementation, verification, and architectural purpose, then tag the end of each chapter so readers can move through the tutorial incrementally.

**Tech Stack:** Go 1.26, standard library `net/http`, YAML config, Markdown docs, git tags, existing `nanocpa` tests under `nanocpa/internal/...`.

---

## Implementation Notes

- Use the current `HEAD` commit as the source of truth for the desired chapter-08 end state.
- Implement this in a dedicated worktree or branch so the chapter-history rewrite does not block unrelated work.
- Do not create per-chapter source copies. The tutorial lives in one code tree plus chapter tags.
- Keep `main` at the latest complete tutorial state.
- Use existing `nanocpa` tests as the starting verification suite, then reshape or add smaller tests where earlier chapters need narrower expectations.

## Planned File Map

**Create**
- `README.md`
- `docs/chapters/README.md`
- `docs/chapters/01-bootstrap.md`
- `docs/chapters/02-config.md`
- `docs/chapters/03-access.md`
- `docs/chapters/04-openai-surface.md`
- `docs/chapters/05-model-registry.md`
- `docs/chapters/06-runtime-skeleton.md`
- `docs/chapters/07-claude-provider.md`
- `docs/chapters/08-routing-and-hardening.md`

**Modify**
- `nanocpa/README.md`
- `nanocpa/cmd/server/main.go`
- `nanocpa/config.example.yaml`
- `nanocpa/internal/access/apikey.go`
- `nanocpa/internal/access/apikey_test.go`
- `nanocpa/internal/api/middleware.go`
- `nanocpa/internal/api/server.go`
- `nanocpa/internal/api/server_test.go`
- `nanocpa/internal/api/handlers/openai.go`
- `nanocpa/internal/api/handlers/openai_test.go`
- `nanocpa/internal/auth/manager.go`
- `nanocpa/internal/auth/manager_test.go`
- `nanocpa/internal/auth/selector.go`
- `nanocpa/internal/auth/types.go`
- `nanocpa/internal/config/config.go`
- `nanocpa/internal/config/config_test.go`
- `nanocpa/internal/registry/model_registry.go`
- `nanocpa/internal/registry/model_registry_test.go`
- `nanocpa/internal/runtime/executor/interface.go`
- `nanocpa/internal/runtime/executor/claude.go`
- `nanocpa/internal/runtime/executor/claude_test.go`
- `nanocpa/internal/translator/openai_claude.go`
- `nanocpa/internal/translator/openai_claude_test.go`

**Reference Only**
- `docs/superpowers/specs/2026-03-26-byte-of-cpa-design.md`

## Task 1: Add Tutorial Entry Points And Docs Skeleton

**Files:**
- Create: `README.md`
- Create: `docs/chapters/README.md`
- Create: `docs/chapters/01-bootstrap.md`
- Create: `docs/chapters/02-config.md`
- Create: `docs/chapters/03-access.md`
- Create: `docs/chapters/04-openai-surface.md`
- Create: `docs/chapters/05-model-registry.md`
- Create: `docs/chapters/06-runtime-skeleton.md`
- Create: `docs/chapters/07-claude-provider.md`
- Create: `docs/chapters/08-routing-and-hardening.md`
- Modify: `nanocpa/README.md`

- [ ] **Step 1: Draft the root tutorial index**

Write `README.md` with:
- tutorial purpose
- audience
- repository layout
- how chapter tags work
- link list for chapters 1-8

- [ ] **Step 2: Draft the chapter index**

Write `docs/chapters/README.md` with:
- chapter order
- start/end tag convention
- how readers should move between chapter tags
- verification expectations for each chapter

- [ ] **Step 3: Create the eight chapter markdown shells**

Each file must contain these headings:

```md
# Chapter N: <Title>

## What Problem This Chapter Solves
## Why The Previous Chapter Is Not Enough
## New Concepts
## Implementation
## Verification
## What You Have Now
## What Comes Next
```

- [ ] **Step 4: Align `nanocpa/README.md` with tutorial framing**

Update `nanocpa/README.md` so it clearly says:
- this directory holds the working tutorial code
- the guided walkthrough lives in `docs/chapters/`
- chapter tags are the teaching milestones

- [ ] **Step 5: Verify the docs skeleton exists**

Run: `find README.md docs/chapters nanocpa/README.md -maxdepth 2 -type f | sort`
Expected: root README, chapter index, eight chapter docs, and `nanocpa/README.md`

- [ ] **Step 6: Commit**

```bash
git add README.md docs/chapters nanocpa/README.md
git commit -m "docs: add tutorial entrypoints and chapter skeletons"
```

## Task 2: Rebuild Chapter 1 Bootstrap Snapshot

**Files:**
- Modify: `nanocpa/cmd/server/main.go`
- Modify: `nanocpa/internal/api/server.go`
- Modify: `nanocpa/internal/api/server_test.go`
- Modify: `docs/chapters/01-bootstrap.md`

- [ ] **Step 1: Narrow the code to a chapter-01 bootstrap surface**

Reduce the runtime to the smallest runnable server:
- `main` loads a simple config path and starts the server
- `server.go` only needs enough to construct and run an HTTP server
- do not include auth middleware, model registry wiring, or provider execution yet

- [ ] **Step 2: Write or reshape the bootstrap tests**

Ensure `nanocpa/internal/api/server_test.go` contains tests for:
- safe server timeouts
- correctly formatted listen address
- basic server construction failure on nil config if that contract remains

- [ ] **Step 3: Run the bootstrap-focused tests**

Run: `cd nanocpa && go test ./internal/api -run 'TestServer_'`
Expected: PASS with only bootstrap server tests running

- [ ] **Step 4: Write `docs/chapters/01-bootstrap.md`**

Document:
- why a CPA starts as a network service
- why chapter 1 intentionally does not proxy anything yet
- which files were introduced
- how to run the service

- [ ] **Step 5: Tag the milestone**

```bash
git add nanocpa/cmd/server/main.go nanocpa/internal/api/server.go nanocpa/internal/api/server_test.go docs/chapters/01-bootstrap.md
git commit -m "feat: add chapter 01 bootstrap milestone"
git tag chapter-01-bootstrap
```

## Task 3: Add Chapter 2 Config Snapshot

**Files:**
- Modify: `nanocpa/config.example.yaml`
- Modify: `nanocpa/internal/config/config.go`
- Modify: `nanocpa/internal/config/config_test.go`
- Modify: `docs/chapters/02-config.md`

- [ ] **Step 1: Write the failing config validation tests**

Cover:
- missing host
- invalid port
- empty `api_keys`
- missing provider fields
- unsupported provider values

Run: `cd nanocpa && go test ./internal/config -run 'TestConfig|TestLoad|TestValidate'`
Expected: FAIL for any validations not yet implemented

- [ ] **Step 2: Implement normalization and validation**

Keep the config limited to first-edition needs:
- `host`
- `port`
- `api_keys`
- `providers[].id`
- `providers[].provider`
- `providers[].api_key`
- `providers[].base_url`
- `providers[].models`

- [ ] **Step 3: Re-run the config test suite**

Run: `cd nanocpa && go test ./internal/config`
Expected: PASS

- [ ] **Step 4: Update the sample config and write chapter docs**

Explain:
- why config is the system boundary
- why invalid config must fail early
- why provider declarations are data, not code

- [ ] **Step 5: Commit and tag**

```bash
git add nanocpa/config.example.yaml nanocpa/internal/config/config.go nanocpa/internal/config/config_test.go docs/chapters/02-config.md
git commit -m "feat: add chapter 02 config milestone"
git tag chapter-02-config
```

## Task 4: Add Chapter 3 Access Snapshot

**Files:**
- Modify: `nanocpa/internal/access/apikey.go`
- Modify: `nanocpa/internal/access/apikey_test.go`
- Modify: `nanocpa/internal/api/middleware.go`
- Modify: `docs/chapters/03-access.md`

- [ ] **Step 1: Write the failing API key tests**

Cover:
- valid Bearer token accepted
- wrong scheme rejected
- missing token rejected
- wrong token rejected

Run: `cd nanocpa && go test ./internal/access`
Expected: FAIL if header parsing does not yet meet chapter requirements

- [ ] **Step 2: Implement Bearer key validation and middleware**

Keep it minimal:
- exact key matching
- `401` for unauthorized requests
- stable JSON error body

- [ ] **Step 3: Add or update middleware integration tests**

Run: `cd nanocpa && go test ./internal/api -run 'Test.*Unauthorized|Test.*Middleware'`
Expected: PASS

- [ ] **Step 4: Write `docs/chapters/03-access.md`**

Explain:
- downstream auth versus upstream provider auth
- why middleware is introduced here instead of inside handlers

- [ ] **Step 5: Commit and tag**

```bash
git add nanocpa/internal/access/apikey.go nanocpa/internal/access/apikey_test.go nanocpa/internal/api/middleware.go docs/chapters/03-access.md
git commit -m "feat: add chapter 03 access milestone"
git tag chapter-03-access
```

## Task 5: Add Chapter 4 OpenAI Surface Snapshot

**Files:**
- Modify: `nanocpa/internal/api/server.go`
- Modify: `nanocpa/internal/api/handlers/openai.go`
- Modify: `nanocpa/internal/api/handlers/openai_test.go`
- Modify: `docs/chapters/04-openai-surface.md`

- [ ] **Step 1: Write failing handler tests for the downstream surface**

Cover:
- `POST /v1/chat/completions` exists
- invalid JSON returns OpenAI-style error
- missing `model` returns validation error
- `GET /v1/models` returns a JSON list shape

Run: `cd nanocpa && go test ./internal/api/... -run 'TestOpenAI|TestChatCompletions|TestModels'`
Expected: FAIL for missing handler behavior

- [ ] **Step 2: Implement the handler surface without full provider execution**

Introduce:
- route registration
- request body size limit
- request validation
- stable error contract

- [ ] **Step 3: Re-run handler tests**

Run: `cd nanocpa && go test ./internal/api/...`
Expected: PASS for handler and server packages touched so far

- [ ] **Step 4: Write `docs/chapters/04-openai-surface.md`**

Explain:
- why the tutorial establishes the downstream contract before upstream integration
- why compatibility starts at the API boundary

- [ ] **Step 5: Commit and tag**

```bash
git add nanocpa/internal/api/server.go nanocpa/internal/api/handlers/openai.go nanocpa/internal/api/handlers/openai_test.go docs/chapters/04-openai-surface.md
git commit -m "feat: add chapter 04 openai surface milestone"
git tag chapter-04-openai-surface
```

## Task 6: Add Chapter 5 Model Registry Snapshot

**Files:**
- Modify: `nanocpa/internal/registry/model_registry.go`
- Modify: `nanocpa/internal/registry/model_registry_test.go`
- Modify: `nanocpa/internal/api/handlers/openai.go`
- Modify: `nanocpa/internal/api/handlers/openai_test.go`
- Modify: `docs/chapters/05-model-registry.md`

- [ ] **Step 1: Write failing registry tests**

Cover:
- client model registration
- list models
- find providers for a model
- client supports model
- unregister behavior if still needed in the final interface

Run: `cd nanocpa && go test ./internal/registry`
Expected: FAIL until registry behavior matches chapter scope

- [ ] **Step 2: Wire model checks into the OpenAI handlers**

Add:
- `/v1/models` listing from registry
- unsupported model rejection in chat completions

- [ ] **Step 3: Run registry and handler tests**

Run: `cd nanocpa && go test ./internal/registry ./internal/api/...`
Expected: PASS

- [ ] **Step 4: Write `docs/chapters/05-model-registry.md`**

Explain:
- why model names are routing inputs
- why `/v1/models` should reflect real configured availability

- [ ] **Step 5: Commit and tag**

```bash
git add nanocpa/internal/registry/model_registry.go nanocpa/internal/registry/model_registry_test.go nanocpa/internal/api/handlers/openai.go nanocpa/internal/api/handlers/openai_test.go docs/chapters/05-model-registry.md
git commit -m "feat: add chapter 05 model registry milestone"
git tag chapter-05-model-registry
```

## Task 7: Add Chapter 6 Runtime Skeleton Snapshot

**Files:**
- Modify: `nanocpa/internal/auth/types.go`
- Modify: `nanocpa/internal/auth/manager.go`
- Modify: `nanocpa/internal/auth/manager_test.go`
- Modify: `nanocpa/internal/auth/selector.go`
- Modify: `nanocpa/internal/api/server.go`
- Modify: `docs/chapters/06-runtime-skeleton.md`

- [ ] **Step 1: Write failing manager tests**

Cover:
- runtime auth registration
- executor registration
- candidate selection by model
- no executor for provider
- disabled or cooldown auths are skipped if that contract is already part of the intended final API

Run: `cd nanocpa && go test ./internal/auth`
Expected: FAIL until manager orchestration exists

- [ ] **Step 2: Implement provider-agnostic runtime boundaries**

Introduce or stabilize:
- `Auth`
- `Result`
- `Executor`
- `Manager`
- `Selector`

Keep the runtime generic. Do not hard-wire Claude logic into the manager.

- [ ] **Step 3: Re-run runtime tests**

Run: `cd nanocpa && go test ./internal/auth ./internal/registry`
Expected: PASS

- [ ] **Step 4: Update server wiring to use the generic manager**

`server.go` should construct:
- runtime auths from config
- model registry
- manager

without yet fully explaining provider translation in this chapter doc.

- [ ] **Step 5: Write `docs/chapters/06-runtime-skeleton.md`**

Explain:
- why the tutorial introduces abstraction here
- why this is not over-engineering
- why provider adapters belong behind stable interfaces

- [ ] **Step 6: Commit and tag**

```bash
git add nanocpa/internal/auth/types.go nanocpa/internal/auth/manager.go nanocpa/internal/auth/manager_test.go nanocpa/internal/auth/selector.go nanocpa/internal/api/server.go docs/chapters/06-runtime-skeleton.md
git commit -m "feat: add chapter 06 runtime skeleton milestone"
git tag chapter-06-runtime-skeleton
```

## Task 8: Add Chapter 7 Claude Provider Snapshot

**Files:**
- Modify: `nanocpa/internal/runtime/executor/interface.go`
- Modify: `nanocpa/internal/runtime/executor/claude.go`
- Modify: `nanocpa/internal/runtime/executor/claude_test.go`
- Modify: `nanocpa/internal/translator/openai_claude.go`
- Modify: `nanocpa/internal/translator/openai_claude_test.go`
- Modify: `nanocpa/internal/api/server.go`
- Modify: `docs/chapters/07-claude-provider.md`

- [ ] **Step 1: Write failing translator tests**

Cover:
- valid OpenAI chat request becomes Claude messages request
- system messages become `system`
- unsupported roles fail
- unsupported content shape fails
- Claude response becomes OpenAI chat response

Run: `cd nanocpa && go test ./internal/translator`
Expected: FAIL until translation logic is correct

- [ ] **Step 2: Implement the request/response translator**

Keep the first-edition translator intentionally narrow:
- string-only message content
- single completion shape
- no streaming
- no tool calling

- [ ] **Step 3: Write failing Claude executor tests**

Cover:
- missing runtime auth errors
- missing `base_url` or `api_key` errors
- successful upstream request translation
- upstream non-2xx becomes wrapped error

Run: `cd nanocpa && go test ./internal/runtime/executor`
Expected: FAIL until executor behavior is in place

- [ ] **Step 4: Implement the Claude executor and register it from the server**

Use:
- `POST /v1/messages`
- `x-api-key`
- `anthropic-version: 2023-06-01`

- [ ] **Step 5: Run the provider path test suite**

Run: `cd nanocpa && go test ./internal/translator ./internal/runtime/executor ./internal/api/...`
Expected: PASS

- [ ] **Step 6: Write `docs/chapters/07-claude-provider.md`**

Explain:
- why the heart of a CPA is translation
- why provider-specific logic is intentionally narrow
- why the manager stays generic while the provider stays concrete

- [ ] **Step 7: Commit and tag**

```bash
git add nanocpa/internal/runtime/executor/interface.go nanocpa/internal/runtime/executor/claude.go nanocpa/internal/runtime/executor/claude_test.go nanocpa/internal/translator/openai_claude.go nanocpa/internal/translator/openai_claude_test.go nanocpa/internal/api/server.go docs/chapters/07-claude-provider.md
git commit -m "feat: add chapter 07 claude provider milestone"
git tag chapter-07-claude-provider
```

## Task 9: Add Chapter 8 Routing And Hardening Snapshot

**Files:**
- Modify: `nanocpa/internal/auth/selector.go`
- Modify: `nanocpa/internal/auth/manager.go`
- Modify: `nanocpa/internal/auth/manager_test.go`
- Modify: `nanocpa/internal/api/server_test.go`
- Modify: `nanocpa/internal/api/handlers/openai.go`
- Modify: `nanocpa/internal/api/handlers/openai_test.go`
- Modify: `docs/chapters/08-routing-and-hardening.md`

- [ ] **Step 1: Write failing routing tests**

Cover:
- same model across multiple auths alternates with round-robin
- round-robin state is isolated per model
- unsupported models fail cleanly
- disabled or cooldown auths are skipped

Run: `cd nanocpa && go test ./internal/auth ./internal/api -run 'TestManager_|TestServer_'`
Expected: FAIL until routing behavior matches chapter requirements

- [ ] **Step 2: Implement or finalize round-robin routing**

Keep the strategy intentionally simple:
- deterministic candidate order
- per-model next index tracking
- no cooldown scheduler
- no weighted routing

- [ ] **Step 3: Tighten downstream error handling**

Ensure:
- validation errors return `invalid_request_error`
- upstream errors are normalized into stable API errors
- default content type is correct

- [ ] **Step 4: Run the full `nanocpa` test suite**

Run: `cd nanocpa && go test ./...`
Expected: PASS

- [ ] **Step 5: Write `docs/chapters/08-routing-and-hardening.md`**

Explain:
- why multi-instance routing is the first real “CPA value add”
- why the tutorial stops here for the first edition
- what production features are intentionally deferred

- [ ] **Step 6: Commit and tag**

```bash
git add nanocpa/internal/auth/selector.go nanocpa/internal/auth/manager.go nanocpa/internal/auth/manager_test.go nanocpa/internal/api/server_test.go nanocpa/internal/api/handlers/openai.go nanocpa/internal/api/handlers/openai_test.go docs/chapters/08-routing-and-hardening.md
git commit -m "feat: add chapter 08 routing and hardening milestone"
git tag chapter-08-routing-and-hardening
```

## Task 10: Final Tutorial Polish And Cross-Checks

**Files:**
- Modify: `README.md`
- Modify: `docs/chapters/README.md`
- Modify: `docs/chapters/01-bootstrap.md`
- Modify: `docs/chapters/02-config.md`
- Modify: `docs/chapters/03-access.md`
- Modify: `docs/chapters/04-openai-surface.md`
- Modify: `docs/chapters/05-model-registry.md`
- Modify: `docs/chapters/06-runtime-skeleton.md`
- Modify: `docs/chapters/07-claude-provider.md`
- Modify: `docs/chapters/08-routing-and-hardening.md`
- Modify: `nanocpa/README.md`

- [ ] **Step 1: Verify every chapter document names its start and end tags**

Run: `rg -n "Start Tag|End Tag" README.md docs/chapters`
Expected: every chapter doc and the chapter index contains explicit tag references

- [ ] **Step 2: Verify every chapter includes a runnable verification command**

Run: `rg -n "^## Verification|Run:" docs/chapters`
Expected: each chapter has a verification section and at least one concrete command

- [ ] **Step 3: Verify the final codebase still passes all tests**

Run: `cd nanocpa && go test ./...`
Expected: PASS

- [ ] **Step 4: Verify the chapter tags exist**

Run: `git tag --list 'chapter-*' | sort`
Expected:
- `chapter-01-bootstrap`
- `chapter-02-config`
- `chapter-03-access`
- `chapter-04-openai-surface`
- `chapter-05-model-registry`
- `chapter-06-runtime-skeleton`
- `chapter-07-claude-provider`
- `chapter-08-routing-and-hardening`

- [ ] **Step 5: Commit tutorial polish**

```bash
git add README.md docs/chapters nanocpa/README.md
git commit -m "docs: polish byte-of-cpa tutorial flow"
```

## Local Plan Review

Because subagent delegation requires explicit user approval in this session, do a local review before execution:

- Read `docs/superpowers/specs/2026-03-26-byte-of-cpa-design.md`
- Read this plan top to bottom
- Confirm every spec requirement is covered by at least one task
- Confirm no task introduces first-edition scope creep
- Confirm each task ends in a runnable and teachable state

## Execution Order Summary

1. Add tutorial docs scaffold.
2. Rebuild chapter history from bootstrap to final CPA.
3. Tag each chapter milestone.
4. Polish cross-links and verification text.
5. Run the full `nanocpa` test suite and verify tag set.
