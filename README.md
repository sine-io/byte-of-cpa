# Byte of CPA Tutorial

The `byte-of-cpa` repository is a tutorial-first guide to building a minimal CPA (Chatbot Proxy API) while walking readers through a controlled progression from a bare service to a routed CPA that currently exposes the Claude provider while remaining architected to admit additional providers later.

## Purpose

Teach working programmers how to start with a runnable Go service, layer in configuration and auth, define an OpenAI-compatible surface, register upstream models, introduce a provider-agnostic runtime, add a concrete Claude adapter, and finally route across multiple upstream instances. Every chapter makes one clear architectural move so readers understand why each layer exists.

## Audience

This material targets programmers who are already comfortable reading Go code and reasoning about HTTP services. It does not re-teach basic syntax, HTTP, or JSON, but it does explain the architectural boundaries that make a CPA a teachable, real system.

## Repository Layout

- `nanocpa/` contains the working tutorial code. Every chapter builds on this single codebase rather than shipping separate copies.
- `docs/chapters/` hosts the guided walkthrough. Each markdown file explains the chapter goal, implementation decisions, tests, and git tags.
- `docs/superpowers/` keeps specs, plans, and other supporting material for planning the tutorial progression.

## How Chapter Tags Work

Each chapter ends with a stable git tag (`chapter-01-bootstrap` through `chapter-08-routing-and-hardening`). Chapter 1 starts from the pre-chapter baseline commit `62f02a2`, and every later chapter starts from the previous chapter's tag. The [chapter guide](docs/chapters/README.md) lists the exact sequence and verification commands.

## Chapters

Each chapter doc captures a concrete milestone: the problem it solves, the architectural change, the associated Start/End references, and the verification commands for that snapshot.

1. [Chapter 01: Bootstrap](docs/chapters/01-bootstrap.md) — start a safe HTTP server with explicit construction boundaries, timeouts, and no downstream surface yet.
2. [Chapter 02: Config](docs/chapters/02-config.md) — load and validate YAML configuration so upstreams and access rules become data-driven, not hard-coded.
3. [Chapter 03: Access](docs/chapters/03-access.md) — add middleware that enforces downstream Bearer API keys to separate downstream auth from upstream credentials.
4. [Chapter 04: OpenAI Surface](docs/chapters/04-openai-surface.md) — expose `POST /v1/chat/completions` and `GET /v1/models` with stable error shapes before wiring providers.
5. [Chapter 05: Model Registry](docs/chapters/05-model-registry.md) — source model availability and routing data from configuration and reject unsupported models.
6. [Chapter 06: Runtime Skeleton](docs/chapters/06-runtime-skeleton.md) — introduce manager, executor, and selector interfaces so provider logic sits behind stable runtime boundaries.
7. [Chapter 07: Claude Provider](docs/chapters/07-claude-provider.md) — implement the Claude adapter that translates OpenAI chat requests to Claude messages and back, proving the translation step.
8. [Chapter 08: Routing and Hardening](docs/chapters/08-routing-and-hardening.md) — route across multiple upstream instances with round-robin selection, hardened error handling, and stabilized tests.
