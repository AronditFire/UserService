package grpcapp

import (
	"fmt"
	"github.com/AronditFire/User-Service/internal/grpc/auth"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	GRPCServer *grpc.Server
	port       int
	jwtSecret  string
}

func New(log *slog.Logger, auth authgrpc.Auth, port int, jwtSecret string) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.RegisterUserService(gRPCServer, auth)

	return &App{
		log:        log,
		GRPCServer: gRPCServer,
		port:       port,
		jwtSecret:  jwtSecret,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcAPP.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is running", slog.String("addr", lis.Addr().String()))

	if err := a.GRPCServer.Serve(lis); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcAPP.Stop"

	a.log.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", a.port))

	a.GRPCServer.GracefulStop()
}
