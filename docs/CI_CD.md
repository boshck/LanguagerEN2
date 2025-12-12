# CI/CD с GitLab

Автоматическое тестирование, сборка и деплой Languager Bot через GitLab CI/CD.

## Обзор Pipeline

```
Push to 'private' → Test → Build Docker → Push to Registry → Deploy to Server
                      ↓
                   Failed? Stop
```

### Стадии

1. **Test** - запуск unit тестов с проверкой coverage
2. **Build** - сборка Docker образа
3. **Deploy** - деплой на сервер через SSH (manual trigger)

## Настройка GitLab

### 1. Создать ветку `private`

```bash
git checkout -b private
git push -u origin private
```

### 2. Настроить GitLab Variables

Перейди в GitLab: **Settings → CI/CD → Variables**

Добавь следующие переменные:

| Variable | Value | Protected | Masked | Description |
|----------|-------|-----------|--------|-------------|
| `SSH_PRIVATE_KEY` | `-----BEGIN...` | ✅ | ✅ | Приватный SSH ключ для доступа к серверу |
| `SERVER_HOST` | `your-server.com` | ✅ | ❌ | IP или домен сервера |
| `SERVER_USER` | `deploy` | ✅ | ❌ | Пользователь для SSH |
| `DEPLOY_PATH` | `/opt/languager` | ✅ | ❌ | Путь к проекту на сервере |
| `BOT_TOKEN` | `123456:ABC...` | ✅ | ✅ | Токен Telegram бота |
| `BOT_PASSWORD` | `secret123` | ✅ | ✅ | Пароль доступа к боту |
| `DB_PASSWORD` | `db_pass` | ✅ | ✅ | Пароль PostgreSQL |

### 3. Настроить SSH ключ

На локальной машине:

```bash
# Генерация SSH ключа (если нет)
ssh-keygen -t ed25519 -C "gitlab-deploy" -f ~/.ssh/gitlab_deploy

# Скопируй ПРИВАТНЫЙ ключ
cat ~/.ssh/gitlab_deploy
# Вставь в GitLab Variable SSH_PRIVATE_KEY

# Скопируй ПУБЛИЧНЫЙ ключ на сервер
ssh-copy-id -i ~/.ssh/gitlab_deploy.pub deploy@your-server.com
```

На сервере:

```bash
# Проверь что ключ добавлен
cat ~/.ssh/authorized_keys
```

## Pipeline Stages

### Stage 1: Test

**Что происходит:**
1. Устанавливается Go 1.21
2. Скачиваются зависимости
3. Запускаются тесты с race detector
4. Генерируется coverage отчёт
5. Проверяется минимальное покрытие (80%)

**Если тесты падают:** Pipeline останавливается, деплой не происходит ❌

**Логи:**

```bash
# В GitLab: CI/CD → Pipelines → Jobs → test
# Или локально повтори:
make test-ci
```

### Stage 2: Build

**Что происходит:**
1. Собирается Docker образ
2. Образ тегируется двумя тегами:
   - `latest` - всегда последняя версия
   - `$CI_COMMIT_SHORT_SHA` - конкретный коммит (для rollback)
3. Push в GitLab Container Registry

**Образ доступен:**
```
registry.gitlab.com/your-username/languager:latest
registry.gitlab.com/your-username/languager:abc123f
```

### Stage 3: Deploy

**⚠️ Manual trigger** - требует ручного подтверждения в GitLab UI.

**Что происходит:**
1. Подключается к серверу по SSH
2. Копирует скрипты деплоя на сервер
3. Логинится в GitLab Container Registry
4. Пуллит свежий Docker образ
5. Останавливает старые контейнеры
6. Запускает новые контейнеры
7. Запускает health check
8. При ошибках - rollback к предыдущей версии

**Ручной запуск деплоя:**

1. Перейди в GitLab: **CI/CD → Pipelines**
2. Найди нужный pipeline
3. Нажми на play ▶️ возле `deploy` job

## Структура .gitlab-ci.yml

```yaml
stages:
  - test       # Автоматически
  - build      # Автоматически
  - deploy     # Вручную (when: manual)

# Запускается только для ветки 'private'
workflow:
  rules:
    - if: $CI_COMMIT_BRANCH == "private"
```

## Скрипты деплоя

### deploy.sh

Основной скрипт деплоя:

```bash
#!/bin/bash
# 1. Login to registry
# 2. Pull latest image
# 3. Backup current state
# 4. Stop old containers
# 5. Start new containers
# 6. Run health check
# 7. Rollback if failed
```

### health_check.sh

Проверка здоровья после деплоя:

```bash
#!/bin/bash
# 1. Check container is running
# 2. Check container status
# 3. Check logs for errors
# 4. Check bot started message
# 5. Check PostgreSQL is running
```

## Workflow

### Обычный деплой

```bash
# 1. Разработка
git checkout private
# ... изменения кода ...

# 2. Коммит
git add .
git commit -m "feat: add new feature"

# 3. Push (запустит pipeline)
git push origin private

# 4. Проверь в GitLab что тесты прошли
# Settings → CI/CD → Pipelines

# 5. Запусти деплой вручную
# Нажми ▶️ возле deploy job
```

### Проверка деплоя

```bash
# На сервере
ssh deploy@your-server.com

# Проверить что контейнеры работают
cd /opt/languager
docker-compose ps

# Посмотреть логи
docker-compose logs -f bot

# Проверить здоровье
./scripts/health_check.sh
```

## Rollback (откат)

### Вариант 1: Через GitLab (рекомендуется)

1. Перейди в **CI/CD → Pipelines**
2. Найди **предыдущий успешный** pipeline
3. Нажми **Retry** на deploy job
4. Подтверди деплой

### Вариант 2: Вручную на сервере

```bash
ssh deploy@your-server.com
cd /opt/languager

# Посмотреть доступные образы
docker images | grep languager

# Откатиться на конкретный коммит
docker pull registry.gitlab.com/username/languager:abc123f
docker-compose down bot
# Измени docker-compose.yml или export переменную
export DOCKER_IMAGE=registry.gitlab.com/username/languager:abc123f
docker-compose up -d bot

# Проверь
./scripts/health_check.sh
```

### Вариант 3: Через бекап

```bash
cd /opt/languager
ls backups/deployments/

# Восстанови предыдущий docker-compose.yml
cp backups/deployments/docker-compose.20241212_143022.yml.bak docker-compose.yml

# Перезапусти
docker-compose up -d bot
```

## Мониторинг

### GitLab Pipeline Status

Добавь badge в README:

```markdown
[![Pipeline](https://gitlab.com/username/languager/badges/private/pipeline.svg)](https://gitlab.com/username/languager/pipelines)
```

### Coverage Badge

```markdown
[![Coverage](https://gitlab.com/username/languager/badges/private/coverage.svg)](https://gitlab.com/username/languager/-/graphs/private/charts)
```

### Логи в GitLab

- **CI/CD → Pipelines** - список всех pipeline
- Клик на pipeline → список jobs
- Клик на job → логи выполнения

### Логи на сервере

```bash
# Все логи
docker-compose logs -f

# Только бот
docker-compose logs -f bot

# Последние 100 строк
docker-compose logs --tail=100 bot
```

## Troubleshooting

### Pipeline не запускается

**Проблема:** Push не триггерит pipeline

**Решение:**
1. Проверь что push в ветку `private`
2. Проверь `.gitlab-ci.yml` синтаксис: **CI/CD → CI Lint**
3. Проверь что CI/CD включен: **Settings → General → Visibility**

### Тесты падают в CI

**Проблема:** Тесты проходят локально, но падают в CI

**Решение:**
```bash
# Запусти локально как в CI
make test-ci

# Проверь race conditions
go test -race ./...

# Очисть кеш
go clean -testcache
go test ./...
```

### Deploy failed: SSH connection

**Проблема:** `Permission denied (publickey)`

**Решение:**
1. Проверь что SSH_PRIVATE_KEY правильно добавлен в Variables
2. Проверь что публичный ключ на сервере:
   ```bash
   ssh deploy@your-server.com
   cat ~/.ssh/authorized_keys
   ```
3. Проверь права:
   ```bash
   chmod 700 ~/.ssh
   chmod 600 ~/.ssh/authorized_keys
   ```

### Deploy failed: Docker pull

**Проблема:** `unauthorized: authentication required`

**Решение:**

На сервере создай deploy token:

GitLab: **Settings → Repository → Deploy Tokens**

```
Name: server-deploy
Scopes: ✅ read_registry
```

На сервере:

```bash
docker login registry.gitlab.com -u <token-name> -p <token>
```

### Health check failed

**Проблема:** Контейнер запустился, но health check не проходит

**Решение:**
```bash
# На сервере
ssh deploy@your-server.com
cd /opt/languager

# Проверь логи
docker-compose logs --tail=50 bot

# Проверь что БД работает
docker-compose ps postgres

# Ручной health check
./scripts/health_check.sh
```

### Rollback не работает

**Решение:**

```bash
# Используй бекап compose файла
cd /opt/languager
ls -lah backups/deployments/

# Восстанови последний рабочий
cp backups/deployments/docker-compose.<timestamp>.yml.bak docker-compose.yml

# Перезапусти
docker-compose up -d bot
```

## Оптимизация Pipeline

### Кеширование

Go модули кешируются между runs:

```yaml
cache:
  paths:
    - .go/pkg/mod/
```

### Параллельный запуск

Можно разделить тесты на несколько jobs:

```yaml
test-service:
  script:
    - go test ./internal/service/... -v

test-repository:
  script:
    - go test ./internal/repository/... -v
```

### Docker Layer Caching

Используется автоматически в GitLab Docker executor.

## Best Practices

### 1. Всегда проверяй тесты локально

```bash
make test-ci  # Перед push
```

### 2. Деплой только вручную (manual)

Для production всегда требуй ручного подтверждения:

```yaml
deploy:
  when: manual
```

### 3. Используй коммит-хеши для rollback

Тегируй образы по коммитам:

```yaml
$CI_REGISTRY_IMAGE:$CI_COMMIT_SHORT_SHA
```

### 4. Мониторинг после деплоя

После каждого деплоя:
1. Проверь health check
2. Посмотри логи (2-3 минуты)
3. Проверь что бот отвечает в Telegram

### 5. Бекапы перед деплоем

Всегда делается автоматически в `deploy.sh`:

```bash
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
cp docker-compose.yml backups/deployments/docker-compose.${TIMESTAMP}.yml.bak
```

## Дополнительные ресурсы

- [GitLab CI/CD Documentation](https://docs.gitlab.com/ee/ci/)
- [GitLab Container Registry](https://docs.gitlab.com/ee/user/packages/container_registry/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)

