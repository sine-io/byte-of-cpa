# Chapter 01: Bootstrap

This chapter introduces the smallest runnable CPA shell to explain why every service needs a reliable constructor before it routes traffic.

## What Problem This Chapter Solves

Establishing a service boundary with explicit configuration and safe HTTP timeouts so we can build on a solid network foundation.

## Why The Previous Chapter Is Not Enough

Start here: there isn’t any previous chapter. During the rewrite, commit `62f02a2` serves only as a temporary pre-tag baseline reference. The intended Chapter 1 bootstrap snapshot will exist later, when `chapter-01-bootstrap` is published.

## New Concepts

- Service constructors
- Safe listen address creation
- HTTP server lifecycle peeling back to basics

## Implementation

- Start Baseline (temporary rewrite reference): `62f02a2`
- End Tag (planned): `chapter-01-bootstrap`
- Build a `main` function that loads config and creates a server.
- Keep the server simple: no auth, no handlers, only timeouts and listen address wiring.
- As of 2026-03-26, this chapter snapshot is not yet published. Readers will be able to inspect it once `chapter-01-bootstrap` exists.

## Verification

Planned: `cd nanocpa && go test ./internal/api -run 'TestServer_'` once the bootstrap milestone is implemented and tagged.

## What You Have Now

- A runnable Go binary that starts an HTTP server with the desired timeouts and listen address validation.

## What Comes Next

- Add configuration so the service becomes data-driven (`Chapter 02: Config`).
