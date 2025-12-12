# Тестирование Languager Bot

Полное руководство по написанию и запуску тестов.

## Быстрый старт

```bash
# Запустить все тесты
make test

# С покрытием кода
make test-coverage

# Открыть HTML отчёт о покрытии
open coverage.html  # или xdg-open на Linux
```

## Структура тестов

```
internal/
├── service/
│   ├── auth.go
│   ├── auth_test.go        # ✅ Тесты сервиса авторизации
│   ├── word.go
│   ├── word_test.go        # ✅ Тесты работы со словами
│   ├── stats.go
│   └── stats_test.go       # ✅ Тесты статистики
├── repository/postgres/
│   ├── user.go
│   ├── user_test.go        # ✅ Тесты user репозитория
│   ├── word.go
│   └── word_test.go        # ✅ Тесты word репозитория
├── domain/
│   ├── day.go
│   └── day_test.go         # ✅ Тесты форматирования дат
└── testutil/
    ├── testutil.go         # Вспомогательные функции
    └── mocks.go            # Mock объекты
```

## Команды для запуска тестов

### Все тесты

```bash
make test
# или
go test ./...
```

### Unit тесты (только internal/)

```bash
make test-unit
```

### С подробным выводом

```bash
make test-verbose
```

### С покрытием кода

```bash
make test-coverage
```

Откроет HTML отчёт с визуализацией покрытия.

### CI тесты (как в GitLab)

```bash
make test-ci
```

Запускает тесты с:
- Race detector (`-race`)
- Coverage проверкой (минимум 80%)
- Подробным выводом

### Watch mode (автозапуск)

Требует установки `entr`:

```bash
# Linux
sudo apt install entr  # или sudo pacman -S entr

# Запуск
make test-watch
```

Тесты будут перезапускаться при изменении `.go` файлов.

## Типы тестов

### 1. Service Layer Tests (Бизнес-логика)

Тестируют бизнес-правила без реальной БД.

**Пример:** `internal/service/word_test.go`

```go
func TestWordService_SaveWordPair(t *testing.T) {
    mockRepo := new(testutil.MockWordRepository)
    mockRepo.On("SaveWord", int64(123), "hello", "привет").Return(nil)
    
    service := NewWordService(mockRepo)
    err := service.SaveWordPair(123, "hello", "привет")
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

**Что тестируем:**
- Валидация входных данных
- Бизнес-правила (например, слова не пустые)
- Правильность вызовов репозитория

**Используем:**
- Mock репозитории (`testutil.Mock*Repository`)
- Table-driven tests для разных сценариев
- `testify/assert` для проверок

### 2. Repository Layer Tests (Работа с БД)

Тестируют SQL запросы и маппинг данных.

**Пример:** `internal/repository/postgres/user_test.go`

```go
func TestUserRepo_IsAuthorized(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()
    
    repo := NewUserRepo(db)
    
    mock.ExpectQuery("SELECT authorized FROM users").
        WithArgs(int64(123)).
        WillReturnRows(sqlmock.NewRows([]string{"authorized"}).AddRow(true))
    
    authorized, err := repo.IsAuthorized(123)
    
    assert.NoError(t, err)
    assert.True(t, authorized)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

**Что тестируем:**
- Правильность SQL запросов
- Маппинг результатов в структуры
- Обработка ошибок БД
- Edge cases (пустые результаты, ошибки)

**Используем:**
- `sqlmock` для мокирования БД
- Проверка что запросы вызываются с правильными параметрами

### 3. Domain Layer Tests (Модели)

Тестируют логику моделей данных.

**Пример:** `internal/domain/day_test.go`

```go
func TestDay_DisplayString(t *testing.T) {
    day := Day{Date: time.Now()}
    assert.Equal(t, "Сегодня", day.DisplayString())
}
```

**Что тестируем:**
- Методы моделей
- Форматирование данных
- Вычисления

## Написание новых тестов

### Шаблон service теста

```go
package service

import (
    "testing"
    "languager/internal/testutil"
    "github.com/stretchr/testify/assert"
)

func TestMyService_MyMethod(t *testing.T) {
    tests := []struct {
        name          string
        input         string
        expectedError bool
    }{
        {
            name:          "valid input",
            input:         "test",
            expectedError: false,
        },
        {
            name:          "invalid input",
            input:         "",
            expectedError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(testutil.MockWordRepository)
            // Setup mock expectations
            
            service := NewMyService(mockRepo)
            err := service.MyMethod(tt.input)
            
            if tt.expectedError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### Шаблон repository теста

```go
package postgres

import (
    "testing"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
)

func TestMyRepo_MyMethod(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()
    
    repo := NewMyRepo(db)
    
    // Setup expected query
    mock.ExpectQuery("SELECT .* FROM table").
        WithArgs(123).
        WillReturnRows(sqlmock.NewRows([]string{"col"}).AddRow("value"))
    
    result, err := repo.MyMethod(123)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

## Best Practices

### 1. Именование тестов

```go
// ✅ Хорошо: Test{Type}_{Method}
func TestWordService_SaveWordPair(t *testing.T) {}

// ❌ Плохо: неясное имя
func TestSave(t *testing.T) {}
```

### 2. Table-Driven Tests

```go
tests := []struct {
    name     string
    input    int
    expected string
}{
    {"positive", 5, "five"},
    {"negative", -1, "error"},
    {"zero", 0, "zero"},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Тест
    })
}
```

### 3. Использовать testify/assert

```go
// ✅ Хорошо: понятное сообщение об ошибке
assert.Equal(t, expected, actual)

// ❌ Плохо: неинформативная ошибка
if expected != actual {
    t.Fatal("not equal")
}
```

### 4. Изолированные тесты

Каждый тест должен быть независим:

```go
// ✅ Хорошо: свой mock для каждого теста
func TestSomething(t *testing.T) {
    mockRepo := new(testutil.MockWordRepository)
    // ...
}

// ❌ Плохо: общий state между тестами
var globalMock *MockRepository

func TestA(t *testing.T) {
    globalMock.On(...)  // влияет на другие тесты
}
```

### 5. Очистка ресурсов

```go
func TestSomething(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()  // ✅ Всегда закрывай ресурсы
    
    // Тест
}
```

## Покрытие кода (Coverage)

### Целевое покрытие

- **Общее:** минимум **80%**
- **Service layer:** **90%+** (критичная логика)
- **Repository layer:** **80%+**
- **Domain layer:** **80%+**
- **Handler layer:** опционально (сложно мокировать telebot)

### Проверка покрытия

```bash
# Генерация отчёта
make test-coverage

# Открыть HTML отчёт
open coverage.html

# Консольный вывод
go tool cover -func=coverage.out
```

### Что показывает coverage

- **Зелёный:** код выполнялся в тестах
- **Красный:** код НЕ выполнялся
- **Серый:** не исполняемый код (комментарии и т.д.)

## Troubleshooting

### Тесты падают локально

```bash
# Очистить кеш
go clean -testcache

# Установить зависимости
go mod download

# Запустить снова
make test
```

### Race conditions

```bash
# Запустить с race detector
go test -race ./...
```

### Mock не работает

Проверь что:
1. Вызываешь `mockRepo.On()` перед использованием
2. Параметры совпадают (типы и значения)
3. Вызываешь `mockRepo.AssertExpectations(t)` в конце

```go
// Правильная последовательность
mockRepo := new(testutil.MockWordRepository)
mockRepo.On("SaveWord", int64(123), "hello", "привет").Return(nil)  // Setup
service.SaveWordPair(123, "hello", "привет")  // Use
mockRepo.AssertExpectations(t)  // Verify
```

### SQL Mock ошибки

Используй `\\` для экранирования в regex:

```go
// ✅ Хорошо
mock.ExpectQuery("SELECT \\* FROM words WHERE user_id = \\$1")

// ❌ Плохо (regex не сработает)
mock.ExpectQuery("SELECT * FROM words WHERE user_id = $1")
```

## CI/CD Integration

Тесты запускаются автоматически в GitLab CI при push в ветку `private`.

См. [CI_CD.md](CI_CD.md) для деталей.

## Дополнительные ресурсы

- [testify documentation](https://pkg.go.dev/github.com/stretchr/testify)
- [go-sqlmock documentation](https://pkg.go.dev/github.com/DATA-DOG/go-sqlmock)
- [Go testing package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)

