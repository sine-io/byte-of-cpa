# Chapter 07: Claude Provider

This chapter adds the first real provider adapter. The important part is not "Claude support" by itself. The important part is that the CPA now proves its core job: accept an OpenAI-compatible request, translate it into a provider-specific upstream shape, execute it, then translate the response back into the downstream contract.

## What Problem This Chapter Solves

Chapter 06 introduced a generic runtime manager, but the request path still ended in a stable `502` because no concrete provider executor existed. This chapter closes that gap for one provider.

With Claude wired in, a configured Claude model can now complete an OpenAI-style `/v1/chat/completions` request end to end:

- OpenAI-style request enters the handler
- runtime manager selects a configured Claude auth
- Claude executor translates and sends `POST /v1/messages`
- Claude response is translated back into OpenAI chat completion shape

That round trip is the heart of a CPA. The product value is translation, not just proxying bytes.

## Why Translation Is The Core Of A CPA

If all providers already spoke the same wire format, a compatibility proxy would not be interesting. The real work is adapting one public contract to many upstream contracts without forcing clients to care which provider sits behind the model name.

For this chapter, the adapter is intentionally narrow:

- string-only message content
- system prompts lifted to Claude's top-level `system`
- one non-streaming completion response shape
- no tool calling
- no multimodal content

That narrowness is a feature, not a limitation in disguise. It keeps the first provider path small enough to verify completely before later chapters add more surface area.

## Why The Manager Stays Generic While The Provider Stays Concrete

The runtime manager should not know Claude request bodies, headers, or response formats. Its job is still generic:

- keep runtime auth registrations
- decide which configured auth can serve a model
- dispatch through the executor registered for that provider

The Claude executor is concrete on purpose. It owns the Claude-only details:

- `POST /v1/messages`
- `x-api-key`
- `anthropic-version: 2023-06-01`
- OpenAI-to-Claude request translation
- Claude-to-OpenAI response translation
- wrapping non-2xx upstream responses

That split keeps future providers additive. The manager does not need to become more provider-aware just because another adapter appears.

## New Concepts

- Request translation as a first-class provider responsibility
- Response translation back into a stable downstream contract
- Concrete executor registration during server startup
- Narrow provider snapshots instead of speculative "universal" adapters

## Implementation

- Start Tag: `chapter-06-runtime-skeleton`
- End Tag: `chapter-07-claude-provider`
- Add a Claude request translator from OpenAI chat completions to Claude messages.
- Add a Claude response translator back to OpenAI chat completion shape.
- Implement a Claude executor that validates runtime auth, sends the upstream request, and preserves safe response metadata.
- Register the Claude executor from server boot so configured Claude models execute through the generic runtime manager.

## Verification

- `cd nanocpa && go test ./internal/translator ./internal/runtime/executor ./internal/api/...`

## What You Have Now

- A working provider path for configured Claude models behind the existing OpenAI-compatible surface.
- A runtime manager that remains generic while dispatching to a concrete provider executor.
- Tests that lock the narrow translation contract and Claude executor behavior for this chapter snapshot.

## What Comes Next

Chapter 08 can build on this path instead of redesigning it. The next step is better routing and hardening across configured providers, not collapsing provider logic back into handlers.
