# Chapter 05: Model Registry

This chapter teaches that model names are routing inputs and that the `/v1/models` endpoint must reflect the models actually configured on the server.

## What Problem This Chapter Solves

It moves model availability out of handler constants and into a registry built from configuration. That gives the API a real source of truth for which upstream providers can satisfy a request.

## Why The Previous Chapter Is Not Enough

The OpenAI handlers exist, but they do not yet know which model IDs are real. Without a registry, the API can accept a model name that no configured provider supports, or advertise a `/v1/models` list that has nothing to do with the current config.

Model IDs are routing inputs. A request arrives with `model: "gpt-4o-mini"` or `model: "claude-3-5-haiku"`, and the rest of the system uses that value to decide which provider is even eligible. If the model name is wrong, routing is wrong. That means model validation belongs at the edge, before any runtime execution exists.

## New Concepts

- Model registry abstraction
- Listing configured models through `/v1/models`
- Rejecting chat requests for unsupported models before hitting the runtime

## Implementation

- Start Tag: `chapter-04-openai-surface`
- End Tag: `chapter-05-model-registry`
- Build a registry that records models per configured client/provider and exposes lookup helpers.
- Construct the registry from `config.Providers` when the API server starts.
- Wire `/v1/models` to the registry so the response contains only real configured model IDs.
- Reject unsupported chat completion requests before any hypothetical upstream call.
- This chapter snapshot is captured by `chapter-05-model-registry`.

`/v1/models` should reflect real configured availability because clients use it as discovery. If the endpoint lists models that are not actually configured, the API is lying about what it can route. If it omits configured models, clients lose visibility into usable capacity. The handler should therefore render the registry snapshot, not a static placeholder.

## Verification

Run: `cd nanocpa && go test ./internal/registry ./internal/api/...`

## What You Have Now

- The server builds a model registry from configured providers.
- `/v1/models` returns the real configured model snapshot.
- Chat completions reject unsupported model IDs early and keep the stable `502 api_error` path for supported models with no runtime behind them yet.

## What Comes Next

- Introduce the runtime skeleton so provider-specific code sits behind stable executor interfaces (`Chapter 06: Runtime Skeleton`).
