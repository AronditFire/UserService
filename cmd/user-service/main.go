package main

import (
	"github.com/AronditFire/User-Service/internal/app"
	"github.com/AronditFire/User-Service/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	// TODO: при логировании два раза, refresh token от первого логирования не удаляется
	log := SetupLogger(cfg.Env)

	MigratePostgres(cfg.PostgresDSN, cfg.MigrationURL, log)

	log.Info("Starting User Service", slog.Any("config", cfg)) // TODO: убрать конфиг

	application := app.New(log, cfg.GRPC.Port, cfg.PostgresDSN, cfg.AccessTTL, cfg.RefreshTTL, cfg.JWTSecret)

	go application.GRPCServer.MustRun()

	// TODO: shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	application.GRPCServer.Stop()

	log.Info("Gracefully stopped")
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
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

func MigratePostgres(postgresDSN, MigrationURL string, log *slog.Logger) {

	if postgresDSN == "" {
		panic("PostgresDSN is empty")
	}
	if MigrationURL == "" {
		panic("MigrationURL is empty")
	}

	migration, err := migrate.New(MigrationURL, postgresDSN)
	if err != nil {
		log.Error("Migration failed", slog.Any("error", err))
		panic(err)
	}

	if err := migration.Up(); err != nil && (err != migrate.ErrNoChange) {
		log.Error("Migration failed", slog.Any("error", err))
		panic(err)
	}

	log.Info("Migration complete")
}
