package app

import (
	grpcapp "github.com/AronditFire/User-Service/internal/app/grpc"
	"github.com/AronditFire/User-Service/internal/services/auth"
	uprofile "github.com/AronditFire/User-Service/internal/services/userProfile"
	repo "github.com/AronditFire/User-Service/internal/storage/postgres/auth"
	"log/slog"
	"net/http"
	"time"
)

const DEFAULT_ROLE = "buyer"

type App struct {
	GRPCServer     *grpcapp.App
	GINHTTPGateway *http.Server
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
	profileService := uprofile.New(log, storage, storage)

	grpcApp := grpcapp.New(log, authService, profileService, grpcPort, tokenSecret)

	return &App{GRPCServer: grpcApp}
}
