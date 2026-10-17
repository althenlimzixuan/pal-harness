# Django Admin Portal Practices

## Tech Stack

| Concern | Library |
|---|---|
| Language | Python 3.14 |
| Framework | Django 5.2 LTS |
| Admin UI theme | `django-unfold` (Tailwind-based) |
| Database | MySQL via `mysqlclient` — shared with Go API |
| Cache + sessions | `django-redis` → same Redis instance as Go API (separate namespace) |
| Background tasks | `celery[redis]` + `django-celery-beat` (scheduled) |
| Config | `python-decouple` + `.env` |
| Auth / SSO | `mozilla-django-oidc` → Keycloak |
| Static files | `whitenoise` |
| Data import/export | `django-import-export` |
| Linting | `ruff` + `black` |
| Type checking | `mypy` + `django-stubs` (strict, enforced in CI) |
| Testing | Django `TestCase` (built-in) + JSON fixtures |

## requirements.txt Structure

```
# requirements.txt — production
Django==5.2.*
mysqlclient
django-redis
celery[redis]
django-celery-beat
django-unfold
mozilla-django-oidc
python-decouple
whitenoise
django-import-export
mysqlclient

# requirements-dev.txt
-r requirements.txt
ruff
black
mypy
django-stubs
django-extensions   # shell_plus
```

## Project Structure

```
manage.py
requirements.txt
requirements-dev.txt
config/
  settings/
    base.py           # shared settings
    development.py    # DEBUG=True, relaxed security
    production.py     # security headers, HSTS, strict
  urls.py
  wsgi.py
  celery.py           # Celery app initialisation
apps/
  core/               # custom AdminSite, base model, shared utils
  config_mgmt/        # SystemConfig model — runtime config read by Go API
  audit/              # AuditLog model + admin integration
  [domain]/           # one app per domain area (e.g. trips/, users/)
templates/
  admin/              # Unfold + Django admin template overrides
static/
tests/
  unit_test/          # JSON fixtures: tests/unit_test/{use_case}.json
  integration_test/   # JSON fixtures: tests/integration_test/{use_case}.json
```

## Shared Database — Critical Rule

Django shares the MySQL instance with the Go API.
**Go API owns all core domain table schemas. Django must never create or alter them.**

```python
# apps/users/models.py — read-only mirror of Go API table
class User(models.Model):
    """Read-only view of Go API users table."""
    id = models.CharField(max_length=36, primary_key=True)
    email = models.EmailField()
    created_at = models.DateTimeField()

    class Meta:
        managed = False       # Django will NOT run migrations on this table
        db_table = 'users'    # must match exact Go API table name
```

Rules:
- All Go API domain tables: `managed = False`
- All Django-owned tables (config, audit, auth, celery): `managed = True` (default)
- Never run `makemigrations` for unmanaged models
- Django migrations only ever touch Django-owned tables

## Admin UI Theme (Unfold)

```python
# config/settings/base.py
INSTALLED_APPS = [
    'unfold',
    'unfold.contrib.filters',
    'unfold.contrib.import_export',
    'django.contrib.admin',
    ...
]

UNFOLD = {
    'SITE_TITLE': 'ETP Admin',
    'SITE_HEADER': 'Essential Traveller Planner',
    'SIDEBAR': {
        'show_search': True,
        'show_all_applications': False,
        'navigation': [...],   # define app groupings here
    },
}
```

- Register all `ModelAdmin` with `@admin.register(Model)` — import `ModelAdmin` from `unfold.admin`, not `django.contrib.admin`
- Unfold extends Django admin — all standard `ModelAdmin` attributes still apply

## ModelAdmin Conventions

```python
from unfold.admin import ModelAdmin

@admin.register(Trip, site=admin_site)
class TripAdmin(ModelAdmin):
    list_display = ['id', 'title', 'user_email', 'status', 'created_at']
    list_filter = ['status', 'created_at']
    search_fields = ['title', 'user__email']
    readonly_fields = ['id', 'created_at', 'updated_at']
    ordering = ['-created_at']

    @admin.display(description='User Email')
    def user_email(self, obj: Trip) -> str:
        return obj.user.email
```

- `list_display` always includes `id`, primary identifier, `created_at`
- `readonly_fields` always includes `id`, `created_at`, `updated_at`
- All fields on `managed = False` models are `readonly_fields` — never editable
- `list_filter` on status/type fields and date ranges
- Type annotations on all `ModelAdmin` methods

## Runtime Config (SystemConfig)

Django-owned table. Go API reads it at startup and caches in Redis with short TTL:

```python
# apps/config_mgmt/models.py
class SystemConfig(models.Model):
    key = models.CharField(max_length=100, unique=True)
    value = models.TextField()
    description = models.TextField(blank=True)
    updated_by = models.ForeignKey(
        settings.AUTH_USER_MODEL, on_delete=models.SET_NULL, null=True
    )
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        db_table = 'system_config'
        verbose_name = 'System Config'
        verbose_name_plural = 'System Config'
```

Managed keys include: `JWT_ACCESS_TOKEN_TTL` (default `15m`), feature flags, rate limit thresholds.
After any update, a Celery task invalidates the Redis cache key so Go API picks up the new value immediately.

## Celery

```python
# config/celery.py
from celery import Celery
app = Celery('etp_admin')
app.config_from_object('django.conf:settings', namespace='CELERY')
app.autodiscover_tasks()
```

```python
# config/settings/base.py
CELERY_BROKER_URL = config('REDIS_URL') + '/1'   # DB 1, separate from Go API's DB 0
CELERY_RESULT_BACKEND = config('REDIS_URL') + '/1'
CELERY_BEAT_SCHEDULER = 'django_celery_beat.schedulers:DatabaseScheduler'
```

- Redis DB namespace separation: Go API uses DB 0, Django/Celery uses DB 1
- `django-celery-beat` stores schedules in DB — manage periodic tasks from admin UI
- Use `@shared_task` decorator, not `@app.task`
- Celery worker runs as a separate process/container alongside Django

## Authentication (Keycloak OIDC)

Staff-only portal. All auth delegated to Keycloak:

```python
AUTHENTICATION_BACKENDS = ['mozilla_django_oidc.auth.OIDCAuthenticationBackend']
OIDC_RP_CLIENT_ID = config('OIDC_RP_CLIENT_ID')
OIDC_RP_CLIENT_SECRET = config('OIDC_RP_CLIENT_SECRET')
OIDC_OP_AUTHORIZATION_ENDPOINT = config('OIDC_OP_AUTHORIZATION_ENDPOINT')
OIDC_OP_TOKEN_ENDPOINT = config('OIDC_OP_TOKEN_ENDPOINT')
OIDC_OP_USER_ENDPOINT = config('OIDC_OP_USER_ENDPOINT')
OIDC_CREATE_USER = False   # only allow pre-existing staff accounts
```

- `is_staff = True` mapped from Keycloak `admin` role
- `is_superuser` mapped from Keycloak `superadmin` role
- No `createsuperuser` in production — all access via Keycloak
- MFA handled by Keycloak — no Django MFA package needed

## Configuration via ENV

```python
from decouple import config, Csv

SECRET_KEY = config('SECRET_KEY')
DEBUG = config('DEBUG', default=False, cast=bool)
ALLOWED_HOSTS = config('ALLOWED_HOSTS', cast=Csv())
DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.mysql',
        'NAME': config('DB_NAME'),
        'USER': config('DB_USER'),
        'PASSWORD': config('DB_PASSWORD'),
        'HOST': config('DB_HOST'),
        'PORT': config('DB_PORT', default='3306'),
        'OPTIONS': {'charset': 'utf8mb4'},
    }
}
CACHES = {
    'default': {
        'BACKEND': 'django_redis.cache.RedisCache',
        'LOCATION': config('REDIS_URL') + '/1',
    }
}
SESSION_ENGINE = 'django.contrib.sessions.backends.cache'
SESSION_CACHE_ALIAS = 'default'
```

Never use `os.environ.get` inline — always `config()` from `python-decouple`.

## Make Targets

```bash
make dev            # python manage.py runserver
make shell          # python manage.py shell_plus
make migrate        # python manage.py migrate
make migrations     # python manage.py makemigrations
make worker         # celery -A config worker -l info
make beat           # celery -A config beat -l info
make test           # python manage.py test
make coverage       # coverage run manage.py test && coverage report, gate: 80%
make lint           # ruff check . && black --check .
make format         # black . && ruff check --fix .
make typecheck      # mypy .
make static         # python manage.py collectstatic --noinput
```

## Testing Conventions

Django built-in `TestCase` with JSON fixtures at `tests/{test_type}/{use_case}.json`:

```python
from django.test import TestCase

class SystemConfigTestCase(TestCase):
    fixtures = ['tests/unit_test/system_config.json']

    def test_jwt_ttl_default(self) -> None:
        from apps.config_mgmt.models import SystemConfig
        config = SystemConfig.objects.get(key='JWT_ACCESS_TOKEN_TTL')
        self.assertEqual(config.value, '15m')
```

- Use `self.client` (Django test client) for admin view tests
- `TestCase` wraps each test in a transaction and rolls back — no test pollution
- Integration tests use `TransactionTestCase` or `LiveServerTestCase` when needed
- Never test Django admin rendering — test your custom `ModelAdmin` methods and Celery tasks

## Audit Logging

```python
# apps/audit/models.py
class AuditLog(models.Model):
    user = models.ForeignKey(settings.AUTH_USER_MODEL, on_delete=models.SET_NULL, null=True)
    action = models.CharField(max_length=100)
    target_model = models.CharField(max_length=100)
    target_id = models.CharField(max_length=100)
    diff = models.JSONField(default=dict)
    ip_address = models.GenericIPAddressField(null=True)
    timestamp = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['-timestamp']
```

Override `save_model` and `delete_model` in sensitive `ModelAdmin` classes to write `AuditLog` entries.

## Type Checking (mypy)

```ini
# mypy.ini
[mypy]
plugins = mypy_django_plugin.main
strict = true
ignore_missing_imports = true

[mypy.plugins.django-stubs]
django_settings_module = config.settings.development
```

- All function signatures have type annotations
- No `Any` except where third-party libraries force it
- `mypy` runs in CI — failures block merge

## Security Defaults

- `DEBUG = False` in production
- `ALLOWED_HOSTS` explicit — never `['*']`
- `CSRF_COOKIE_SECURE = True`, `SESSION_COOKIE_SECURE = True` in production
- `SECURE_HSTS_SECONDS = 31536000` in production
- Admin accessible only from internal network / VPN
- No superuser created via CLI in production — Keycloak only

## Code Conventions

- One Django app per domain area — not one monolithic `admin` app
- All apps under `apps/` directory
- `__str__` defined on every model
- `verbose_name` and `verbose_name_plural` in every model's `Meta`
- Custom admin views decorated with `@staff_member_required`
- Type annotations on all function signatures
