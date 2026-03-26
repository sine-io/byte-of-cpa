# Chapter 08: Routing and Hardening

This chapter delivers the first minimal CPA value add by routing across multiple upstream instances and hardening error handling.

## What Problem This Chapter Solves

It shows how a CPA balances work across multiple configured upstreams and guards the surface with clear error contracts for validation and upstream failures.

## Why The Previous Chapter Is Not Enough

Claude requests now succeed, but only with a single upstream. Real CPAs need routing and better error hygiene to stay reliable at scale.

## New Concepts

- Round-robin selection per model
- Normalized validation and upstream error responses
- Hardening tests that cover routing boundaries

## Implementation

- Start Tag (planned): `chapter-07-claude-provider`
- End Tag (planned): `chapter-08-routing-and-hardening`
- Implement deterministic per-model round-robin selector behavior and SKIP disabled auths.
- Improve downstream error handling so validation errors return `invalid_request_error` objects and upstream failures are normalized.
- These planned tags will be published once the routing and hardening milestone is complete so readers can inspect the stabilized snapshot.

## Verification

- `cd nanocpa && go test ./internal/auth ./internal/api -run 'TestManager_|TestServer_'`
- `cd nanocpa && go test ./...`

## What You Have Now

- A CPA that routes across multiple upstream auths per model and exposes hardened downstream error semantics.

## What Comes Next

- The tutorial ends here for the first edition. Future chapters can explore production features, but this milestone already reflects the full minimal CPA we set out to teach.
