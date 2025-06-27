package authgrpc

import (
	"context"
	"errors"
	uservicev1 "github.com/AronditFire/UService-ProtobufNew/gen/user-service"
	"github.com/AronditFire/User-Service/internal/domain/models"
	uprofile "github.com/AronditFire/User-Service/internal/services/userProfile"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
)

var (
	BuyerRole = "BUYER"
	AdminRole = "ADMIN"
)

type UserProfile interface {
	GetProfile(ctx context.Context, userID int64) (*models.UserWithRole, error)
	GetAllProfiles(ctx context.Context) ([]models.UserWithRole, error)
	ChangeRole(ctx context.Context, userID int64, role string) error
}

func (s *ServerAPI) GetProfile(ctx context.Context, req *uservicev1.GetProfileRequest) (*uservicev1.UserProfileResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID must be greater than 0")
	}
	if (req.GetUserId() != ctx.Value("user_id")) && (ctx.Value("role") != "admin") {
		return nil, status.Error(codes.PermissionDenied, "user ID is not allowed")
	}

	user, err := s.uProf.GetProfile(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, uprofile.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	roleStr := strings.ToUpper(user.Role) // "BUYER", "ADMIN"

	// смотрим в сгенерированную мапу name->value
	val, ok := uservicev1.Roles_value[roleStr]
	if !ok {
		val = int32(uservicev1.Roles_UNKNOWN)
	}

	return &uservicev1.UserProfileResponse{
		Id:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FIO:         user.FIO,
		PhoneNumber: user.PhoneNumber,
		Role:        uservicev1.Roles(val),
	}, nil
}

func (s *ServerAPI) ListUsers(ctx context.Context, _ *emptypb.Empty) (*uservicev1.UserListResponse, error) {
	users, err := s.uProf.GetAllProfiles(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &uservicev1.UserListResponse{Users: make([]*uservicev1.UserProfileResponse, len(users))}
	for i, user := range users {

		roleStr := strings.ToUpper(user.Role) // "BUYER", "ADMIN"

		val, ok := uservicev1.Roles_value[roleStr]
		if !ok {
			val = int32(uservicev1.Roles_UNKNOWN)
		}

		resp.Users[i] = &uservicev1.UserProfileResponse{
			Id:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			FIO:         user.FIO,
			PhoneNumber: user.PhoneNumber,
			Role:        uservicev1.Roles(val),
		}
	}

	return resp, nil
}

func (s *ServerAPI) ChangeRole(ctx context.Context, req *uservicev1.AdminRoleRequest) (*emptypb.Empty, error) {
	if err := ValidateChangeRole(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.uProf.ChangeRole(ctx, req.GetUserId(), req.GetRole().String()); err != nil {
		if errors.Is(err, uprofile.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func ValidateChangeRole(req *uservicev1.AdminRoleRequest) error {
	if req.GetUserId() <= 0 {
		return errors.New("user ID must be greater than 0")
	}
	if (req.GetRole().String() != BuyerRole) && (req.GetRole().String() != AdminRole) {
		return errors.New("invalid role")
	}

	return nil
}
