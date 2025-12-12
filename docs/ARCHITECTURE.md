# Архитектура Languager Bot

## Обзор

Languager - это Telegram бот для изучения слов, построенный на принципах Clean Architecture для простого масштабирования и поддержки.

## Архитектурные слои

```
┌─────────────────────────────────────┐
│      Telegram API (External)        │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│     Handler Layer (Presentation)    │
│  - start.go                         │
│  - word.go                          │
│  - callbacks.go                     │
│  - State Machine управление         │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│    Service Layer (Business Logic)   │
│  - auth.go (авторизация)            │
│  - word.go (работа со словами)      │
│  - stats.go (статистика и cleanup)  │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│    Repository Layer (Data Access)   │
│  - user.go (пользователи)           │
│  - word.go (слова)                  │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│        PostgreSQL Database          │
│  - users table                      │
│  - words table                      │
└─────────────────────────────────────┘
```

## Структура проекта

```
LanguagerEN2/
├── cmd/bot/
│   └── main.go                 # Точка входа приложения
│
├── internal/                   # Приватный код приложения
│   ├── config/
│   │   └── config.go          # Загрузка конфигурации
│   │
│   ├── domain/                # Модели данных (entities)
│   │   ├── user.go           # Модель пользователя
│   │   ├── word.go           # Модель слова
│   │   └── day.go            # Модель дня со словами
│   │
│   ├── repository/           # Слой работы с БД
│   │   ├── repository.go    # Интерфейсы
│   │   └── postgres/
│   │       ├── user.go      # Реализация для users
│   │       └── word.go      # Реализация для words
│   │
│   ├── service/             # Бизнес-логика
│   │   ├── auth.go         # Авторизация
│   │   ├── word.go         # Управление словами
│   │   └── stats.go        # Статистика
│   │
│   ├── handler/            # Обработчики Telegram
│   │   ├── handler.go     # Базовая структура
│   │   ├── start.go       # /start команда
│   │   ├── word.go        # Работа со словами
│   │   └── callbacks.go   # Inline кнопки
│   │
│   └── middleware/
│       └── auth.go        # Middleware авторизации
│
├── migrations/            # SQL миграции
│   ├── 001_init.sql      # Создание таблиц
│   └── 002_add_indexes.sql
│
├── scripts/
│   └── backup.sh         # Автоматический бекап
│
├── docker-compose.yml    # Оркестрация сервисов
├── Dockerfile           # Сборка бота
└── Makefile            # Удобные команды
```

## Слои и их ответственность

### 1. Handler Layer (Presentation)

**Ответственность:**
- Получение сообщений от Telegram API
- Валидация входных данных
- Управление состоянием пользователя (State Machine)
- Форматирование ответов

**Не делает:**
- Бизнес-логику
- Прямые запросы к БД

**State Machine:**
```
StateIdle → StateWaitingTranslation → StateIdle
                ↓
              (save word pair)
```

### 2. Service Layer (Business Logic)

**Ответственность:**
- Бизнес-правила (например, валидация слов)
- Оркестрация между несколькими репозиториями
- Вычисления и трансформации данных

**Не делает:**
- Работу с Telegram API
- SQL запросы напрямую

**Пример:**
```go
func (s *WordService) SaveWordPair(userID int64, word, translation string) error {
    // Бизнес-правило: слова не должны быть пустыми
    if word == "" || translation == "" {
        return fmt.Errorf("word and translation cannot be empty")
    }
    
    // Делегируем сохранение репозиторию
    return s.wordRepo.SaveWord(userID, word, translation)
}
```

### 3. Repository Layer (Data Access)

**Ответственность:**
- SQL запросы
- Маппинг между БД и domain моделями
- Оптимизация запросов

**Не делает:**
- Бизнес-логику
- Работу с внешними API

## База данных

### Схема

```sql
-- Пользователи
CREATE TABLE users (
    user_id BIGINT PRIMARY KEY,
    authorized BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Слова
CREATE TABLE words (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    word TEXT NOT NULL,
    translation TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Индексы
CREATE INDEX idx_words_user_id ON words(user_id);
CREATE INDEX idx_words_created_at ON words(created_at);
CREATE INDEX idx_words_user_date ON words(user_id, created_at DESC);
```

### Особенности

- **Автоматическая очистка**: Слова старше 60 дней удаляются автоматически
- **Каскадное удаление**: При удалении пользователя удаляются его слова
- **Оптимизированные индексы**: Для быстрых запросов по пользователю и дате

## Как добавить новую функцию

### Пример: Добавить экспорт слов в CSV

#### 1. Добавить метод в Repository

```go
// internal/repository/repository.go
type WordRepository interface {
    // ... existing methods
    GetAllWords(userID int64) ([]domain.Word, error)
}

// internal/repository/postgres/word.go
func (r *WordRepo) GetAllWords(userID int64) ([]domain.Word, error) {
    query := `SELECT id, user_id, word, translation, created_at
              FROM words WHERE user_id = $1 ORDER BY created_at DESC`
    // ... implementation
}
```

#### 2. Добавить метод в Service

```go
// internal/service/word.go
func (s *WordService) ExportToCSV(userID int64) (string, error) {
    words, err := s.wordRepo.GetAllWords(userID)
    if err != nil {
        return "", err
    }
    
    // Convert to CSV format
    var csv strings.Builder
    csv.WriteString("Word,Translation,Date\n")
    for _, w := range words {
        csv.WriteString(fmt.Sprintf("%s,%s,%s\n", 
            w.Word, w.Translation, w.CreatedAt.Format("2006-01-02")))
    }
    
    return csv.String(), nil
}
```

#### 3. Добавить Handler

```go
// internal/handler/callbacks.go
var btnExport = tele.Btn{Unique: "export", Text: "📥 Экспорт"}

func (h *Handler) handleExport(c tele.Context) error {
    csv, err := h.wordService.ExportToCSV(c.Sender().ID)
    if err != nil {
        return c.Respond(&tele.CallbackResponse{Text: "Ошибка экспорта"})
    }
    
    return c.SendDocument(&tele.Document{
        File: tele.FromReader(strings.NewReader(csv)),
        FileName: "words.csv",
    })
}
```

## Масштабирование

### Горизонтальное масштабирование бота

Для запуска нескольких инстансов бота:

1. **Добавить Redis для состояний:**
   ```go
   // Вместо in-memory map использовать Redis
   type RedisStateStore struct {
       client *redis.Client
   }
   ```

2. **Настроить load balancer** перед PostgreSQL

3. **Разделить обработку:**
   - Bot instance 1: обработка сообщений
   - Bot instance 2: обработка callback queries
   - Worker: фоновые задачи (cleanup, уведомления)

### Добавление новых фич

**REST API для веб-версии:**
```
internal/
  └── handler/
      ├── telegram/     # Telegram handlers
      └── http/         # HTTP REST handlers
          ├── auth.go
          └── words.go
```

**Добавление других платформ (VK, Discord):**
```
internal/
  └── handler/
      ├── telegram/
      ├── vk/
      └── discord/
```

Все они используют одни и те же `service` и `repository` слои!

## Безопасность

- **Пароль** хранится в environment variables, не в коде
- **PostgreSQL пароль** только в .env файле
- **Токен бота** только в .env файле
- **.gitignore** настроен чтобы не коммитить секреты

## Мониторинг и логирование

- **Структурированное логирование** через zap (JSON формат)
- **Graceful shutdown** для корректного завершения операций
- **Database connection pooling** для оптимальной производительности
- **Автоматические бекапы** каждые 24 часа

## Тестирование

Архитектура позволяет легко писать тесты:

```go
// Мокаем repository для тестирования service
type MockWordRepo struct {
    mock.Mock
}

func (m *MockWordRepo) SaveWord(userID int64, word, translation string) error {
    args := m.Called(userID, word, translation)
    return args.Error(0)
}

// Тестируем service
func TestWordService_SaveWordPair(t *testing.T) {
    mockRepo := new(MockWordRepo)
    mockRepo.On("SaveWord", int64(123), "hello", "привет").Return(nil)
    
    service := NewWordService(mockRepo)
    err := service.SaveWordPair(123, "hello", "привет")
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

## Troubleshooting

### Проблема: Бот не отвечает

1. Проверить логи: `make logs`
2. Проверить подключение к БД: `make db`
3. Проверить токен в .env

### Проблема: Миграции не применяются

```bash
# Зайти в контейнер
docker-compose exec bot sh

# Проверить миграции вручную
cd /app/migrations
ls -la
```

### Проблема: Бекапы не создаются

```bash
# Проверить логи бекапа
docker-compose logs backup

# Проверить директорию
ls -lah backups/
```

## Полезные команды

```bash
# Посмотреть логи
make logs

# Перезапустить только бота
make restart

# Подключиться к БД
make db

# Создать ручной бекап
make backup

# Очистить всё (включая volumes)
make clean
```

