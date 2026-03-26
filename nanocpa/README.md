# NanoCPA

NanoCPA contains the working tutorial code referenced by the chapters in `docs/chapters/`. Each chapter document maps to a tagged milestone, so readers can inspect the repository at each stage of the build.

## Tutorial Structure

- `nanocpa/` holds the runnable code that the tutorial walks through; treat it as the working tree readers build toward chapter by chapter.
- `docs/chapters/` contains the guided walkthrough. Chapter 1 starts from baseline commit `62f02a2`, and every later chapter starts from the previous chapter tag.
- Each `chapter-*` tag is a concrete milestone snapshot that readers can check out directly.

## How To Follow The Tutorial

- Read the chapter doc in `docs/chapters/`.
- Check out the chapter's End Tag to inspect the completed milestone.
- Run the verification commands from that chapter before moving on.

## Config

Set `NANOCPA_CONFIG` to a YAML file path, or place `config.yaml` in the working directory.

Supported fields:
- `host`
- `port`
- `api_keys`
- `providers[]` with `id`, `provider`, `api_key`, `base_url`, `models[]`

See [`config.example.yaml`](./config.example.yaml) for a sample.
