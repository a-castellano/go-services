# go-services

[![pipeline status](https://git.windmaker.net/a-castellano/go-services/badges/master/pipeline.svg)](https://git.windmaker.net/a-castellano/go-services/pipelines)[![coverage report](https://git.windmaker.net/a-castellano/go-services/badges/master/coverage.svg)](https://a-castellano.gitpages.windmaker.net/go-services/coverage.html)[![Quality Gate Status](https://sonarqube.windmaker.net/api/project_badges/measure?project=a-castellano_go-services_7930712b-1aab-4ea2-a917-853d91ec9cc6&metric=alert_status&token=sqb_a42785fa06f27139e2134dd8221c060aa2324877)](https://sonarqube.windmaker.net/dashboard?id=a-castellano_go-services_7930712b-1aab-4ea2-a917-853d91ec9cc6)

This repository stores reusable services used by many of my projects. The aim is to save time and reduce code duplication by unifying common functionality in a single source.

## Architecture

The library is split into two layers that follow the dependency injection pattern:

- **`services/`** — high-level service abstractions. Each service defines a `Client` interface and delegates the actual work to whatever implementation you inject. This is what your application code depends on, which keeps it decoupled and easy to test with mocks.
- **`infra/`** — standalone backend clients (for example, a Redis driver or a RabbitMQ driver). They are kept separate from the services because they are reusable on their own: a driver happens to satisfy a given service's `Client` interface, but it can be used directly or back any other service whose interface it satisfies. Swap drivers without touching the code that uses the service.

To use a service you pick a driver from `infra/`, build it, and inject it into the matching service.

## Services

| Service | Description | Documentation |
|---------|-------------|---------------|
| **MemoryDatabase** | Key-value storage with TTL support over a memory database. | [services/memorydatabase](services/memorydatabase/Readme.md) |
| **MessageBroker** | Asynchronous messaging with persistent delivery over a message broker. | [services/messagebroker](services/messagebroker/Readme.md) |

Each service README documents its interface, the matching driver, a complete wired example, configuration, and testing.

## Drivers

The `infra/` packages are standalone backend clients, kept separate from the services on purpose: each one happens to satisfy a service `Client` interface today, but it is reusable on its own and can back any other service whose interface it satisfies. Use a driver injected into a service, or directly on its own.

| Driver | Backend | Satisfies | Documentation |
|--------|---------|-----------|---------------|
| **RedisClient** | Redis / Valkey | `memorydatabase.Client` | [infra/redis](infra/redis/Readme.md) |
| **RabbitmqClient** | RabbitMQ | `messagebroker.Client` | [infra/rabbitmq](infra/rabbitmq/Readme.md) |

## Installation

```bash
go get github.com/a-castellano/go-services
```

## Development

Go is not installed locally — all Go tasks run inside a Podman container using the same image as CI.

### Setting up the development environment

1. Clone the repository:

```bash
git clone https://git.windmaker.net/a-castellano/go-services.git
cd go-services
```

2. Start all services and the Go container with `podman-compose`:

```bash
podman-compose -f development/docker-compose.yml up -d
```

This starts:

- A Go development container
- A Valkey (Redis-compatible) server at `172.17.0.2`
- A RabbitMQ server at `172.17.0.30`

> **Note**: The Compose configuration uses hardcoded IP addresses (`172.17.0.0/16` subnet) to match the integration tests and ensure consistent behavior across environments. The Go module cache is persisted in `development/gomodcache/` on the host and mounted into the container, so dependencies are not downloaded on every run.

To stop and remove the containers when done:

```bash
podman-compose -f development/docker-compose.yml down
```

### Running tests

Exec into the Go container and run any `make` target:

```bash
podman exec -it development_golang_1 make test          # all unit tests
podman exec -it development_golang_1 make test_integration  # integration tests (need running services)
podman exec -it development_golang_1 make coverage      # coverage report
```

> **Note**: The container name (`development_golang_1`) may vary. Use `podman ps` to confirm the actual name. For an interactive shell, run `podman exec -it development_golang_1 /bin/bash`.

### Available make targets

| Target | Description |
|--------|-------------|
| `make help` | Show all available targets |
| `make lint` | Lint sources with `go vet` |
| `make test` | Run unit tests |
| `make test_integration` | Run integration tests (requires running services) |
| `make test_memorydatabase` | Run all MemoryDatabase tests |
| `make test_messagebroker_unit` | Run MessageBroker unit tests |
| `make test_redis` | Run all Redis driver tests |
| `make test_rabbitmq` | Run all RabbitMQ driver tests |
| `make race` | Run tests with the data race detector |
| `make msan` | Run tests with the memory sanitizer |
| `make coverage` | Generate the coverage report |
| `make coverhtml` | Generate the HTML coverage report |

Each service README lists the test targets scoped to that service.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

## Dependencies

- [go-types](https://git.windmaker.net/a-castellano/go-types) — configuration types
- [go-redis](https://github.com/redis/go-redis) — Redis client
- [amqp091-go](https://github.com/rabbitmq/amqp091-go) — RabbitMQ client
- [redismock](https://github.com/go-redis/redismock) — Redis mocking for tests
</content>
</invoke>
