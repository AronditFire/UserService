package grpcuserver

import (
	uservicev1 "github.com/AronditFire/User-Service-Protobuf/gen"
	authgrpc "github.com/AronditFire/User-Service/internal/grpc/auth"
	profgrpc "github.com/AronditFire/User-Service/internal/grpc/profile"
	"google.golang.org/grpc"
)

type ServerAPI struct {
	uservicev1.UnimplementedUserServiceServer
	auth  authgrpc.Auth
	uProf profgrpc.UserProfile
}

func RegisterUserService(s *grpc.Server, auth authgrpc.Auth, uProf profgrpc.UserProfile) {
	uservicev1.RegisterUserServiceServer(s, &ServerAPI{
		auth:  auth,
		uProf: uProf,
	})
}
