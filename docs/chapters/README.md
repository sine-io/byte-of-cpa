# Chapter Guide

This directory holds the guided walkthrough. Each chapter file explains the problem, the code change, the verification produced, and how the milestone maps to a stable git tag.

## Chapter Order

Each chapter builds on the previous one in a single linear progression. The Start Tag for a chapter is the End Tag of the prior chapter (or `pre-tutorial` for chapter 1). The End Tag names the commit readers can checkout to review or return to that milestone.

| Chapter | Start Tag | End Tag | Verification |
| --- | --- | --- | --- |
| 01 Bootstrap | `pre-tutorial` | `chapter-01-bootstrap` | `cd nanocpa && go test ./internal/api -run 'TestServer_'` |
| 02 Config | `chapter-01-bootstrap` | `chapter-02-config` | `cd nanocpa && go test ./internal/config` |
| 03 Access | `chapter-02-config` | `chapter-03-access` | `cd nanocpa && go test ./internal/access` |
| 04 OpenAI Surface | `chapter-03-access` | `chapter-04-openai-surface` | `cd nanocpa && go test ./internal/api/... -run 'TestOpenAI|TestChatCompletions|TestModels'` |
| 05 Model Registry | `chapter-04-openai-surface` | `chapter-05-model-registry` | `cd nanocpa && go test ./internal/registry ./internal/api/...` |
| 06 Runtime Skeleton | `chapter-05-model-registry` | `chapter-06-runtime-skeleton` | `cd nanocpa && go test ./internal/auth ./internal/registry` |
| 07 Claude Provider | `chapter-06-runtime-skeleton` | `chapter-07-claude-provider` | `cd nanocpa && go test ./internal/translator ./internal/runtime/executor ./internal/api/...` |
| 08 Routing and Hardening | `chapter-07-claude-provider` | `chapter-08-routing-and-hardening` | `cd nanocpa && go test ./internal/auth ./internal/api -run 'TestManager_|TestServer_'` |

## Tag Conventions

- Chapter tags all use the prefix `chapter-` followed by a two-digit number and a short name (e.g., `chapter-04-openai-surface`).
- Start tags are the prior chapter's end tag; the first chapter begins at `pre-tutorial` because there is no prior tutorial tag.
- After finishing a chapter, tag the commit with the End Tag so readers can checkout that milestone.

## Moving Between Chapters

1. Checkout the current chapter's Start Tag (or `pre-tutorial`).
2. Follow the chapter document to understand the goal and walk through the implementation.
3. Run the verification command listed in the chapter's `## Verification` section to confirm the milestone.
4. Tag the commit with the End Tag to preserve the milestone.
5. Use the End Tag as the Start Tag for the next chapter.

## Verification Expectations

Every chapter document includes a `## Verification` section with one or more concrete commands exposing the milestone's runnable status. The commands in the table above are the minimal checks readers should run before moving to the next chapter. The final chapter rounds up with the more comprehensive routing checks listed in its verification section.
