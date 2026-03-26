# Chapter Guide

This directory holds the guided walkthrough. Each chapter file explains the problem, the code change, and the verification steps that will map to a planned stable git tag once the milestone is implemented.

## Chapter Order

Each chapter builds on the previous one in a single linear progression. The Start Tag for a chapter is the End Tag of the prior chapter; Chapter 1 begins from the baseline at commit `62f02a2` that precedes `chapter-01-bootstrap`. The End Tag names the commit readers can checkout to review or return to that milestone once it is published.

| Chapter | Start Reference (planned) | End Tag (planned) | Planned verification command(s) |
| --- | --- | --- | --- |
| 01 Bootstrap | baseline at commit `62f02a2` (`git checkout 62f02a2`) | `chapter-01-bootstrap` | `cd nanocpa && go test ./internal/api -run 'TestServer_'` (planned) |
| 02 Config | `chapter-01-bootstrap` | `chapter-02-config` | `cd nanocpa && go test ./internal/config` (planned) |
| 03 Access | `chapter-02-config` | `chapter-03-access` | `cd nanocpa && go test ./internal/access` and `cd nanocpa && go test ./internal/api -run 'Test.*Middleware'` (planned) |
| 04 OpenAI Surface | `chapter-03-access` | `chapter-04-openai-surface` | `cd nanocpa && go test ./internal/api/... -run 'TestOpenAI|TestChatCompletions|TestModels'` (planned) |
| 05 Model Registry | `chapter-04-openai-surface` | `chapter-05-model-registry` | `cd nanocpa && go test ./internal/registry ./internal/api/...` (planned) |
| 06 Runtime Skeleton | `chapter-05-model-registry` | `chapter-06-runtime-skeleton` | `cd nanocpa && go test ./internal/auth ./internal/registry` and `cd nanocpa && go test ./internal/api -run 'TestServer_'` (planned) |
| 07 Claude Provider | `chapter-06-runtime-skeleton` | `chapter-07-claude-provider` | `cd nanocpa && go test ./internal/translator ./internal/runtime/executor ./internal/api/...` (planned) |
| 08 Routing and Hardening | `chapter-07-claude-provider` | `chapter-08-routing-and-hardening` | `cd nanocpa && go test ./internal/auth ./internal/api -run 'TestManager_|TestServer_'` and `cd nanocpa && go test ./...` (planned) |

## Tag Conventions

- Chapter tags all use the prefix `chapter-` followed by a two-digit number and a short name (e.g., `chapter-04-openai-surface`).
- Start tags are the prior chapter's end tag; the first chapter begins at the baseline commit `62f02a2` before `chapter-01-bootstrap` because there is no prior tutorial tag.
- These tag names are planned milestones. Maintainers will publish the tags when each chapter is complete so readers can checkout the referenced commits.

## Roadmap Mode

1. Follow each chapter document in order to understand the problem, the planned change, and the verification guidance before any tags are published.
2. For Chapter 1, start from the baseline commit `62f02a2` (`git checkout 62f02a2`). For later chapters, treat the Start Tag as the prior chapter's planned End Tag.
3. Run the planned verification commands listed in each chapter once its milestone code exists.
4. After you implement and verify a chapter, move toward the next planned tag by continuing with the next chapter.

## Snapshot Mode

1. After a chapter milestone is tagged, checkout that chapter's published End Tag to inspect the snapshot. For Chapter 1, the start reference remains baseline commit `62f02a2` (`git checkout 62f02a2`), while the first published milestone snapshot is `chapter-01-bootstrap`.
2. Run the verification commands documented in the chapter to confirm the milestone's behavior.
3. When you are ready for the next milestone, checkout the next published End Tag and repeat.

## Verification Expectations

Every chapter document includes a `## Verification` section describing the planned commands readers can run once that chapter's code is available. The table above summarizes those planned commands, but the exact test suites will align with the implemented milestone rather than assuming the current final-state suite already matches earlier chapters.
