# Configuration Guide

Reference for configuring Migra.

## Table of Contents

- [Configuration File](#configuration-file)
- [Services](#services)
- [Execution](#execution)
- [Tenancy](#tenancy)
- [Logging](#logging)
- [Environment Variables](#environment-variables)
- [Examples](#examples)

## Configuration File

Migra uses `migra.yaml` for configuration.

### Structure

```yaml
services:
  - name: service1
    type: framework
    path: ./path

execution:
  strategy: sequential
  stop_on_failure: true

tenancy:
  enabled: false

logging:
  level: info
  format: console

global_env:
  ENV_VAR: value

parallel_limit: 5
```

## Services

Define each microservice and its migration configuration.

### Required Fields

#### `name`

Unique service identifier.

```yaml
name: api-service
```

#### `type`

Framework type: `django`, `laravel`, or `prisma`.

```yaml
type: django
```

#### `path`

Path to service directory (relative or absolute).

```yaml
path: ./services/api
```

### Optional Fields

#### `env`

Service-specific environment variables (optional).

```yaml
env:
  DATABASE_URL: postgres://localhost/db
  DEBUG: "false"
```

#### `working_dir`

Working directory override (defaults to `path`).

```yaml
working_dir: ./services/api/src
```

### Example

```yaml
services:
  - name: api
    type: django
    path: ./services/api
    env:
      DATABASE_URL: postgres://localhost/api_db
      DJANGO_SETTINGS_MODULE: api.settings.production
    working_dir: ./services/api

  - name: web
    type: laravel
    path: ./services/web
    env:
      DB_CONNECTION: mysql
      DB_HOST: localhost
      DB_DATABASE: web_db
```

## Execution

Control migration execution strategy.

### `strategy`

`sequential` (default) or `parallel`.

```yaml
execution:
  strategy: sequential
```

### `stop_on_failure`

Stop on first failure (default: `true`).

```yaml
execution:
  stop_on_failure: true
```

### `parallel_limit`

Max concurrent executions for parallel strategy (default: 5).

```yaml
execution:
  strategy: parallel
  parallel_limit: 10
```

## Tenancy

Multi-tenant configuration (optional).

### `enabled`

Enable multi-tenant mode (default: `false`).

```yaml
tenancy:
  enabled: true
```

### `mode`

Tenancy model: `database_per_tenant` (default).

```yaml
tenancy:
  mode: database_per_tenant
```

### `tenant_source`

How to load tenants: `env`, `file`, or `command`.

```yaml
tenancy:
  tenant_source: env
```

Environment source:

```yaml
tenancy:
  tenant_source: env
```

Set environment variable:
```bash
export MIGRA_TENANTS="tenant1:postgres://host/db1,tenant2:postgres://host/db2"
```

File source:

```yaml
tenancy:
  tenant_source: file
```

Set file path:
```bash
export MIGRA_TENANTS_FILE="tenants.json"
```

File format (JSON):
```json
[
  {
    "id": "tenant1",
    "connection": {
      "DATABASE_URL": "postgres://host/tenant1_db"
    }
  }
]
```

File format (YAML):
```yaml
- id: tenant1
  connection:
    DATABASE_URL: postgres://host/tenant1_db
```

Command source:

```yaml
tenancy:
  tenant_source: command
```

Set command:
```bash
export MIGRA_TENANTS_COMMAND="./scripts/get-tenants.sh"
```

Command must output JSON array to stdout.

### `stop_on_failure`

Stop on first tenant failure (default: `false`).

```yaml
tenancy:
  stop_on_failure: false
```

### `max_parallel`

Max concurrent tenant executions (default: 5).

```yaml
tenancy:
  max_parallel: 20
```

## Logging

Control log output.

### `level`

Log level: `debug`, `info` (default), `warn`, or `error`.

```yaml
logging:
  level: info
```

### `format`

Output format: `console` (default) or `json`.

```yaml
logging:
  format: console
```

### `file`

Optional log file path.

```yaml
logging:
  file: /var/log/migra.log
```

## Environment Variables

Define variables for all services using `global_env`. Service-specific `env` overrides global values.

CLI environment variables:

- `MIGRA_TENANTS` - Tenant list for env source
- `MIGRA_TENANTS_FILE` - Tenant file path for file source
- `MIGRA_TENANTS_COMMAND` - Command for command source

## Examples

### Basic Sequential

```yaml
services:
  - name: api
    type: django
    path: ./api

  - name: web
    type: laravel
    path: ./web

execution:
  strategy: sequential
  stop_on_failure: true

logging:
  level: info
  format: console
```

### Parallel Execution

```yaml
services:
  - name: svc1
    type: django
    path: ./svc1

  - name: svc2
    type: django
    path: ./svc2

  - name: svc3
    type: laravel
    path: ./svc3

execution:
  strategy: parallel
  parallel_limit: 3
  stop_on_failure: false

logging:
  level: debug
  format: json
```

### Multi-Tenant

```yaml
services:
  - name: api
    type: django
    path: ./api

tenancy:
  enabled: true
  mode: database_per_tenant
  tenant_source: file
  stop_on_failure: false
  max_parallel: 10

logging:
  level: info
  format: console
```

### Production

```yaml
services:
  - name: api
    type: django
    path: /app/api
    env:
      DJANGO_SETTINGS_MODULE: api.settings.production
      DATABASE_URL: ${DATABASE_URL}

  - name: billing
    type: laravel
    path: /app/billing
    env:
      APP_ENV: production
      DB_CONNECTION: ${DB_CONNECTION}

execution:
  strategy: sequential
  stop_on_failure: true

logging:
  level: warn
  format: json
  file: /var/log/migra/migrations.log

global_env:
  APP_ENV: production
```

## Validation

Validate your configuration:

```bash
migra validate
```

Common errors:

- Missing required fields
- Invalid strategy or tenant source
- Duplicate service names
- Path doesn't exist

## Best Practices

- Use environment variables for sensitive data
- Version control `migra.yaml`
- Test with `--dry-run` before production
- Set reasonable parallel limits
- Monitor logs in production

## See Also

- [README](../README.md)
- [CONTRIBUTING](../CONTRIBUTING.md)
- [Examples](../examples/)
