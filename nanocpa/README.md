# NanoCPA

NanoCPA is a minimal Go service scaffold for a proxy API server.

## Config

Set `NANOCPA_CONFIG` to a YAML file path, or place `config.yaml` in the working directory.

Supported fields:
- `host`
- `port`
- `api_keys`
- `providers[]` with `id`, `provider`, `api_key`, `base_url`, `models[]`

See [`config.example.yaml`](./config.example.yaml) for a sample.
