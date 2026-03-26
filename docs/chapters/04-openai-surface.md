# Chapter 04: OpenAI Surface

This chapter freezes the downstream OpenAI-shaped API surface before any provider execution path exists. The goal is to make the CPA speak the right protocol to clients first, then teach it how to satisfy that protocol in later chapters.

## What Problem This Chapter Solves

It defines concrete routes, request validation, request body limits, and stable error shapes for `POST /v1/chat/completions` and `GET /v1/models` without yet translating requests upstream.

## Why The Previous Chapter Is Not Enough

Access control is in place, but the service does not yet speak the protocols downstream clients expect. Without handlers, readers cannot use the API.

## Why Start With The Downstream Contract

The tutorial establishes the downstream contract before upstream integration because the HTTP boundary is the part external clients depend on. If routes, validation rules, response envelopes, and status codes are unstable, every later chapter has to change both the internals and the public API at the same time. Locking the surface first gives the rest of the tutorial a fixed target.

That also keeps responsibilities clean. Chapter 04 answers "what does this service promise to clients?" Later chapters answer "how does the service fulfill that promise?" Those are different problems, and teaching them separately makes the architecture easier to reason about.

## Why Compatibility Starts At The API Boundary

Compatibility does not begin when the first upstream provider call succeeds. It begins when a client can send a familiar request and receive familiar JSON back, including failure cases. For an OpenAI-compatible service, that means route names, required fields, list response shapes, and error envelopes matter immediately.

This is why Chapter 04 returns a stable API error even though no runner exists yet. A client should already be talking to an API that looks and behaves like the intended surface. Upstream execution is an implementation detail added later, not the start of compatibility.

## New Concepts

- OpenAI-compatible payload shapes
- Handler-level validation for required fields and JSON parsing
- Request body size limits at the HTTP boundary
- Stable JSON error objects for invalid downstream requests

## Implementation

- Start Tag: `chapter-03-access`
- End Tag: `chapter-04-openai-surface`
- Register the chat and models handlers behind the access middleware.
- Validate request bodies, enforce content limits, and return consistent OpenAI-style errors.
- Return the OpenAI `list` response shape for `GET /v1/models` even before any models are exposed.
- Return a stable `502 api_error` for chat requests that pass boundary validation but still have no upstream implementation.

## Verification

`cd nanocpa && go test ./internal/api/... -run 'TestOpenAI|TestChatCompletions|TestModels'`

## What You Have Now

- An HTTP service that exposes the OpenAI-compatible boundary clients expect even though it still has no upstream integration behind it.

## What Comes Next

- Teach readers why models belong outside the handlers by registering availability from the configuration (`Chapter 05: Model Registry`).
