# Auto-Discovery Example

Minimal configuration using service auto-discovery.

## Project Structure

```
project/
├── migra.yaml
└── services/
    ├── api/          # Django - has manage.py
    ├── web/          # Laravel - has artisan
    └── analytics/    # Prisma - has prisma/*.prisma
```

## Configuration

Migra automatically finds services by scanning for framework indicators:
- Django: `manage.py`
- Laravel: `artisan`
- Prisma: `prisma/*.prisma` files

No need to list services explicitly. Just point to the root directory.

## Usage

Services inherit environment from parent process:

```bash
export DATABASE_URL="postgres://localhost/db"
export DB_HOST="localhost"

migra validate  # Shows discovered services
migra deploy
```

## Output

```
2024-01-01 12:00:00 [INFO] Discovered 3 services
2024-01-01 12:00:00 [INFO] - api (django) at ./services/api
2024-01-01 12:00:00 [INFO] - web (laravel) at ./services/web
2024-01-01 12:00:00 [INFO] - analytics (prisma) at ./services/analytics
2024-01-01 12:00:00 [INFO] Starting migration deployment
```

## Override Discovered Services

You can still explicitly define services to override discovery:

```yaml
discovery:
  enabled: true
  root: ./services

services:
  - name: api
    type: django
    path: ./services/api
    env:
      DJANGO_SETTINGS_MODULE: api.settings.production
```

Explicit services take precedence over discovered ones.
