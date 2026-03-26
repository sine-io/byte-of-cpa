# Chapter 05: Model Registry

This chapter teaches that model names are routing inputs and that the `/v1/models` endpoint should reflect real availability.

## What Problem This Chapter Solves

It sources model availability from configuration instead of handler constants, allowing multiple upstream instances to advertise the models they support.

## Why The Previous Chapter Is Not Enough

The OpenAI handlers exist, but they have no idea which models are real versus hard-coded placeholders. We need to register models from configuration so routing logic can rely on actual data.

## New Concepts

- Model registry abstraction
- Listing configured models through `/v1/models`
- Rejecting chat requests for unsupported models before hitting the runtime

## Implementation

- Start Tag (planned): `chapter-04-openai-surface`
- End Tag (planned): `chapter-05-model-registry`
- Build a registry that records models per provider and exposes lookup helpers.
- Wire `/v1/models` to the registry and enforce support checks in the chat handler.
- The planned tags will be published with the chapter so readers can review the model registry milestone.

## Verification

- `cd nanocpa && go test ./internal/registry ./internal/api/...`

## What You Have Now

- The handlers understand which models the configuration actually enables, and unsupported requests are rejected early.

## What Comes Next

- Introduce the runtime skeleton so provider-specific code sits behind stable executor interfaces (`Chapter 06: Runtime Skeleton`).
