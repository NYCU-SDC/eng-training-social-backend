package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/NYCU-SDC/eng-training-social-backend/internal"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/auth"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/comment"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/config"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/database"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/jwt"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/post"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/reaction"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// initialize configuration
	cfg, cfgLog := config.Load()
	err := cfg.Validate()
	if err != nil {
		if errors.Is(err, config.ErrDatabaseURLRequired) {
			title := "Databae URL is required"
			message := "Please set the DATABASE_URL environment variable or provide it in the configuration file."
			message = EarlyApplicationFailed(title, message)
			log.Fatal(message)
		} else {
			log.Fatalf("failed to validate configuration: %v", err)
		}
	}

	// initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	cfgLog.FlushToZap(logger)

	logger.Info("Logger initialized successfully")

	logger.Info("Starting database migration")
	err = database.MigrationUp(cfg.MigrationSource, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("Failed to apply database migrations", zap.Error(err))
	}

	dbPool, err := initDatabasePool(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to initialize database pool", zap.Error(err))
	}
	defer dbPool.Close()

	// initialize validator
	validator := internal.NewValidator()

	// initialize services
	userService := user.NewService(logger, dbPool)
	postService := post.NewService(logger, dbPool)
	jwtService := jwt.NewService(logger, cfg.Secret, time.Minute*15, time.Hour*24, dbPool)
	commentService := comment.NewService(logger, dbPool)
	reactionService := reaction.NewService(logger, dbPool)

	// initialize handlers
	authHandler := auth.NewHandler(logger, cfg, validator, userService, jwtService)
	userHandler := user.NewHandler(logger, validator, userService)
	postHandler := post.NewHandler(logger, validator, postService)
	commentHandler := comment.NewHandler(logger, validator, commentService)
	reactionHandler := reaction.NewHandler(logger, validator, reactionService)

	// initialize middleware
	jwtMiddleware := jwt.NewMiddleware(logger, jwtService)

	// initialize mux
	mux := http.NewServeMux()

	// set up routes
	mux.HandleFunc("GET /api/login/oauth/{provider}", authHandler.Oauth2Start)
	mux.HandleFunc("GET /api/oauth/{provider}/callback", authHandler.Callback)
	mux.HandleFunc("GET /api/oauth/debug/token", authHandler.DebugToken)

	mux.HandleFunc("GET /api/posts", postHandler.GetAllHandler)
	mux.HandleFunc("POST /api/posts", jwtMiddleware.HandlerFunc(postHandler.CreateHandler))
	mux.HandleFunc("GET /api/post/{id}", postHandler.GetByIDHandler)
	mux.HandleFunc("PUT /api/post/{id}", jwtMiddleware.HandlerFunc(postHandler.UpdateHandler))
	mux.HandleFunc("DELETE /api/post/{id}", jwtMiddleware.HandlerFunc(postHandler.DeleteHandler))
	mux.HandleFunc("POST /api/post/{id}/react", jwtMiddleware.HandlerFunc(reactionHandler.ReactToPost))

	mux.HandleFunc("GET /api/users/{id}", jwtMiddleware.HandlerFunc(userHandler.GetByIDHandler))

	mux.HandleFunc("GET /api/comment/{id}", commentHandler.GetByIDHandler)
	mux.HandleFunc("PUT /api/comment/{id}", jwtMiddleware.HandlerFunc(commentHandler.UpdateHandler))
	mux.HandleFunc("DELETE /api/comment/{id}", jwtMiddleware.HandlerFunc(commentHandler.DeleteHandler))
	mux.HandleFunc("GET /api/post/{id}/comments", commentHandler.GetAllByPostIDHandler)
	mux.HandleFunc("POST /api/post/{id}/comments", jwtMiddleware.HandlerFunc(commentHandler.CreateHandler))
	mux.HandleFunc("POST /api/comment/{id}/react", jwtMiddleware.HandlerFunc(reactionHandler.ReactToComment))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		logger.Info("Starting listening request", zap.String("host", cfg.Host), zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Fail to start server with error", zap.Error(err))
		}
	}()

	// wait for context close
	<-ctx.Done()
	logger.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Successfully shutdown")
}

func initDatabasePool(databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func EarlyApplicationFailed(title, action string) string {
	result := `
-----------------------------------------
Application Failed to Start
-----------------------------------------

# What's wrong?
%s

# How to fix it?
%s

`

	result = fmt.Sprintf(result, title, action)
	return result
}
