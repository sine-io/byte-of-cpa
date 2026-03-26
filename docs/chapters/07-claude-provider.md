# Chapter 07: Claude Provider

This chapter makes the provider logic concrete by translating OpenAI chat requests to Claude messages and back.

## What Problem This Chapter Solves

It implements the adapter between the OpenAI-compatible downstream surface and a Claude upstream so the system can actually complete a request.

## Why The Previous Chapter Is Not Enough

The runtime skeleton exists, but no provider code uses it yet. Without a concrete translator/executor, the system cannot service requests.

## New Concepts

- Translation between OpenAI chat payloads and Claude message APIs
- Claude executor wiring with API key and version headers
- Stable error handling for upstream responses

## Implementation

- Start Tag: `chapter-06-runtime-skeleton`
- End Tag: `chapter-07-claude-provider`
- Write translators and executor code that send `POST /v1/messages` to Claude with `x-api-key` and `anthropic-version: 2023-06-01`.
- Translate Claude responses back into OpenAI-compatible chat completions.

## Verification

- `cd nanocpa && go test ./internal/translator ./internal/runtime/executor ./internal/api/...`

## What You Have Now

- A runnable CPA that translates OpenAI chat requests through a Claude executor behind the runtime manager.

## What Comes Next

- Add model-aware routing and hardening so multiple upstream instances can be balanced (`Chapter 08: Routing and Hardening`).
