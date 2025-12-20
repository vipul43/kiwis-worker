package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vipul43/kiwis-worker/internal/config"
	"github.com/vipul43/kiwis-worker/internal/database"
	"github.com/vipul43/kiwis-worker/internal/gmail"
	"github.com/vipul43/kiwis-worker/internal/openrouter"
	"github.com/vipul43/kiwis-worker/internal/repository"
	"github.com/vipul43/kiwis-worker/internal/service"
	"github.com/vipul43/kiwis-worker/internal/watcher"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	log.Println("Database connected successfully")

	// Run migrations
	log.Println("Running database migrations...")
	if err := database.RunMigrations(db); err != nil {
		return err
	}
	log.Println("Migrations completed successfully")

	// Initialize repositories
	accountJobRepo := repository.NewAccountSyncJobRepository(db)
	emailJobRepo := repository.NewEmailSyncJobRepository(db)
	llmJobRepo := repository.NewLLMSyncJobRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	// Initialize services
	accountProcessor := service.NewAccountProcessor(accountRepo)

	// Initialize Gmail client
	gmailClient := gmail.NewClient(cfg.GoogleClientID, cfg.GoogleClientSecret)
	emailProcessor := service.NewEmailProcessor(accountRepo, emailJobRepo, llmJobRepo, gmailClient)

	// Initialize OpenRouter client
	openRouterClient := openrouter.NewClient(cfg.OpenRouterAPIKey)
	llmProcessor := service.NewLLMProcessor(accountRepo, llmJobRepo, paymentRepo, gmailClient, openRouterClient)

	// Initialize watcher
	w := watcher.New(cfg, accountJobRepo, emailJobRepo, llmJobRepo, accountProcessor, emailProcessor, llmProcessor)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start watcher in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- w.Start(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Println("Shutdown signal received")
		cancel()

		// Wait for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeout)*time.Second)
		defer shutdownCancel()

		select {
		case <-shutdownCtx.Done():
			log.Println("Shutdown timeout exceeded")
		case err := <-errChan:
			if err != nil && err != context.Canceled {
				log.Printf("Watcher error: %v", err)
			}
		}

		log.Println("Application stopped")
		return nil

	case err := <-errChan:
		return err
	}
}
