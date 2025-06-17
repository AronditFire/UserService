package app

import (
	grpcapp "github.com/AronditFire/User-Service/internal/app/grpc"
	"github.com/AronditFire/User-Service/internal/services/auth"
	repo "github.com/AronditFire/User-Service/internal/storage/postgres/auth"
	"log/slog"
	"time"
)

const DEFAULT_ROLE = "buyer"

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	postgresDSN string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	tokenSecret string,
) *App {
	storage, err := repo.New(postgresDSN)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, storage, storage, accessTTL, refreshTTL, tokenSecret, DEFAULT_ROLE)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{GRPCServer: grpcApp}
}
