# Alcatraz

Alcatraz provides safe, containerized Linux environments for AI agents. Each environment is a Docker container managed through the [docker/go-sdk](https://github.com/docker/go-sdk).

## Requirements

- Go 1.26+
- Docker daemon running locally

## Configuration

All settings live in `cmd/cfg/config.yml` and load through `pkg/config`:

```yaml
errnie:
  level: debug

environment:
  image: ubuntu:24.04
  exec_timeout: 10m
  memory_bytes: 4294967296
  nano_cpus: 2000000000
  pids_limit: 512
  working_dir: /workspace
  network_mode: none
  max_environments: 8
```

Override paths, in order:

1. `--config /path/to/config.yml`
2. `./cmd/cfg/config.yml`
3. `./config.yml`
4. `$HOME/.alcatraz/config.yml`
5. embedded default

## Usage

Build:

```bash
go build -o alcatraz ./cmd/main/
```

Start, stop, or destroy an environment by ID:

```bash
./alcatraz dev start
./alcatraz dev stop
./alcatraz dev destroy
```

## Library API

The root `alcatraz` package exposes a single `Environment` handle:

```go
environment := alcatraz.NewEnvironment(ctx, "dev")
err := environment.Start()
output, err := environment.Exec("uname -a")
err = environment.Stop()
err = environment.Close()
```

`Environment` also implements `io.Reader` and `io.Writer`. After `Exec`, reads come from the attached exec stdout/stderr stream and writes go to exec stdin.

Defaults come from `pkg/config.Environment` (`image`, command `/bin/bash`, etc.).

## Hardened agent environments

`pkg/environment` provides `AgentEnvironment` and `Registry` for multi-tenant, hardened containers. These types apply the full security profile from config: capability drops, read-only root filesystem, resource limits, no network by default, and exec timeouts.

Integration tests under `pkg/environment` start real containers and skip automatically when Docker is unavailable.

## Development

```bash
go test ./...
```
