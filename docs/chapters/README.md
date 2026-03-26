# Chapter Guide

This directory holds the guided walkthrough. Each chapter file explains the problem, the code change, and the verification steps that will map to a planned stable git tag once the milestone is implemented.

As of 2026-03-26, no `chapter-*` tags are published yet. This document currently defines the roadmap for the rewrite and the snapshot navigation that will exist after the tags are published.

## Chapter Order

Each chapter builds on the previous one in a single linear progression. Later chapters are planned to start from the prior chapter's published End Tag. During the rewrite, Chapter 1 uses commit `62f02a2` only as a temporary pre-tag reference while the milestone history is being rebuilt. The End Tag names the future published snapshot for that chapter.

| Chapter | Start Reference (planned) | End Tag (planned) | Planned verification command(s) |
| --- | --- | --- | --- |
| 01 Bootstrap | temporary rewrite reference `62f02a2` before the first tag exists | `chapter-01-bootstrap` | `cd nanocpa && go test ./internal/api -run 'TestServer_'` (planned) |
| 02 Config | `chapter-01-bootstrap` | `chapter-02-config` | `cd nanocpa && go test ./internal/config` (planned) |
| 03 Access | `chapter-02-config` | `chapter-03-access` | `cd nanocpa && go test ./internal/access` and `cd nanocpa && go test ./internal/api -run 'Test.*Middleware'` (planned) |
| 04 OpenAI Surface | `chapter-03-access` | `chapter-04-openai-surface` | `cd nanocpa && go test ./internal/api/... -run 'TestOpenAI|TestChatCompletions|TestModels'` (planned) |
| 05 Model Registry | `chapter-04-openai-surface` | `chapter-05-model-registry` | `cd nanocpa && go test ./internal/registry ./internal/api/...` (planned) |
| 06 Runtime Skeleton | `chapter-05-model-registry` | `chapter-06-runtime-skeleton` | `cd nanocpa && go test ./internal/auth ./internal/registry` and `cd nanocpa && go test ./internal/api -run 'TestServer_'` (planned) |
| 07 Claude Provider | `chapter-06-runtime-skeleton` | `chapter-07-claude-provider` | `cd nanocpa && go test ./internal/translator ./internal/runtime/executor ./internal/api/...` (planned) |
| 08 Routing and Hardening | `chapter-07-claude-provider` | `chapter-08-routing-and-hardening` | `cd nanocpa && go test ./internal/auth ./internal/api -run 'TestManager_|TestServer_'` and `cd nanocpa && go test ./...` (planned) |

## Tag Conventions

- Chapter tags all use the prefix `chapter-` followed by a two-digit number and a short name (e.g., `chapter-04-openai-surface`).
- Start tags are planned to be the prior chapter's End Tag; Chapter 1 has no published Start Tag and currently uses `62f02a2` only as a temporary rewrite reference before `chapter-01-bootstrap` exists.
- These tag names are planned milestones. Maintainers will publish the tags when each chapter is complete so readers can checkout the referenced commits.

## Roadmap Mode

1. Follow each chapter document in order to understand the intended progression, milestone boundaries, and planned verification work before any tags are published.
2. Treat the Start/End references in these docs as planning markers for the rewrite, not as published tutorial snapshots you can navigate today.
3. Use the chapter notes to understand what the eventual milestone history is supposed to contain and how each milestone will be verified once implemented.

## Snapshot Mode

1. Once the rewrite is complete and tags are published, inspect a chapter milestone by checking out that chapter's published End Tag.
2. For Chapter 1, there will be no published Start Tag; the first published snapshot will be `chapter-01-bootstrap`.
3. Run the verification commands documented in the chapter to confirm the milestone's behavior, then move to the next published End Tag when needed.

## Verification Expectations

Every chapter document includes a `## Verification` section describing the planned commands readers can run once that chapter's code is available. The table above summarizes those planned commands, but the exact test suites will align with the implemented milestone rather than assuming the current final-state suite already matches earlier chapters.
