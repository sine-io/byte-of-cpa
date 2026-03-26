# Chapter 01: Bootstrap

Chapter 1 builds the smallest useful CPA milestone: a process that reads startup config, constructs an HTTP server safely, and begins listening on the network.

## Why A CPA Starts As A Network Service

A CPA is ultimately an API-facing system. Before it can translate requests, authenticate callers, or talk to providers, it needs a stable process boundary that can accept connections, expose predictable listen settings, and avoid dangerous default timeouts. This chapter isolates that foundation so later chapters can add behavior without hiding the server lifecycle under proxy logic.

## Why This Chapter Does Not Proxy Anything Yet

Proxying too early mixes several concerns at once: request routing, auth, provider selection, upstream execution, and error translation. Chapter 1 deliberately avoids all of that. The goal here is narrower: prove that the service can boot from a minimal config file and construct an HTTP server with safe defaults. That gives the tutorial a real bootstrap milestone instead of a hand-waved starting point.

## Files In This Milestone

- `nanocpa/cmd/server/main.go` loads a config path from `-config`, reads the bootstrap config, and starts the server.
- `nanocpa/internal/config/config.go` defines the Chapter 1 config contract: only `host` and `port`.
- `nanocpa/internal/api/server.go` constructs the HTTP server with a formatted listen address and safe timeout values.
- `nanocpa/internal/api/server_test.go` verifies bootstrap server behavior.
- `nanocpa/internal/config/config_test.go` verifies the minimal config loader.
- `nanocpa/config.example.yaml` shows the Chapter 1 config shape.

## How To Run The Service

From the repository root:

```bash
cp nanocpa/config.example.yaml nanocpa/config.yaml
cd nanocpa
go run ./cmd/server -config config.yaml
```

The server will start listening on the configured `host:port`. At this stage it is only a bootstrap service, so it does not proxy requests yet.

## Verification

Run the bootstrap server tests:

```bash
cd nanocpa
go test ./internal/api -run 'TestServer_'
```

Optional config verification:

```bash
cd nanocpa
go test ./internal/config -run 'TestLoad_'
```

## What Comes Next

Chapter 2 expands the config layer beyond `host` and `port`, preparing the service for real routing and provider behavior.
