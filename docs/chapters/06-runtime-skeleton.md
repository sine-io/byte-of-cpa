# Chapter 06: Runtime Skeleton

This chapter adds the first provider-agnostic runtime boundary: auth records, a manager, executors, and a selector. The system still does not talk to a real upstream provider yet. That is intentional. The goal here is to lock in the runtime shape before Chapter 07 adds a concrete adapter.

## What Problem This Chapter Solves

Chapter 05 gave us configuration loading, request access control, OpenAI-style HTTP surfaces, and a model registry. What it did not have was a stable runtime seam between:

- request handlers that accept OpenAI-style input
- runtime auth records loaded from config
- provider-specific execution code that will arrive later

Without that seam, the next chapter would have to mix provider logic directly into handlers or server boot code. This chapter prevents that coupling by introducing a manager that can select a runtime auth for a model and dispatch through a provider executor.

## Why The Previous Chapter Is Not Enough

The model registry can answer "which configured providers advertise this model?" That is necessary, but it is not sufficient for runtime behavior. At request time we also need to know:

- which runtime auths are active
- which auths are disabled or cooling down
- which provider should execute the request
- where provider-specific code can plug in without changing handler contracts

That coordination belongs in a runtime manager, not in HTTP handlers.

## Why This Is Not Over-Engineering

This abstraction appears before the first concrete executor because it solves a real boundary problem that already exists. It is not speculative layering for hypothetical providers.

The manager in this chapter has a narrow job:

- hold runtime auth registrations
- expose model support based on active auths
- select candidates for a model through a selector
- dispatch to a provider executor when one exists

That is the minimum structure needed to keep the request path stable while future chapters add concrete adapters. The alternative would be to let provider-specific execution leak into handlers now and then unwind that coupling later.

## New Concepts

- `Auth`
  Runtime auth state derived from config, including provider identity, credentials, and runtime lifecycle fields. In this chapter the config-to-server wiring still registers configured auths as active.
- `Selector`
  Chooses one auth from the candidate set for a model. The first implementation is round-robin, but the manager stays generic.
- `Executor`
  A provider adapter interface that accepts an OpenAI-style request payload plus the selected runtime auth.
- `Manager`
  Owns auth registration, executor registration, candidate lookup, selection, and dispatch.

## Implementation

- Start Tag: `chapter-05-model-registry`
- End Tag: `chapter-06-runtime-skeleton`
- Add runtime auth registration and provider executor registration to the generic manager.
- Keep candidate discovery independent from executor availability so model support and request execution are separate concerns.
- Wire server startup so config produces both the model registry and runtime auth registrations.
- Keep OpenAI handlers pointed at stable runtime interfaces, even though no concrete provider executor is registered yet.

## Why Provider Adapters Belong Behind Stable Interfaces

Provider adapters are the most volatile part of the runtime. They deal with upstream HTTP shapes, auth headers, retry behavior, translation, and provider-specific error semantics. If that code sits directly in handlers, every provider addition changes the core request path.

Putting adapters behind `Executor` keeps the stable core small:

- handlers validate and normalize HTTP input
- the manager owns runtime selection and dispatch
- executors own provider-specific behavior

That split lets later chapters add Claude or OpenAI adapters without rewriting the handler or server contracts introduced here.

## Verification

- `cd nanocpa && go test ./internal/auth ./internal/registry`
- `cd nanocpa && go test ./internal/api/...`

## What You Have Now

- A provider-agnostic runtime skeleton with config-backed auth registrations.
- Stable handler-to-runtime wiring that can reject unsupported models and still return a consistent `502 api_error` while executors are not implemented yet.
- A clean slot for future provider adapters without reintroducing handler-specific runtime logic.

## What Comes Next

Chapter 07 adds the first concrete provider adapter behind these interfaces. The request path should not need another structural rewrite; the next step is to implement an executor and translator, not to redesign the runtime again.
