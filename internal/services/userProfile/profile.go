package uprofile

import (
	"context"
	"github.com/AronditFire/User-Service/internal/domain/models"
	"log/slog"
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
	GetProfile(ctx context.Context, userID int64) (models.User, error)
}

type AdminFunctions interface {
	GetAllProfiles(ctx context.Context) ([]models.User, error)
	ChangeRole(ctx context.Context, userID int64, role string) error
}
