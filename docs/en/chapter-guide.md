# Chapter Guide

This directory holds the guided walkthrough. Each chapter file explains the problem, the code change, and the verification steps for one tagged tutorial milestone.

## Chapter Order

Each chapter builds on the previous one in a single linear progression. Chapter 1 starts from the pre-chapter baseline commit `62f02a2`. Every later chapter starts from the prior chapter's End Tag.

| Chapter | Start Reference | End Tag | Verification command(s) |
| --- | --- | --- | --- |
| 01 Bootstrap | baseline commit `62f02a2` | `chapter-01-bootstrap` | `cd nanocpa && go test ./internal/api -run 'TestServer_'` |
| 02 Config | `chapter-01-bootstrap` | `chapter-02-config` | `cd nanocpa && go test ./internal/config -run 'TestLoad|TestValidate'` and `cd nanocpa && go test ./internal/config` |
| 03 Access | `chapter-02-config` | `chapter-03-access` | `cd nanocpa && go test ./internal/access` and `cd nanocpa && go test ./internal/api -run 'Test.*Unauthorized|Test.*Middleware'` |
| 04 OpenAI Surface | `chapter-03-access` | `chapter-04-openai-surface` | `cd nanocpa && go test ./internal/api/... -run 'TestOpenAI|TestChatCompletions|TestModels'` |
| 05 Model Registry | `chapter-04-openai-surface` | `chapter-05-model-registry` | `cd nanocpa && go test ./internal/registry ./internal/api/...` |
| 06 Runtime Skeleton | `chapter-05-model-registry` | `chapter-06-runtime-skeleton` | `cd nanocpa && go test ./internal/auth ./internal/registry` and `cd nanocpa && go test ./internal/api/...` and `cd nanocpa && go test ./internal/runtime/executor/...` |
| 07 Claude Provider | `chapter-06-runtime-skeleton` | `chapter-07-claude-provider` | `cd nanocpa && go test ./internal/translator ./internal/runtime/executor ./internal/api/...` |
| 08 Routing and Hardening | `chapter-07-claude-provider` | `chapter-08-routing-and-hardening` | `cd nanocpa && go test ./internal/auth ./internal/api -run 'TestManager_|TestServer_'` and `cd nanocpa && go test ./...` |

## Tag Conventions

- Chapter tags all use the prefix `chapter-` followed by a two-digit number and a short name (e.g., `chapter-04-openai-surface`).
- Chapter 1 starts from baseline commit `62f02a2`.
- Every later chapter starts from the prior chapter's End Tag.
- Each `chapter-*` tag is a concrete milestone snapshot that readers can check out directly.

## Suggested Flow

1. Start from the chapter's Start Reference.
2. Read the chapter doc to understand the goal and the architectural change.
3. Check out the chapter's End Tag to inspect the completed milestone.
4. Run the chapter verification commands.
5. Move to the next chapter and repeat.

## Verification Expectations

Every chapter document includes a `## Verification` section with runnable commands for that milestone. The table above is the quick index; the chapter docs provide the context for why those checks matter.
