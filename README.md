# Migra

A CLI tool for orchestrating database migrations across microservices.

Migra coordinates migration execution across multiple services and frameworks without replacing your existing migration tools. It handles Django, Laravel, Prisma, and other framework-specific migration systems in a unified workflow.

## Features

- Supports multiple frameworks: Django, Laravel, Prisma
- Multi-tenant deployments with database-per-tenant support
- Sequential or parallel execution strategies
- Dry-run mode and stop-on-failure options
- Local state tracking for migration history
- Structured logging (console or JSON)
- Designed for CI/CD pipelines

## Installation

Download binaries from [GitHub Releases](https://github.com/migra/migra/releases) or install via Go:

```bash
go install github.com/migra/migra/cmd/migra@latest
```

Docker:

```bash
docker pull migra/migra:latest
docker run --rm -v $(pwd):/workspace migra/migra --help
```

## Quick Start

Create a `migra.yaml` configuration file:

```yaml
services:
  - name: api
    type: django
    path: ./services/api
    # No env needed - inherits from parent

  - name: billing
    type: laravel
    path: ./services/billing

execution:
  strategy: sequential
  stop_on_failure: true

logging:
  level: info
  format: console
```

Services inherit environment variables from the parent process. Set them once:

```bash
export DATABASE_URL="postgres://localhost/api_db"
export DB_HOST="localhost"
export DB_DATABASE="billing_db"

migra validate
migra deploy
```

## Auto-Discovery

Let Migra find services automatically:

```yaml
discovery:
  enabled: true
  root: ./services
  
execution:
  strategy: sequential
```

Migra scans for framework indicators (manage.py, artisan, prisma/*.prisma) and configures services automatically.

## Multi-Tenant Support

Configure multi-tenant deployments:

```yaml
services:
  - name: api
    type: django
    path: ./services/api
    env:
      DATABASE_URL: ${DATABASE_URL}
      DJANGO_SETTINGS_MODULE: api.settings

tenancy:
  enabled: true
  mode: database_per_tenant
  tenant_source: env
  stop_on_failure: false
  max_parallel: 10
```

Tenant sources:

Environment variable:
```bash
export MIGRA_TENANTS="tenant1:postgres://host/db1,tenant2:postgres://host/db2"
migra tenants deploy
```

JSON file:
```json
[
  {
    "id": "tenant1",
    "connection": {
      "DATABASE_URL": "postgres://host/tenant1_db"
    }
  },
  {
    "id": "tenant2",
    "connection": {
      "DATABASE_URL": "postgres://host/tenant2_db"
    }
  }
]
```

```bash
export MIGRA_TENANTS_FILE="tenants.json"
migra tenants deploy
```

External command:
```bash
export MIGRA_TENANTS_COMMAND="./scripts/get-tenants.sh"
migra tenants deploy --max-parallel 20
```

## Configuration Reference

### Services

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique service identifier |
| `type` | string | Yes | Framework type (django, laravel, prisma) |
| `path` | string | Yes | Service directory path |
| `env` | map | No | Environment variables |
| `working_dir` | string | No | Working directory (defaults to path) |

### Execution

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `strategy` | string | sequential | Execution strategy (sequential, parallel) |
| `stop_on_failure` | boolean | true | Stop on first failure |
| `parallel_limit` | integer | 5 | Max parallel executions |

### Tenancy

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | false | Enable multi-tenant mode |
| `mode` | string | database_per_tenant | Tenancy mode |
| `tenant_source` | string | - | Source type (env, file, command) |
| `stop_on_failure` | boolean | false | Stop on tenant failure |
| `max_parallel` | integer | 5 | Max parallel tenant executions |

### Logging

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `level` | string | info | Log level (debug, info, warn, error) |
| `format` | string | console | Output format (console, json) |
| `file` | string | - | Log file path (optional) |

## Framework Support

Django - executes `python manage.py migrate`

```yaml
- name: api
  type: django
  path: ./api
  env:
    DJANGO_SETTINGS_MODULE: api.settings
```

Laravel - executes `php artisan migrate --force`

```yaml
- name: web
  type: laravel
  path: ./web
  env:
    APP_ENV: production
```

Prisma - executes `npx prisma migrate deploy`

```yaml
- name: data
  type: prisma
  path: ./data
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install Migra
        run: |
          curl -sSL https://get.migra.dev | sh
      
      - name: Run Migrations
        run: migra deploy --verbose
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

### GitLab CI

```yaml
migrate:
  image: migra/migra:latest
  script:
    - migra validate
    - migra deploy
  only:
    - main
```

### Docker Compose

```yaml
version: '3.8'
services:
  migra:
    image: migra/migra:latest
    volumes:
      - ./migra.yaml:/workspace/migra.yaml
      - ./services:/workspace/services
    command: deploy
    environment:
      - DATABASE_URL=${DATABASE_URL}
```

## Examples

See the [examples/](examples/) directory for working configurations.

## Development

Requirements: Go 1.24+

Build from source:

```bash
git clone https://github.com/migra/migra.git
cd migra
make build
```

Run tests and create releases:

```bash
make test
make snapshot
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on adding framework adapters and submitting changes.

## Roadmap

Planned features include dependency graphs, schema drift detection, and advanced tenant batching.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- [Documentation](https://docs.migra.dev)
- [GitHub Issues](https://github.com/migra/migra/issues)
- [Discussions](https://github.com/migra/migra/discussions)
