package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/config"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/database"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/httpapi"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/repository"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/security"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/service"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	mongoDB, err := database.ConnectMongo(startupCtx, cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		startupCancel()
		logger.Error("MongoDB connection failed", "error", err)
		os.Exit(1)
	}

	userRepository := repository.NewMongoUserRepository(mongoDB.Database())
	if err := userRepository.EnsureIndexes(startupCtx); err != nil {
		_ = mongoDB.Close(startupCtx)
		startupCancel()
		logger.Error("MongoDB index initialization failed", "error", err)
		os.Exit(1)
	}
	startupCancel()
	logger.Info("MongoDB connected", "database", cfg.MongoDatabase)

	passwordHasher := security.NewBcryptPasswordHasher()
	authService := service.NewAuthService(userRepository, passwordHasher)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(mongoDB, authService),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP server started", "address", cfg.HTTPAddr)
		serverErr <- server.ListenAndServe()
	}()

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case <-signalCtx.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed", "error", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown failed", "error", err)
	}

	if err := mongoDB.Close(shutdownCtx); err != nil {
		logger.Error("MongoDB disconnect failed", "error", err)
	}

	logger.Info("application stopped")
}
