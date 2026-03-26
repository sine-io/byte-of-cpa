# Chapter 03: Access

This chapter adds a protected middleware layer so only callers with valid Bearer tokens reach the downstream surface.

## What Problem This Chapter Solves

Downstream caller authentication is a different concern from upstream provider authentication:

- Downstream auth: a client proves it can use this proxy (our `Authorization: Bearer <key>` check).
- Upstream provider auth: this service proves it can call OpenAI (provider credentials configured on the server side).

Keeping those separate prevents accidental trust leakage between boundaries and makes the access policy explicit at the proxy edge.

## Why The Previous Chapter Is Not Enough

Configuration makes the server data-driven, but without access checks anybody can send requests. We need a thin middleware guard before unlocking handlers.

Middleware is introduced here (instead of inside handlers) because access control is a cross-cutting transport concern. A single middleware check guarantees unauthorized requests are rejected before route logic runs, keeps handler code focused on business behavior, and gives one stable unauthorized response shape everywhere.

## New Concepts

- Bearer API key parsing
- Middleware enforcement of downstream access rules
- Clear JSON error responses for auth failures

## Implementation

- Start Tag: `chapter-02-config`
- End Tag: `chapter-03-access`
- API key validation accepts only `Authorization: Bearer <key>` with exact key matching.
- Middleware blocks invalid or missing credentials with `401`.
- Unauthorized responses use a stable JSON body:
  `{"error":{"message":"unauthorized","type":"invalid_request_error"}}`

## Verification

Run: `cd nanocpa && go test ./internal/access`

Run: `cd nanocpa && go test ./internal/api -run 'Test.*Unauthorized|Test.*Middleware'`

## What You Have Now

- A service boundary that consistently rejects unauthorized callers before entering downstream handlers.

## What Comes Next

- Define the OpenAI-compatible API surface so the service looks like the downstream proxy we are teaching (`Chapter 04: OpenAI Surface`).
