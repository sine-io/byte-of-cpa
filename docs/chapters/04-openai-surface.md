# Chapter 04: OpenAI Surface

This chapter ensures the downstream protocol surface matches OpenAI expectations so the CPA looks familiar before wiring up providers.

## What Problem This Chapter Solves

It defines concrete routes, request validation, and error shapes for `POST /v1/chat/completions` and `GET /v1/models` without yet translating requests upstream.

## Why The Previous Chapter Is Not Enough

Access control is in place, but the service does not yet speak the protocols downstream clients expect. Without handlers, readers cannot use the API.

## New Concepts

- OpenAI-compatible payload shapes
- Handler-level validation for required fields and JSON parsing
- Stable JSON error objects for invalid downstream requests

## Implementation

- Start Tag (planned): `chapter-03-access`
- End Tag (planned): `chapter-04-openai-surface`
- Register the chat and models handlers behind the access middleware.
- Validate request bodies, enforce content limits, and return consistent errors.
- Once the chapter is finalized the planned tags will be published so readers can inspect the resulting snapshot.

## Verification

Planned: `cd nanocpa && go test ./internal/api/... -run 'TestOpenAI|TestChatCompletions|TestModels'` once the downstream handlers are implemented.

## What You Have Now

- An HTTP service that looks like a downstream OpenAI API surface even though it has no upstream integration yet.

## What Comes Next

- Teach readers why models belong outside the handlers by registering availability from the configuration (`Chapter 05: Model Registry`).
