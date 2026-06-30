# OpenTelemetry

OpenTelemetry SDK **startup** for the services in this module. The package wires
the trace pipeline once, at program start, and hands back a `shutdown` function.
It deliberately exposes **no client object**: instrumentation across `services/`
and `infra/` talks to the **global OpenTelemetry API** (`otel.Tracer(...)`), not
to a type from this package.

- Package: `opentelemetry`
- Import: `github.com/a-castellano/go-services/infra/opentelemetry`

> Scope. The package currently sets up traces only, exporting to stdout. Metrics,
> log correlation, OTLP export and cross-service propagation are built on top of
> this same startup seam later, without changing how applications call it.

## Why no `Client` interface

The logger exposes a `Client` because business code calls it on every log line
and the logger travels in the `context`. Telemetry is different: OpenTelemetry
**splits the API from the SDK** — the API is what you instrument with, the SDK is
configured once at startup. Instrumented code never holds an SDK object; it calls
the global `otel.Tracer(...)`, whose provider this package registers at startup.
The only behaviour that outlives initialization is shutdown, so the package
surface is a single constructor returning a `shutdown func(context.Context) error`.

This mirrors the logger's non-nil guarantee, but for free: if the SDK is never
registered, OpenTelemetry's default global providers and propagator are **no-ops**.
A span started with telemetry off (`otel.Tracer(...).Start`, `span.End()`,
`trace.SpanFromContext`) runs harmlessly and never panics, so no service has to
check whether telemetry is active.

## Behaviour

`SetupOpenTelemetry` builds, in the enabled path:

1. A **`Resource`** whose `service.name` is taken from `APP_NAME` (the same
   variable the logger uses for its `app` attribute), via `semconv.ServiceName`.
   This is what labels each service in a distributed trace, so it must be
   distinct per service and stable across runs.
2. A **`TracerProvider`** with a batch span processor exporting to stdout
   (`stdouttrace`, pretty-printed).
3. A composite **W3C propagator** (`TraceContext` + `Baggage`), set as the global
   `TextMapPropagator`. `TraceContext` carries the **trace identity** across
   processes; `Baggage` carries cross-cutting key-values. Registering it is
   mandatory: a missing propagator does not error, it silently makes inject and
   extract no-ops, which would later disconnect cross-service traces.

It returns a `shutdown` that flushes and closes the provider. Two guarantees
make it uniform to call from `main`:

- **Disabled** (`Enabled == false`): the SDK is not wired and a no-op `shutdown`
  is returned. `main` always calls and `defer`s the same way.
- **Enabled but init fails**: a no-op `shutdown` is returned together with the
  error, so the application can log it and keep running rather than aborting
  over an observability failure.

## Usage

### 1. Wire it at application startup

```go
package main

import (
    "context"
    "log"

    "github.com/a-castellano/go-services/infra/opentelemetry"
    otelconfig "github.com/a-castellano/go-types/types/opentelemetry"
)

func main() {
    // Build and validate the configuration from environment variables.
    config, err := otelconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // shutdown is always non-nil, so the defer below is safe even on error.
    shutdown, err := opentelemetry.SetupOpenTelemetry(ctx, config)
    if err != nil {
        // Telemetry failed to start; the app keeps running without it.
        log.Printf("telemetry setup failed: %v", err)
    }
    defer func() {
        if err := shutdown(ctx); err != nil {
            log.Printf("telemetry shutdown failed: %v", err)
        }
    }()

    run(ctx)
}
```

### 2. Instrument code with the global API

Anything that already receives a `context.Context` opens a span through the
global tracer and propagates the `ctx` to its children. No dependency on this
package is needed:

```go
func (m *MemoryDatabase) WriteString(ctx context.Context, key, value string, ttl int) error {
    ctx, span := otel.Tracer("memorydatabase").Start(ctx, "WriteString")
    defer span.End()

    span.SetAttributes(attribute.String("operation", "write_string"))
    // ... do the work, pass ctx down so child spans nest correctly ...
    return nil
}
```

With telemetry disabled these calls resolve to no-ops, so the same code is safe
in both modes.

## API reference

#### `SetupOpenTelemetry(ctx context.Context, config *opentelemetryconfig.Config) (func(context.Context) error, error)`

Initializes the trace pipeline and registers the global `TracerProvider` and
`TextMapPropagator`. Returns a `shutdown` function and an error.

The returned `shutdown` is always non-nil: in the disabled path and on init
failure it is a no-op, so `main` can `defer` it unconditionally. The `config`
must already be generated and validated by `opentelemetryconfig.NewConfig()`;
`SetupOpenTelemetry` does not re-validate it.

## Configuration

Configuration comes from the [`go-types`](https://git.windmaker.net/a-castellano/go-types)
library via `opentelemetryconfig.NewConfig()` (package
`github.com/a-castellano/go-types/types/opentelemetry`), which reads:

- `APP_NAME` — application name, used as the telemetry `service.name`. Required,
  no default. It is the **single source** of the service name, shared with the
  logger so logs and traces never diverge.
- `ENABLE_TELEMETRY` — whether telemetry is active: `true` or `false`. Optional;
  when unset it defaults to `false` (telemetry is **opt-in**).

Two standard variables are deliberately rejected by `NewConfig()`:

- `OTEL_SERVICE_NAME` — forbidden, because `APP_NAME` is the only accepted source
  for `service.name`.
- `OTEL_RESOURCE_ATTRIBUTES` — forbidden for the time being.

Everything else (exporter endpoint, sampling, protocol) is left to the SDK's own
standard `OTEL_*` variables.

```bash
export APP_NAME=my-service
export ENABLE_TELEMETRY=true
```

## Testing

Unit tests inject a `tracetest.SpanRecorder` through the provider builder, so
emitted spans are inspected in memory as structs (name, attributes, the Resource
`service.name`) instead of parsing the stdout JSON. The propagator is tested
with an inject/extract round-trip over a `propagation.MapCarrier`. See the
[development guide](../../Readme.md#development) for the container setup.

```bash
make test_opentelemetry_unit
```

## Dependencies

- [go-types](https://git.windmaker.net/a-castellano/go-types) — configuration types (`opentelemetry` config)
- [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go) — SDK, API and stdout exporter
