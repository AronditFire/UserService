package uprofile

import (
	"context"
	"errors"
	"fmt"
	"github.com/AronditFire/User-Service/internal/domain/models"
	"github.com/AronditFire/User-Service/internal/storage"
	"log/slog"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserProfile struct {
	log             *slog.Logger
	profileProvider ProfileProvider
	adminFunctions  AdminFunctions
}

func New(
	log *slog.Logger,
	profileProvider ProfileProvider,
	adminFunctions AdminFunctions,
) *UserProfile {
	return &UserProfile{
		log:             log,
		profileProvider: profileProvider,
		adminFunctions:  adminFunctions,
	}
}

type ProfileProvider interface {
	GetProfile(ctx context.Context, userID int64) (models.UserWithRole, error)
}

type AdminFunctions interface {
	GetAllProfiles(ctx context.Context) ([]models.UserWithRole, error)
	ChangeRole(ctx context.Context, userID int64, role string) error
}

func (u *UserProfile) GetProfile(ctx context.Context, userID int64) (*models.UserWithRole, error) {
	const op = "uprofile.GetProfile"
	log := u.log.With(slog.String("op", op), slog.Int64("userID", userID))
	log.Info("Fetching user profile")

	user, err := u.profileProvider.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			u.log.Warn("user not found", slog.String("error", err.Error()))
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		u.log.Error("failed to get user profile", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Successfully fetched user profile")

	return &user, nil
}

func (u *UserProfile) GetAllProfiles(ctx context.Context) ([]models.UserWithRole, error) {
	const op = "uprofile.GetAllProfiles"
	log := u.log.With(slog.String("op", op))
	log.Info("Fetching all user profiles")

	profiles, err := u.adminFunctions.GetAllProfiles(ctx)
	if err != nil {
		u.log.Error("failed to get all user profiles", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Successfully fetched all user profiles")
	return profiles, nil
}

func (u *UserProfile) ChangeRole(ctx context.Context, userID int64, role string) error {
	const op = "uprofile.ChangeRole"
	log := u.log.With(slog.String("op", op), slog.Int64("userID", userID), slog.String("role", role))
	log.Info("Changing user role")

	if err := u.adminFunctions.ChangeRole(ctx, userID, role); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			u.log.Warn("user not found", slog.String("error", err.Error()))
			return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		u.log.Error("failed to change user role", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Successfully changed user role")
	return nil
}
