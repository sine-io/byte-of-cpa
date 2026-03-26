# NanoCPA

NanoCPA contains the working tutorial code referenced by the chapters in `docs/chapters/`. Each chapter document describes a teaching milestone and names the planned tag that will capture that milestone once it is published so readers can follow along.

As of 2026-03-26, no `chapter-*` tags are published yet. The chapter docs currently describe the planned tutorial rewrite and the future milestone snapshots.

## Tutorial Structure

- `nanocpa/` holds the runnable code that the tutorial walks through; treat it as the working tree readers build toward chapter by chapter.
- `docs/chapters/` contains the guided walkthrough. Each chapter defines a temporary or planned start reference and a planned `End Tag`; later chapters are intended to start from the previous chapter's published milestone, while Chapter 1 currently uses `62f02a2` only as a temporary rewrite reference.
- Tags are planned milestones that will be published once each chapter is complete; readers can checkout the chapter's End Tag to review that milestone's code once it exists.

## Roadmap Mode

- `nanocpa/` always contains the living code while the tutorial rewrite is in progress.
- Read each chapter doc to understand the intended change, the planned milestone boundaries, and the verification work that will be finalized when the chapter is implemented.
- Treat the references in these docs as rewrite scaffolding until the chapter tags are eventually published.

## Snapshot Mode

- Once chapter tags are published in the future, checkout a chapter's End Tag to inspect that milestone snapshot.
- For Chapter 1, there will be no published Start Tag; the first published snapshot will be `chapter-01-bootstrap`.
- Run the chapter's verification commands to confirm the behavior, then move to the next published End Tag when you're ready to progress.

## Config

Set `NANOCPA_CONFIG` to a YAML file path, or place `config.yaml` in the working directory.

Supported fields:
- `host`
- `port`
- `api_keys`
- `providers[]` with `id`, `provider`, `api_key`, `base_url`, `models[]`

See [`config.example.yaml`](./config.example.yaml) for a sample.
