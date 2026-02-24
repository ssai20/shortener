package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"url-shortender/internal/http-server/url/save"
	"url-shortender/internal/lib/logger/handlers/redirect"

	//"log/slog"
	"golang.org/x/exp/slog"
	"os"
	"url-shortender/internal/config"
	"url-shortender/internal/http-server/middleware/mwLogger"
	"url-shortender/internal/http-server/url/delete"
	"url-shortender/internal/lib/logger/sl"
	"url-shortender/internal/storage/sqlite"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	//TODO: init config: cleanenv
	cfg := config.MustLoad()
	fmt.Println(cfg)

	//TODO: init logger: slog
	log := setupLogger(cfg.Env)
	log.Info(
		"starting url-shortender",
		slog.String("env", cfg.Env),
		slog.String("version", "123"),
	)
	log.Debug(`debug messages are enabled`)
	//TODO: init storage: sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	//id, err := storage.SaveURL("https://google.com", "google")
	//if err != nil {
	//	log.Error("failed to save url", sl.Err(err))
	//	os.Exit(1)
	//}
	//
	//log.Info("saved url", slog.Int64("id", id))
	//
	//id, err = storage.SaveURL("https://google.com", "google")
	//if err != nil {
	//	log.Error("failed to save url", sl.Err(err))
	//	os.Exit(1)
	//}
	//
	_ = storage
	//TODO:init router: chi, "chi render"

	router := chi.NewRouter()

	//middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", delete.New(log, storage))

	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

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
