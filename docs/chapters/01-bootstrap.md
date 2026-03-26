# Chapter 01: Bootstrap

This chapter introduces the smallest runnable CPA shell to explain why every service needs a reliable constructor before it routes traffic.

## What Problem This Chapter Solves

Establishing a service boundary with explicit configuration and safe HTTP timeouts so we can build on a solid network foundation.

## Why The Previous Chapter Is Not Enough

Start here: there isn’t any previous chapter. The bootstrapping work moves the repository from the untagged baseline into something that compiles and listens.

## New Concepts

- Service constructors
- Safe listen address creation
- HTTP server lifecycle peeling back to basics

## Implementation

- Start Baseline (planned): untagged baseline before `chapter-01-bootstrap`
- End Tag (planned): `chapter-01-bootstrap`
- Build a `main` function that loads config and creates a server.
- Keep the server simple: no auth, no handlers, only timeouts and listen address wiring.
- These planned tags will be published once the chapter milestone is recorded so readers can jump directly to the completed snapshot.

## Verification

- `cd nanocpa && go test ./internal/api -run 'TestServer_'`

## What You Have Now

- A runnable Go binary that starts an HTTP server with the desired timeouts and listen address validation.

## What Comes Next

- Add configuration so the service becomes data-driven (`Chapter 02: Config`).
