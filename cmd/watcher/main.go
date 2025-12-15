package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/payment-tracker/internal/config"
	"github.com/yourusername/payment-tracker/internal/database"
	"github.com/yourusername/payment-tracker/internal/repository"
	"github.com/yourusername/payment-tracker/internal/service"
	"github.com/yourusername/payment-tracker/internal/watcher"
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
	jobRepo := repository.NewAccountSyncJobRepository(db)
	accountRepo := repository.NewAccountRepository(db)

	// Initialize services
	accountProcessor := service.NewAccountProcessor(accountRepo)

	// Initialize watcher
	w := watcher.New(cfg, jobRepo, accountProcessor)

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
