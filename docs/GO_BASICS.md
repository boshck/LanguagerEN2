# Go для начинающих - объяснение кода

Ты говорил что никогда не писал на Go, поэтому вот краткое объяснение того, что происходит в проекте.

## Базовые концепции Go

### 1. Package (пакеты)

Каждый файл начинается с `package`:

```go
package main  // Главный пакет (точка входа)
package service  // Другие пакеты
```

- `package main` - это точка входа программы
- Остальные пакеты - это модули кода

### 2. Import (импорты)

```go
import (
    "fmt"  // Стандартная библиотека
    "languager/internal/config"  // Наш код
    "github.com/lib/pq"  // Внешняя библиотека
)
```

Импортируем нужные модули.

### 3. Структуры (struct)

```go
type User struct {
    UserID     int64
    Authorized bool
}
```

Это как класс в других языках - набор полей.

### 4. Функции

```go
// Обычная функция
func Hello() {
    fmt.Println("Hello")
}

// Функция с параметрами и возвратом
func Add(a int, b int) int {
    return a + b
}

// Функция с несколькими возвратами (типично для Go)
func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}
```

### 5. Методы (привязаны к структурам)

```go
type Calculator struct {
    value int
}

// Метод структуры Calculator
func (c *Calculator) Add(x int) {
    c.value += x
}
```

`(c *Calculator)` - это receiver, говорит что функция принадлежит структуре.

### 6. Интерфейсы

```go
type Repository interface {
    Save(data string) error
    Load() (string, error)
}
```

Интерфейс - это контракт. Любая структура с этими методами реализует интерфейс.

### 7. Error handling (обработка ошибок)

В Go нет try-catch, используется возврат ошибок:

```go
result, err := SomeFunction()
if err != nil {
    // Обработка ошибки
    return err
}
// Продолжаем работу с result
```

## Наш проект - как это работает

### Структура (Clean Architecture)

```
main.go → Handler → Service → Repository → Database
```

**Аналогия:**
- **main.go** - директор компании (запускает всё)
- **Handler** - секретарь (общается с клиентами через Telegram)
- **Service** - менеджер (бизнес-логика, принимает решения)
- **Repository** - архивариус (работает с базой данных)

### Пример потока: Сохранение слова

#### 1. Handler получает сообщение от Telegram

```go
// internal/handler/word.go
func (h *Handler) handleText(c tele.Context) error {
    text := c.Text()  // "hello"
    
    // Сохраняем через service
    err := h.wordService.SaveWordPair(userID, word, translation)
    if err != nil {
        return c.Send("Ошибка!")
    }
    
    return c.Send("Сохранено!")
}
```

**Что происходит:**
- Получили текст от пользователя
- Позвали service для сохранения
- Ответили пользователю

#### 2. Service проверяет и передаёт в Repository

```go
// internal/service/word.go
func (s *WordService) SaveWordPair(userID int64, word, translation string) error {
    // Бизнес-правило: слова не пустые
    if word == "" || translation == "" {
        return fmt.Errorf("empty word")
    }
    
    // Передаём в repository для сохранения
    return s.wordRepo.SaveWord(userID, word, translation)
}
```

**Что происходит:**
- Проверили что слова не пустые (бизнес-логика)
- Позвали repository

#### 3. Repository сохраняет в БД

```go
// internal/repository/postgres/word.go
func (r *WordRepo) SaveWord(userID int64, word, translation string) error {
    query := `INSERT INTO words (user_id, word, translation) VALUES ($1, $2, $3)`
    _, err := r.db.Exec(query, userID, word, translation)
    return err
}
```

**Что происходит:**
- Выполнили SQL запрос
- Вернули ошибку (если была) или nil

### Dependency Injection (внедрение зависимостей)

Вместо создания объектов внутри, мы передаём их снаружи:

```go
// Плохо (жёсткая связь):
type Service struct {}

func (s *Service) Do() {
    repo := NewRepository()  // Создаём здесь
    repo.Save()
}

// Хорошо (гибкая связь):
type Service struct {
    repo Repository  // Интерфейс
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}  // Передаём снаружи
}

func (s *Service) Do() {
    s.repo.Save()  // Используем переданный
}
```

**Зачем:**
- Легко тестировать (подставляем mock вместо реальной БД)
- Легко менять реализацию (PostgreSQL → MongoDB)

### State Machine (машина состояний)

Бот запоминает в каком состоянии пользователь:

```go
// internal/handler/handler.go
states := map[int64]*StateData{}  // userID → состояние

// Пользователь отправил слово
states[userID] = &StateData{
    State: StateWaitingTranslation,
    CurrentWord: "hello",
}

// Следующее сообщение - это перевод
if state.State == StateWaitingTranslation {
    word := state.CurrentWord  // "hello"
    translation := newMessage  // "привет"
    SavePair(word, translation)
}
```

### Goroutines (легковесные потоки)

```go
// Запускаем в фоне
go func() {
    // Этот код выполняется параллельно
    bot.Start()
}()

// Продолжаем основной код
waitForShutdown()
```

`go` - ключевое слово для запуска в фоне.

### Channels (каналы)

```go
sigChan := make(chan os.Signal, 1)  // Создали канал
signal.Notify(sigChan, os.Interrupt)  // Слушаем сигналы

<-sigChan  // Ждём пока не придёт сигнал
```

Каналы - это очереди для передачи данных между goroutines.

## Важные моменты Go

### 1. Указатели

```go
func Modify(x int) {
    x = 10  // Меняем копию
}

func ModifyPtr(x *int) {
    *x = 10  // Меняем оригинал
}

a := 5
Modify(a)     // a всё ещё 5
ModifyPtr(&a) // a теперь 10
```

- `*Type` - указатель на тип
- `&variable` - взять адрес переменной
- `*pointer` - получить значение по указателю

### 2. Defer (отложенное выполнение)

```go
func DoSomething() {
    file, _ := os.Open("file.txt")
    defer file.Close()  // Выполнится в конце функции
    
    // Работаем с файлом
    // ...
}  // <- Здесь автоматически вызовется file.Close()
```

### 3. Nil (пустое значение)

```go
var x *int = nil  // Пустой указатель
var s []int = nil // Пустой slice
var m map[string]int = nil  // Пустая map

if err != nil {  // Проверка на ошибку
    // Есть ошибка
}
```

### 4. := vs =

```go
var x int = 5      // Явное объявление
x := 5             // Короткое объявление (только внутри функций)
x = 10             // Присваивание (переменная уже существует)
```

## Как читать наш код

### Пример: internal/service/word.go

```go
package service  // Пакет

import (
    "fmt"
    "languager/internal/domain"
    "languager/internal/repository"
)

// WordService - структура сервиса
type WordService struct {
    wordRepo repository.WordRepository  // Зависимость
}

// NewWordService - конструктор (создаёт новый сервис)
func NewWordService(wordRepo repository.WordRepository) *WordService {
    return &WordService{wordRepo: wordRepo}
}

// SaveWordPair - метод сервиса
func (s *WordService) SaveWordPair(userID int64, word, translation string) error {
    // Валидация
    if word == "" || translation == "" {
        return fmt.Errorf("word and translation cannot be empty")
    }
    
    // Делегируем repository
    return s.wordRepo.SaveWord(userID, word, translation)
}
```

**Читаем как:**
1. Это пакет service
2. WordService хранит wordRepo (для работы с БД)
3. NewWordService создаёт новый сервис (конструктор)
4. SaveWordPair проверяет данные и сохраняет через repo

## Docker команды для понимания

```bash
# Запустить контейнеры
docker-compose up -d

# Посмотреть логи
docker-compose logs -f bot

# Зайти внутрь контейнера
docker-compose exec bot sh

# Остановить всё
docker-compose down

# Пересобрать образ
docker-compose up -d --build
```

## Если хочешь изменить код

### Добавить новую команду бота

1. **Handler** - добавь обработчик:
```go
// internal/handler/myfeature.go
func (h *Handler) handleMyCommand(c tele.Context) error {
    return c.Send("Привет!")
}
```

2. **Регистрация** - зарегистрируй в RegisterHandlers:
```go
// internal/handler/handler.go
func (h *Handler) RegisterHandlers() {
    h.bot.Handle("/mycommand", h.handleMyCommand)
}
```

3. **Перезапуск**:
```bash
make restart
```

### Добавить новое поле в БД

1. **Миграция** - создай migrations/003_add_field.sql:
```sql
ALTER TABLE words ADD COLUMN difficulty INTEGER DEFAULT 0;
```

2. **Domain** - добавь в модель:
```go
// internal/domain/word.go
type Word struct {
    // ...
    Difficulty int
}
```

3. **Repository** - обнови запросы:
```go
// internal/repository/postgres/word.go
query := `SELECT id, word, translation, difficulty FROM words...`
```

4. **Перезапуск** (миграция применится автоматически):
```bash
make restart
```

## Полезные ресурсы

- [Go Tour](https://go.dev/tour/) - интерактивное введение в Go
- [Go by Example](https://gobyexample.com/) - примеры кода
- [Effective Go](https://go.dev/doc/effective_go) - лучшие практики

## Вопросы?

Если что-то непонятно - гляди в код, там много комментариев, или спрашивай!

