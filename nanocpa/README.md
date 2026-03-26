# NanoCPA

NanoCPA is the working tutorial implementation for the chapters in `docs/chapters/`. The code here represents the current milestone of the `byte-of-cpa` tutorial, and each chapter document in `docs/chapters/` explains how to reach the next milestone.

## Tutorial Structure

- `nanocpa/` holds the runnable code that the tutorial walks through; treat it as the working tree readers build toward chapter by chapter.
- `docs/chapters/` contains the guided walkthrough. Each chapter defines a `Start Tag` and `End Tag` (see chapter files for the explicit tags). The tags mark the teaching milestones and let readers checkout each chapter's exact code.

## Config

Set `NANOCPA_CONFIG` to a YAML file path, or place `config.yaml` in the working directory.

Supported fields:
- `host`
- `port`
- `api_keys`
- `providers[]` with `id`, `provider`, `api_key`, `base_url`, `models[]`

See [`config.example.yaml`](./config.example.yaml) for a sample.
