package authgrpc

import (
	"context"
	uservicev1 "github.com/AronditFire/UService-ProtobufNew/gen"
	val "github.com/AronditFire/User-Service/internal/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Auth interface {
	Login(ctx context.Context, username string, password string) (string, string, error)
	RegisterUser(ctx context.Context, username string, email string, FIO string, phoneNumber string, password string) (int64, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error) // new access, new refresh, error
	Logout(ctx context.Context, refreshToken string) error
}

func (s *ServerAPI) Register(ctx context.Context, req *uservicev1.RegisterRequest) (*uservicev1.RegisterResponse, error) {
	if err := ValidateRegister(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error()) // TODO: validate errors
	}
	userID, err := s.auth.RegisterUser(ctx, req.GetUsername(), req.GetEmail(), req.GetFIO(), req.GetPhoneNumber(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &uservicev1.RegisterResponse{UserId: userID}, nil
}

func (s *ServerAPI) Login(ctx context.Context, req *uservicev1.LoginRequest) (*uservicev1.LoginResponse, error) {
	if err := val.CheckUsername(req.GetUsername()); err != nil {
		return nil, err
	}
	if err := val.CheckPassword(req.GetPassword()); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.auth.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error()) // TODO: maybe change the code
	}

	return &uservicev1.LoginResponse{
		Tokens: &uservicev1.Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

// RefreshToken generate new token pair if refresh token not expired
func (s *ServerAPI) RefreshToken(ctx context.Context, req *uservicev1.RefreshRequest) (*uservicev1.RefreshResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is empty")
	}

	newAccessToken, newRefreshToken, err := s.auth.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &uservicev1.RefreshResponse{Tokens: &uservicev1.Tokens{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}}, nil
}

// Logout delete current refresh token from database
func (s *ServerAPI) Logout(ctx context.Context, req *uservicev1.LogoutRequest) (*emptypb.Empty, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is empty")
	}

	if err := s.auth.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, status.Error(codes.Internal, err.Error()) // TODO: maybe 2 error type: expired and internal error
	}

	return &emptypb.Empty{}, nil
}

func ValidateRegister(req *uservicev1.RegisterRequest) error {
	if err := val.CheckUsername(req.GetUsername()); err != nil {
		return err
	}
	if err := val.CheckPassword(req.GetPassword()); err != nil {
		return err
	}
	if err := val.CheckEmail(req.GetEmail()); err != nil {
		return err
	}
	if err := val.CheckPhoneNumber(req.GetPhoneNumber()); err != nil {
		return err
	}
	if err := val.CheckFIO(req.GetFIO()); err != nil {
		return err
	}
	return nil
}
