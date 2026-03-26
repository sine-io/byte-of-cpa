# Chapter 06: Runtime Skeleton

This chapter introduces the generic runtime manager, executor, and selector so provider logic can be swapped in cleanly.

## What Problem This Chapter Solves

It structures the runtime around auth records, executors, and selectors rather than handler-specific plumbing, giving future providers a stable interface.

## Why The Previous Chapter Is Not Enough

Model availability is known, but the runtime still handles requests directly from the handlers. We need a manager/executor boundary before adding provider-specific code.

## New Concepts

- Runtime auth records containing upstream credentials
- Executor and selector interfaces
- Manager that coordinates execution based on model and provider

## Implementation

- Start Tag (planned): `chapter-05-model-registry`
- End Tag (planned): `chapter-06-runtime-skeleton`
- Define runtime interfaces and implement a manager that routes requests to executors via selectors.
- Wire the server setup to create auths from config and register them with the manager.
- These planned tags will be published once the runtime skeleton milestone is complete so the snapshot is accessible.

## Verification

Planned: `cd nanocpa && go test ./internal/auth ./internal/registry` to validate the manager/selector wiring.

Planned: `cd nanocpa && go test ./internal/api -run 'TestServer_'` to confirm the server boots against the new runtime skeleton.

## What You Have Now

- A runtime skeleton prepared for provider adapters while keeping the core architecture provider-agnostic.

## What Comes Next

- Add a concrete Claude executor and translator so requests carry through a real upstream (`Chapter 07: Claude Provider`).
