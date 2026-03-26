# Chapter 03: Access

This chapter adds a protected middleware layer so only callers with valid Bearer tokens reach the downstream surface.

## What Problem This Chapter Solves

Downstream requests must authenticate independently from upstream provider credentials to keep the boundary clear and consistent.

## Why The Previous Chapter Is Not Enough

Configuration makes the server data-driven, but without access checks anybody can send requests. We need a thin middleware guard before unlocking handlers.

## New Concepts

- Bearer API key parsing
- Middleware enforcement of downstream access rules
- Clear JSON error responses for auth failures

## Implementation

- Start Tag: `chapter-02-config`
- End Tag: `chapter-03-access`
- Introduce API key extractor and middleware that validates tokens from headers.
- Return `401` responses with a stable error body when validation fails.

## Verification

- `cd nanocpa && go test ./internal/access`

## What You Have Now

- A service that refuses unauthorized requests before ever entering the business logic.

## What Comes Next

- Define the OpenAI-compatible API surface so the service looks like the downstream proxy we are teaching (`Chapter 04: OpenAI Surface`).
