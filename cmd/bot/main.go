package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"languager/internal/config"
	"languager/internal/handler"
	"languager/internal/repository/postgres"
	"languager/internal/service"

	"github.com/golang-migrate/migrate/v4"
	postgresdb "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Languager Bot")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	logger.Info("Configuration loaded successfully")

	// Connect to database with retries
	db, err := connectDatabase(cfg.DSN(), logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Database connection established")

	// Run migrations
	if err := runMigrations(db, logger); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	logger.Info("Database migrations completed")

	// Initialize repositories
	userRepo := postgres.NewUserRepo(db)
	wordRepo := postgres.NewWordRepo(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.BotPassword)
	wordService := service.NewWordService(wordRepo)
	statsService := service.NewStatsService(wordRepo, logger)

	// Initialize Telegram bot
	bot, err := tele.NewBot(tele.Settings{
		Token:  cfg.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		logger.Fatal("Failed to create bot", zap.Error(err))
	}

	logger.Info("Telegram bot initialized")

	// Initialize handler
	h := handler.NewHandler(bot, authService, wordService, logger)
	h.RegisterHandlers()

	logger.Info("Handlers registered")

	// Start cleanup job in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runCleanupJob(ctx, statsService, logger)

	// Start bot in background
	go func() {
		logger.Info("Bot started successfully")
		bot.Start()
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	logger.Info("Shutdown signal received, stopping bot...")

	// Graceful shutdown
	bot.Stop()
	cancel()

	logger.Info("Bot stopped gracefully")
}

// connectDatabase connects to PostgreSQL with retries
func connectDatabase(dsn string, logger *zap.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error

	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			logger.Warn("Failed to open database connection",
				zap.Int("attempt", i+1),
				zap.Error(err),
			)
			time.Sleep(retryDelay)
			continue
		}

		// Test connection
		if err = db.Ping(); err != nil {
			logger.Warn("Failed to ping database",
				zap.Int("attempt", i+1),
				zap.Error(err),
			)
			db.Close()
			time.Sleep(retryDelay)
			continue
		}

		// Connection successful
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		return db, nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// runMigrations runs database migrations
func runMigrations(db *sql.DB, logger *zap.Logger) error {
	driver, err := postgresdb.WithInstance(db, &postgresdb.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		logger.Info("No new migrations to apply")
	} else {
		logger.Info("Migrations applied successfully")
	}

	return nil
}

// runCleanupJob runs periodic cleanup of old data
func runCleanupJob(ctx context.Context, statsService *service.StatsService, logger *zap.Logger) {
	// Run cleanup once at startup
	if err := statsService.CleanupOldData(); err != nil {
		logger.Error("Failed to run initial cleanup", zap.Error(err))
	}

	// Then run every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Cleanup job stopped")
			return
		case <-ticker.C:
			logger.Info("Running scheduled cleanup")
			if err := statsService.CleanupOldData(); err != nil {
				logger.Error("Failed to run scheduled cleanup", zap.Error(err))
			}
		}
	}
}

