package authgrpc

import (
	"context"
	uservicev1 "github.com/AronditFire/UService-ProtobufNew/gen/user-service"
	"github.com/AronditFire/User-Service/internal/lib/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
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

func UnaryAuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	// Определяем списки методов по уровню доступа
	publicMethods := map[string]struct{}{
		"/user_profile.UserService/Register":     {},
		"/user_profile.UserService/Login":        {},
		"/user_profile.UserService/RefreshToken": {},
		"/user_profile.UserService/Logout":       {},
	}
	buyerMethods := map[string]struct{}{
		"/user_profile.UserService/GetProfile": {},
	}
	adminMethods := map[string]struct{}{
		"/user_profile.UserService/ListUsers":  {},
		"/user_profile.UserService/ChangeRole": {},
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 1) Публичные методы без проверки
		if _, ok := publicMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		// 2) Извлекаем Authorization
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		auth := md.Get("authorization")
		if len(auth) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization header required")
		}
		parts := strings.SplitN(auth[0], " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}
		token := parts[1]

		// 3) Верифицируем JWT
		claims, err := jwt.VerifyToken(token, jwtSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// 4) Проверяем роль для buyerMethods
		if _, ok := buyerMethods[info.FullMethod]; ok {
			if claims.Role != "buyer" && claims.Role != "admin" {
				return nil, status.Error(codes.PermissionDenied, "buyer role required")
			}
		}
		// 5) Проверяем роль для adminMethods
		if _, ok := adminMethods[info.FullMethod]; ok {
			if claims.Role != "admin" {
				return nil, status.Error(codes.PermissionDenied, "admin role required")
			}
		}

		// 6) Кладём в контекст
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "role", claims.Role)

		// 7) Вызов обработчика
		return handler(ctx, req)
	}
}
