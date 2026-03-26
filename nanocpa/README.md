# NanoCPA

# NanoCPA

NanoCPA is the working tutorial implementation for the chapters in `docs/chapters/`. The code here represents the current milestone of the `byte-of-cpa` tutorial, and each chapter document in `docs/chapters/` explains how to reach the next milestone.

## Tutorial Structure

- `nanocpa/` holds the runnable code that the tutorial walks through; treat it as the working tree readers build toward chapter by chapter.
- `docs/chapters/` contains the guided walkthrough. Each chapter defines a `Start Tag` and `End Tag` (see chapter files for the explicit tags). The tags mark the teaching milestones and let readers checkout each chapter's exact code.
- Tags are planned milestones that will be published once each chapter is complete; readers can checkout the chapter's End Tag to review that milestone's code once it exists.

## Reader Flow

- `nanocpa/` always contains the living code; follow `main` until the first published chapter tag.
- When a chapter reaches a milestone, the corresponding `chapter-XX-...` tag is published so readers can checkout that snapshot.
- Start the tutorial on the most recent tag you care about, run the verification commands in the chapter docs, and move forward by checking out the next published End Tag once it exists.

## Config

Set `NANOCPA_CONFIG` to a YAML file path, or place `config.yaml` in the working directory.

Supported fields:
- `host`
- `port`
- `api_keys`
- `providers[]` with `id`, `provider`, `api_key`, `base_url`, `models[]`

See [`config.example.yaml`](./config.example.yaml) for a sample.
