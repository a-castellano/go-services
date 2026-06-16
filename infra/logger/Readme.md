# Logger

Structured logging for the services in this module, built on the standard
library [`log/slog`](https://pkg.go.dev/log/slog). It implements the **"logger in
the context"** pattern: the logger travels inside `context.Context`, so services
and drivers retrieve it from the `ctx` they already receive instead of taking an
extra parameter or a package-level global.

- **Package**: `logger`
- **Import**: `github.com/a-castellano/go-services/infra/logger`

> **Scope.** The package targets `slog` and lives under `infra/`. It is complete
> as it stands; it would only change if we decide to support a different logger
> backend. Because consumers depend on the `Client` interface rather than on
> `slog`, such a backend could be added — and the package promoted to
> `services/` — without changing any call sites.

## Overview

The package has three parts:

1. **`Client`** — the contract every logger satisfies. It deliberately exposes
   only the `...Context` methods, because those forward `ctx` to the handler,
   which is what makes future trace correlation (OpenTelemetry) work without
   touching the call sites.

   ```go
   type Client interface {
       InfoContext(ctx context.Context, msg string, args ...any)
       ErrorContext(ctx context.Context, msg string, args ...any)
       DebugContext(ctx context.Context, msg string, args ...any)
       WarnContext(ctx context.Context, msg string, args ...any)
   }
   ```

2. **Context helpers** — `WithLogger` stores a logger in the context (once, at
   startup) and `FromContext` retrieves it anywhere. `FromContext` **never
   returns nil**: if the context carries no logger it falls back to a wrapped
   `slog.Default()`, which is what prevents nil-pointer panics across every
   service and driver.

3. **`SlogLogger`** — the `slog`-backed implementation. `NewLogger` builds it
   from a configuration and returns the concrete `*SlogLogger` (not `Client`) so
   the construction site keeps access to the runtime level controls (`SetLevel`,
   `SetDefaultLevel`). Code that only logs receives it as `Client` through the
   context and therefore cannot change the level.

Logs always go to **stdout**. There is no stdout/stderr/file selection: apps run
wrapped by systemd (or Podman under systemd) and journald manages the final
destination (forwarding to syslog, files or remote sinks) with no code change.
Severity lives in the structured `level` field, so a single stream is enough and
event ordering is preserved.

## Usage

The logger is meant to be **built once at program startup, stored in the
context, and then read from the context** everywhere else.

### 1. Wire it at application startup

```go
package main

import (
    "context"
    "log"

    "github.com/a-castellano/go-services/infra/logger"
    slogconfig "github.com/a-castellano/go-types/slog"
)

func main() {
    // Build and validate the configuration from environment variables.
    config, err := slogconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Build the logger and place it in the root context.
    appLogger := logger.NewLogger(config)
    ctx := logger.WithLogger(context.Background(), appLogger)

    // Pass ctx down into your services/drivers as usual.
    run(ctx)
}
```

`appLogger` is a `*SlogLogger`: keep that reference if you want to change the log
level at runtime (see below). Everything downstream only needs the `ctx`.

### 2. Log from a function or another service

Any code that already receives a `context.Context` retrieves the logger from it
— no new parameter, no global:

```go
func (m *MemoryDatabase) WriteString(ctx context.Context, key, value string, ttl int) error {
    log := logger.FromContext(ctx)

    if !m.client.IsClientInitiated() {
        log.ErrorContext(ctx, "client not initiated", "key", key)
        return errors.New("client not initiated")
    }

    log.DebugContext(ctx, "writing string", "key", key, "ttl", ttl)
    return m.client.WriteString(ctx, key, value, ttl)
}
```

`ctx` appears twice on purpose: `FromContext(ctx)` retrieves the logger, and
`...Context(ctx, ...)` carries request-scoped data to the handler. Because
`FromContext` never returns nil, this is safe even if the caller forgot to call
`WithLogger` — you simply get the default logger instead of a panic.

> **Always use the `...Context` variants** (`InfoContext`, `ErrorContext`, ...).
> They are the only methods the `Client` interface exposes, and the only ones
> that forward `ctx` to the handler.

### 3. Change the log level at runtime

The handler reads a shared `*slog.LevelVar`, so a level change applies to every
record from then on, with no need to rebuild the logger. Only the holder of the
concrete `*SlogLogger` (the startup code) can do this:

```go
appLogger.SetLevel(slog.LevelError) // raise the level temporarily
appLogger.SetDefaultLevel()         // restore the level from the configuration
```

This is the hook for, e.g., a `SIGHUP` handler or an admin endpoint that toggles
verbose logging without a restart.

## API reference

#### `NewLogger(config *slogconfig.Config) *SlogLogger`

Builds a `slog`-backed logger from a configuration and writes to `os.Stdout`.
The config **must already be generated and validated** by
`slogconfig.NewConfig()`; `NewLogger` does not validate its input, so passing an
invalid or `nil` config is a programming error and is not guarded against.

#### `WithLogger(ctx context.Context, l Client) context.Context`

Returns a child context carrying `l`. Call it once, at startup. Stored under a
private key type, so it cannot collide with other context values.

#### `FromContext(ctx context.Context) Client`

Retrieves the logger from the context. Never returns nil: with no logger present
it returns a wrapped `slog.Default()` fallback.

#### `(*SlogLogger) SetLevel(l slog.Level)`

Sets the active log level at runtime. Affects every logger sharing the same
handler immediately.

#### `(*SlogLogger) SetDefaultLevel()`

Resets the active level back to the one provided in the configuration.

`*SlogLogger` also embeds `*slog.Logger`, so all of its methods (`InfoContext`,
`With`, ...) are available on the concrete type.

## Configuration

Configuration comes from the [`go-types`](https://git.windmaker.net/a-castellano/go-types)
library via `slogconfig.NewConfig()` (package
`github.com/a-castellano/go-types/slog`), which reads these environment
variables:

- `APP_NAME` — application name, added as an `app` attribute on every record.
  Required, no default.
- `SLOG_LEVEL` — default level: `Debug`, `Info`, `Warn` or `Error` (default:
  `Info`).
- `SLOG_FORMAT` — output format: `JSON` or `plain` (default: `JSON`).
- `SLOG_ADD_SOURCE` — add `file:line` to records: `true` or `false` (default:
  `true`).

```bash
export APP_NAME=my-service
export SLOG_LEVEL=Info
export SLOG_FORMAT=JSON
export SLOG_ADD_SOURCE=true
```

## Testing

Unit tests inject a `bytes.Buffer` as the writer and assert on the captured
output (no real stdout needed). See the [development guide](../../Readme.md#development)
for the container setup.

```bash
make test_logger_unit   # unit only
make test_logger        # all logger tests
```

## Dependencies

- [go-types](https://git.windmaker.net/a-castellano/go-types) — configuration types (`slog` config)
- [log/slog](https://pkg.go.dev/log/slog) — standard library structured logging
