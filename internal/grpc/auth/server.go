package authgrpc

import (
	uservicev1 "github.com/AronditFire/UService-ProtobufNew/gen"
	"google.golang.org/grpc"
)

type ServerAPI struct {
	uservicev1.UnimplementedUserServiceServer
	auth  Auth
	uProf UserProfile
}

func RegisterUserService(s *grpc.Server, auth Auth, uProf UserProfile) {
	uservicev1.RegisterUserServiceServer(s, &ServerAPI{
		auth:  auth,
		uProf: uProf,
	})
}
