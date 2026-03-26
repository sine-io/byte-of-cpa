# NanoCPA

NanoCPA contains the working tutorial code referenced by the chapters in `docs/chapters/`. Each chapter document describes a teaching milestone and names the planned tag that will capture that milestone once it is published so readers can follow along.

## Tutorial Structure

- `nanocpa/` holds the runnable code that the tutorial walks through; treat it as the working tree readers build toward chapter by chapter.
- `docs/chapters/` contains the guided walkthrough. Each chapter defines a `Start Tag` and `End Tag` (see chapter files for the explicit tags). The tags mark the teaching milestones and let readers checkout each chapter's exact code.
- Tags are planned milestones that will be published once each chapter is complete; readers can checkout the chapter's End Tag to review that milestone's code once it exists.

## Roadmap Mode

- `nanocpa/` always contains the living code; start from the baseline commit (`git checkout 62f02a2`) or the most recent published chapter tag, if any exist.
- Read each chapter doc to understand the planned change, the Start/End tags, and the verification commands to run once the milestone is implemented.
- Execute the verification steps when the chapter code is ready, then move toward the next planned tag.

## Snapshot Mode

- Once a chapter tag is published, checkout its Start Tag to examine that milestone.
- Run the chapter's verification commands to confirm the behavior, then check out the End Tag when you're ready to progress to the next chapter.

## Config

Set `NANOCPA_CONFIG` to a YAML file path, or place `config.yaml` in the working directory.

Supported fields:
- `host`
- `port`
- `api_keys`
- `providers[]` with `id`, `provider`, `api_key`, `base_url`, `models[]`

See [`config.example.yaml`](./config.example.yaml) for a sample.
