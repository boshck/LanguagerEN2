# Настройка сервера для деплоя

Краткая инструкция по подготовке сервера для автоматического деплоя.

## Требования на сервере

- Docker и Docker Compose установлены
- SSH доступ для пользователя deploy
- Проект в директории `/opt/LanguagerEN2`

## Пошаговая настройка

### 1. Создай пользователя deploy

```bash
# На сервере
sudo useradd -m -s /bin/bash deploy
sudo usermod -aG docker deploy
```

### 2. Создай директорию проекта

```bash
sudo mkdir -p /opt/LanguagerEN2
sudo chown deploy:deploy /opt/LanguagerEN2
```

### 3. Настрой SSH ключ

На локальной машине:

```bash
# Генерация SSH ключа
ssh-keygen -t ed25519 -C "gitlab-deploy" -f ~/.ssh/gitlab_deploy

# Копирование на сервер
ssh-copy-id -i ~/.ssh/gitlab_deploy.pub deploy@your-server.com
```

На сервере:

```bash
# Проверка ключа
cat ~/.ssh/authorized_keys
```

### 4. Скопируй проект на сервер

```bash
# Локально
cd /path/to/LanguagerEN2
rsync -avz --exclude '.git' --exclude 'backups' \
  ./ deploy@your-server.com:/opt/LanguagerEN2/
```

### 5. Создай .env на сервере

```bash
# На сервере
cd /opt/LanguagerEN2
cp .env.example .env
nano .env  # Заполни BOT_TOKEN, BOT_PASSWORD, DB_PASSWORD
```

### 6. Настрой GitLab Variables

GitLab → Settings → CI/CD → Variables:

```bash
SSH_PRIVATE_KEY     # Содержимое ~/.ssh/gitlab_deploy (приватный ключ)
SERVER_HOST         # IP или домен сервера
SERVER_USER         # deploy
DEPLOY_PATH         # /opt/LanguagerEN2
BOT_TOKEN           # Токен бота
BOT_PASSWORD        # Пароль бота
DB_PASSWORD         # Пароль PostgreSQL
```

### 7. Проверка настройки

На сервере:

```bash
# Проверь Docker
docker --version
docker-compose --version

# Проверь права
cd /opt/LanguagerEN2
ls -la

# Проверь что .env настроен
cat .env | grep BOT_TOKEN
```

### 8. Первый деплой

Вручную на сервере:

```bash
cd /opt/LanguagerEN2
docker-compose up -d --build
docker-compose logs -f bot
```

Если всё работает - можно использовать CI/CD!

## Структура на сервере

```
/opt/LanguagerEN2/
├── docker-compose.yml
├── .env
├── backups/
│   ├── deployments/       # Бекапы docker-compose
│   └── backup_*.sql       # Бекапы БД
├── scripts/
│   ├── deploy.sh
│   └── health_check.sh
└── migrations/
```

## Troubleshooting

### Permission denied

```bash
sudo chown -R deploy:deploy /opt/LanguagerEN2
sudo chmod +x /opt/LanguagerEN2/scripts/*.sh
```

### Docker permission

```bash
sudo usermod -aG docker deploy
# Перелогинься
```

### SSH connection refused

```bash
# На сервере
sudo systemctl status sshd
sudo ufw allow 22/tcp  # Если используется UFW
```

## Безопасность

1. **SSH ключи** - используй только SSH ключи, отключи password auth
2. **Firewall** - открой только нужные порты (22, 80, 443)
3. **Docker socket** - доступ только для пользователя deploy
4. **Secrets** - храни в GitLab Variables, никогда не коммить

## Полезные команды на сервере

```bash
# Логи
cd /opt/LanguagerEN2
docker-compose logs -f bot

# Перезапуск
docker-compose restart bot

# Статус
docker-compose ps

# Бекапы
ls -lah backups/

# Health check
./scripts/health_check.sh
```

