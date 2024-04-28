package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"test-task/internal/service/auth"
	"test-task/internal/config"
	"test-task/internal/handler/auth/refresh"
	"test-task/internal/handler/auth/signin"
	"test-task/internal/handler/auth/signup"
	"test-task/internal/storage/mongodb"
	tokenmanager "test-task/pkg/auth"
	"test-task/pkg/client/mongodb"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	pathConfig = "config/local.yaml"
	envLocal    = "local"
	envDev      = "dev"
	envProd     = "prod"
)

func main(){
	// configuration service
	cfg := config.GetConfig(pathConfig)

	address := getAdress(cfg.Listen.BindIP, cfg.Listen.Port)
	// init logger
	log := setupLogger(cfg.Env)
	log.Info(
		"starting auth service",
		slog.String("env", cfg.Env),
	)
	// initialization mongoDB client
	
	db, err := mongodb.NewClient(context.Background(), cfg.Storage, "")
	if err != nil {
		log.Error(fmt.Sprintf("failed to init storage %v", err))
		os.Exit(1)
	}
	defer func() {
		if err := db.Client().Disconnect(context.TODO()); err != nil {
			panic(err)
		}
		log.Info("Disconnected from mongoDB")
	}()
	log.Debug("Create DB connection")
	// auth
	fmt.Println(cfg)
	tm, err := tokenmanager.NewManager(cfg.Auth.SigningKey)
	if err!= nil {
        log.Error(fmt.Sprintf("failed to init token manager - %v", err))
        os.Exit(1)
    }
	repository := mngstorage.New(*db)
	auth := auth.New(log, repository, *tm, cfg.Auth)
	// create a router
	router := chi.NewRouter()
	// init middlewares
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/auth", func(r chi.Router) {
		r.Post("/signin/", signin.New(log, auth))
		r.Post("/signup/", signup.New(log, auth))
		r.Post("/refresh/", refresh.New(log, auth))
	})

	log.Info("starting server", slog.String("address", address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	serv := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  cfg.Listen.Timeout,
		WriteTimeout: cfg.Listen.Timeout,
		IdleTimeout:  cfg.Listen.IdleTimeout,
	}

	go func() {
		if err := serv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")
	// Init context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := serv.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf("failed to stop server - %v", err))
		return
	}

	
	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}

func getAdress(ip, port string) string {
	return fmt.Sprintf("%s:%s", ip, port)
}