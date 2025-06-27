package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/AronditFire/User-Service/internal/domain/models"
	"github.com/AronditFire/User-Service/internal/lib/jwt"
	"github.com/AronditFire/User-Service/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidToken       = errors.New("invalid or expired token")
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	roleSetter   RoleSetter
	roleProvider RoleProvider
	tokenRepo    TokenRepo
	accessTTL    time.Duration
	refreshTTL   time.Duration
	secret       string
	defaultRole  string
}

type UserSaver interface {
	SaveUser(ctx context.Context, username string, email string, FIO string, phoneNumber string, passHash string) (int64, error)
}

type RoleSetter interface {
	SetRole(ctx context.Context, userID int64, role string) error
}

type TokenRepo interface {
	SaveToken(ctx context.Context, refreshToken string, userID int64, expiresAt time.Time) error
	GetToken(ctx context.Context, refreshToken string) (models.RefreshTokenClaims, error)
	DeleteToken(ctx context.Context, refreshToken string) error
}

type UserProvider interface {
	User(ctx context.Context, username string) (models.User, error)
}
type RoleProvider interface {
	Role(ctx context.Context, userID int64) (string, error)
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	provider UserProvider,
	roleSetter RoleSetter,
	roleProvider RoleProvider,
	tokenRepo TokenRepo,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	secret string,
	defaultRole string,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: provider,
		roleSetter:   roleSetter,
		roleProvider: roleProvider,
		tokenRepo:    tokenRepo,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
		secret:       secret,
		defaultRole:  defaultRole,
	}
}

// RegisterUser creates a new user and assigns the default role.
func (a *Auth) RegisterUser(ctx context.Context, username, email, FIO, phoneNumber, password string) (int64, error) {
	const op = "auth.RegisterUser"

	log := a.log.With(slog.String("op", op), slog.String("username", username))
	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to generate password hash", slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := a.userSaver.SaveUser(ctx, username, email, FIO, phoneNumber, string(passHash))
	if err != nil {
		a.log.Error("failed to save user", slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if err := a.roleSetter.SetRole(ctx, userID, a.defaultRole); err != nil {
		a.log.Error("failed to set default role", slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully registered user")
	return userID, nil
}

// Login generate tokens if username and password correct
func (a *Auth) Login(ctx context.Context, username, password string) (string, string, error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("username", username))
	log.Info("trying to login user")

	user, err := a.userProvider.User(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", slog.String("error", err.Error()))
			return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("failed to get user", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Error("invalid password", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	role, err := a.roleProvider.Role(ctx, user.ID) // GET USER ROLE
	if err != nil {
		a.log.Error("failed to get role", err.Error())
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := jwt.GenerateToken(user.ID, role, a.secret, a.accessTTL)
	if err != nil {
		a.log.Error("failed to generate token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken := uuid.NewString()
	refreshExpiresAt := time.Now().Add(a.refreshTTL)
	if err := a.tokenRepo.SaveToken(ctx, refreshToken, user.ID, refreshExpiresAt); err != nil {
		a.log.Error("failed to save refresh token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("successfully logged in", slog.String("username", username))

	return accessToken, refreshToken, nil
}

func (a *Auth) Refresh(ctx context.Context, refreshToken string) (string, string, error) { // TODO: изучить логику подробнее
	const op = "auth.Refresh"
	log := a.log.With(slog.String("op", op))
	log.Info("trying to refresh tokens")

	refreshData, err := a.tokenRepo.GetToken(ctx, refreshToken) // достаём данные из таблицы refreshTokens
	if err != nil {
		a.log.Error("failed to get old refresh token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err := a.tokenRepo.DeleteToken(ctx, refreshToken); err != nil { // Удаляем старый refresh Token
		a.log.Error("failed to delete old refresh token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	role, err := a.roleProvider.Role(ctx, refreshData.UserID) // TODO: убрать ошибку
	if err != nil {
		a.log.Error("failed to get role", err.Error())
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	newAccessToken, err := jwt.GenerateToken(refreshData.UserID, role, a.secret, a.accessTTL)
	if err != nil {
		a.log.Error("failed to generate access token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	expiresAt := time.Now().Add(a.refreshTTL)
	newRefreshToken := uuid.NewString()
	if err := a.tokenRepo.SaveToken(ctx, newRefreshToken, refreshData.UserID, expiresAt); err != nil {
		a.log.Error("failed to save new refresh token", slog.String("error", err.Error()))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully refreshed tokens", slog.Int64("userID", refreshData.UserID))
	return newAccessToken, newRefreshToken, nil
}

func (a *Auth) Logout(ctx context.Context, refreshToken string) error {
	const op = "auth.Logout"
	log := a.log.With(slog.String("op", op), slog.String("refreshToken", refreshToken))
	log.Info("start to delete refresh token")

	if err := a.tokenRepo.DeleteToken(ctx, refreshToken); err != nil {
		a.log.Error("failed to delete refresh token", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("token successfully deleted")

	return nil
}
