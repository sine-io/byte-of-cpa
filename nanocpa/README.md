# NanoCPA

NanoCPA contains the working tutorial code referenced by the chapters in `docs/chapters/`. Each chapter document describes a teaching milestone and names the planned tag that will capture that milestone once it is published so readers can follow along.

## Tutorial Structure

- `nanocpa/` holds the runnable code that the tutorial walks through; treat it as the working tree readers build toward chapter by chapter.
- `docs/chapters/` contains the guided walkthrough. Each chapter defines a start reference and a planned `End Tag`; later chapters start from the previous chapter's published milestone, while Chapter 1 starts from baseline commit `62f02a2`.
- Tags are planned milestones that will be published once each chapter is complete; readers can checkout the chapter's End Tag to review that milestone's code once it exists.

## Roadmap Mode

- `nanocpa/` always contains the living code; start from the baseline commit (`git checkout 62f02a2`) for Chapter 1, or from the published prior milestone for later chapters.
- Read each chapter doc to understand the planned change, the start reference, the planned End Tag, and the verification commands to run once the milestone is implemented.
- Execute the verification steps when the chapter code is ready, then move toward the next planned milestone.

## Snapshot Mode

- Once a chapter tag is published, checkout that chapter's End Tag to inspect the milestone snapshot.
- For Chapter 1, the starting point is baseline commit `62f02a2`, and the first published snapshot is `chapter-01-bootstrap`.
- Run the chapter's verification commands to confirm the behavior, then move to the next published End Tag when you're ready to progress.

## Config

Set `NANOCPA_CONFIG` to a YAML file path, or place `config.yaml` in the working directory.

Supported fields:
- `host`
- `port`
- `api_keys`
- `providers[]` with `id`, `provider`, `api_key`, `base_url`, `models[]`

See [`config.example.yaml`](./config.example.yaml) for a sample.
