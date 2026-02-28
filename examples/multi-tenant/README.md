# Multi-Tenant Example

Multi-tenant migration orchestration with database-per-tenant.

## Project Structure

```
saas-app/
├── migra.yaml
├── tenants.json
├── services/
│   ├── core/         # Core Django service
│   └── billing/      # Billing Laravel service
```

## Configuration

```yaml
services:
  - name: core
    type: django
    path: ./services/core

  - name: billing
    type: laravel
    path: ./services/billing

execution:
  strategy: sequential
  stop_on_failure: true

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

## Tenants File

Create `tenants.json`:

```json
[
  {
    "id": "acme-corp",
    "connection": {
      "DATABASE_URL": "postgres://localhost/acme_db",
      "DB_DATABASE": "acme_billing"
    }
  },
  {
    "id": "globex-inc",
    "connection": {
      "DATABASE_URL": "postgres://localhost/globex_db",
      "DB_DATABASE": "globex_billing"
    }
  },
  {
    "id": "initech-llc",
    "connection": {
      "DATABASE_URL": "postgres://localhost/initech_db",
      "DB_DATABASE": "initech_billing"
    }
  }
]
```

## Usage

Set tenant file and deploy:

```bash
export MIGRA_TENANTS_FILE="tenants.json"
migra tenants deploy
migra tenants deploy --max-parallel 20
migra tenants deploy --stop-on-failure
```

## Output

Typical output:

```
2024-01-01 12:00:00 [INFO] Loading tenants...
2024-01-01 12:00:00 [INFO] Found 3 tenants
2024-01-01 12:00:05 [INFO] Tenant acme-corp completed successfully
2024-01-01 12:00:06 [INFO] Tenant globex-inc completed successfully
2024-01-01 12:00:06 [INFO] Tenant initech-llc completed successfully
```

## Alternatives

Environment variable source:

```yaml
tenancy:
  tenant_source: env
```

```bash
export MIGRA_TENANTS="acme:postgres://localhost/acme_db,globex:postgres://localhost/globex_db"
migra tenants deploy
```

Command source:

```yaml
tenancy:
  tenant_source: command
```

Create script `get-tenants.sh`:

```bash
#!/bin/bash
curl -s https://api.example.com/tenants | jq -c '.'
```

Then:

```bash
export MIGRA_TENANTS_COMMAND="./get-tenants.sh"
migra tenants deploy
```

## Production Notes

- Use secrets manager for credentials
- Set reasonable parallelism limits
- Monitor with verbose logging
- Use `stop_on_failure: false` to process all tenants
- Schedule during maintenance windows
