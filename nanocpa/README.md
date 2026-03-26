# NanoCPA

NanoCPA contains the working tutorial code referenced by the chapters in `docs/chapters/`. Each chapter document describes a teaching milestone and names the planned tag that will capture that milestone once it is published so readers can follow along.

## Tutorial Structure

- `nanocpa/` holds the runnable code that the tutorial walks through; treat it as the working tree readers build toward chapter by chapter.
- `docs/chapters/` contains the guided walkthrough. Each chapter defines a `Start Tag` and `End Tag` (see chapter files for the explicit tags). The tags mark the teaching milestones and let readers checkout each chapter's exact code.
- Tags are planned milestones that will be published once each chapter is complete; readers can checkout the chapter's End Tag to review that milestone's code once it exists.

## Reader Flow

- `nanocpa/` always contains the living code; start from the baseline commit or the most recent published chapter tag that exists.
- Each chapter doc names the planned `chapter-XX-...` tag that will be published when the milestone is ready; readers can checkout that tag once it appears to revisit the snapshot.
- Run the verification commands in the chapter docs before moving on. When the next End Tag is published, checkout that tag to continue the tutorial.

## Config

Set `NANOCPA_CONFIG` to a YAML file path, or place `config.yaml` in the working directory.

Supported fields:
- `host`
- `port`
- `api_keys`
- `providers[]` with `id`, `provider`, `api_key`, `base_url`, `models[]`

See [`config.example.yaml`](./config.example.yaml) for a sample.
