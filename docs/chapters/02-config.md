# Chapter 02: Config

This chapter turns the hard-coded bootstrap surface into a configurable service with validation.

## What Problem This Chapter Solves

It keeps host, port, API keys, and upstream definitions in YAML so the service is no longer tied to a single environment.

## Why The Previous Chapter Is Not Enough

The bootstrap server compiles but relies on constants. Without config validation, invalid inputs leak into the runtime and cause confusing failures later.

## New Concepts

- YAML config loading
- Normalization and validation of downstream/upstream data
- Schema-driven awareness of API keys and providers

## Implementation

- Start Tag: `chapter-01-bootstrap`
- End Tag: `chapter-02-config`
- Add structs for hosts, ports, API keys, providers, and models.
- Load the config file, normalize defaults, validate required fields, and fail fast on bad input.

## Verification

- `cd nanocpa && go test ./internal/config`

## What You Have Now

- A server that reads well-formed YAML and refuses to start if the configuration is incomplete.

## What Comes Next

- Enforce downstream authentication so only authorized clients can hit the proxy API (`Chapter 03: Access`).
