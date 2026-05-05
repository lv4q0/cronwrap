# cronwrap

Lightweight wrapper that adds logging, alerting, and retry logic to cron jobs.

## Installation

```bash
go install github.com/yourusername/cronwrap@latest
```

## Usage

Wrap any command by passing it as an argument to `cronwrap`:

```bash
cronwrap [options] -- <command>
```

**Example:**

```bash
cronwrap --retries 3 --alert-email ops@example.com -- /usr/local/bin/backup.sh
```

**Options:**

| Flag | Description | Default |
|------|-------------|---------|
| `--retries` | Number of retry attempts on failure | `0` |
| `--log-file` | Path to log output file | stdout |
| `--alert-email` | Email address to notify on failure | — |
| `--timeout` | Max execution time (e.g. `30s`, `5m`) | none |

**In your crontab:**

```cron
0 2 * * * cronwrap --retries 2 --log-file /var/log/backup.log -- /usr/local/bin/backup.sh
```

Each run is logged with timestamps, exit codes, and duration. If a command fails after all retries, an alert is triggered via the configured channel.

## Configuration

`cronwrap` can also be configured via a YAML file at `~/.cronwrap.yaml` or passed with `--config`:

```yaml
retries: 3
log_file: /var/log/cronwrap.log
alert_email: ops@example.com
timeout: 10m
```

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

## License

[MIT](LICENSE)