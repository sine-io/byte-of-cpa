# Chapter 02: Config

This chapter turns the hard-coded bootstrap surface into a configurable service with validation.

## What Problem This Chapter Solves

Configuration is the system boundary between deployment data and runtime behavior. Host, port, downstream `api_keys`, and upstream `providers` must come from data so operators can change environments without editing Go code.

## Why The Previous Chapter Is Not Enough

The bootstrap server compiles but relies on constants and a minimal schema. Without strict validation, bad configuration reaches runtime paths and fails later with less actionable errors.

## New Concepts

- YAML config loading
- Normalization and validation of config data before server startup
- Fail-fast validation for missing/invalid host, port, API keys, and providers
- Provider declarations as data (`providers[]` entries), not code paths in the bootstrap server

## Implementation

- Start Tag: `chapter-01-bootstrap`
- End Tag: `chapter-02-config`
- Expand config schema to first-edition fields only:
  - `host`
  - `port`
  - `api_keys`
  - `providers[].id`
  - `providers[].provider`
  - `providers[].api_key`
  - `providers[].base_url`
  - `providers[].models`
- Normalize strings during config loading, then validate before runtime starts.
- Reject unsupported provider values and missing provider fields so invalid configs fail early.
- Keep runtime bootstrap unchanged; this chapter adds data boundaries before later chapters consume them.

## Verification

Run:
- `cd nanocpa && go test ./internal/config -run 'TestConfig|TestLoad|TestValidate'`
- `cd nanocpa && go test ./internal/config`

## What You Have Now

- A server that refuses to start unless config input is complete and valid for first-edition runtime needs.
- A declarative provider snapshot in YAML that can evolve independently of bootstrap code.

## What Comes Next

- Enforce downstream authentication so only authorized clients can hit the proxy API (`Chapter 03: Access`).
