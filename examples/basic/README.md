# Basic Example

Basic multi-service setup with Django and Laravel.

## Project Structure

```
project/
├── migra.yaml
├── services/
│   ├── api/          # Django service
│   │   ├── manage.py
│   │   └── migrations/
│   └── web/          # Laravel service
│       ├── artisan
│       └── database/migrations/
```

## Configuration

```yaml
services:
  - name: api
    type: django
    path: ./services/api
    env:
      DATABASE_URL: postgres://localhost:5432/api_db
      DJANGO_SETTINGS_MODULE: api.settings

  - name: web
    type: laravel
    path: ./services/web
    env:
      DB_CONNECTION: mysql
      DB_HOST: localhost
      DB_DATABASE: web_db
      DB_USERNAME: root
      DB_PASSWORD: secret

execution:
  strategy: sequential
  stop_on_failure: true

logging:
  level: info
  format: console
```

## Usage

Services inherit environment from parent. Set variables once:

```bash
export DATABASE_URL="postgres://localhost:5432/api_db"
export DB_HOST="localhost"
export DB_DATABASE="web_db"
export DB_USERNAME="myuser"
export DB_PASSWORD="mypassword"

migra validate
migra deploy
```

## Output

Typical output:

```
2024-01-01 12:00:00 [INFO] Starting migration deployment
2024-01-01 12:00:00 [INFO] Executing migrations for 2 service(s)
2024-01-01 12:00:02 [INFO] Service api completed successfully
2024-01-01 12:00:04 [INFO] Service web completed successfully
```
