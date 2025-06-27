package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/AronditFire/User-Service/internal/domain/models"
	"github.com/AronditFire/User-Service/internal/storage"
	"github.com/jackc/pgx/v5"
	"strings"
)

func (s *Storage) GetProfile(ctx context.Context, userID int64) (models.UserWithRole, error) {
	const op = "storage.repo.GetProfile"

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return models.UserWithRole{}, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("%s: rollback failed: %v; original error: %w", op, rollbackErr, err)
			}
		}
	}()

	var user models.UserWithRole
	err = tx.QueryRow(ctx,
		`SELECT
  			u.id,
  			u.username,
  			u.email,
  			u.fio,
  			u.phone_number,
  			r.name AS role
			FROM users u
			JOIN user_roles ur ON ur.user_id = u.id
			JOIN roles r ON r.id = ur.role_id
			WHERE u.id = $1;`,
		userID).Scan(&user.ID, &user.Username, &user.Email, &user.FIO, &user.PhoneNumber, &user.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserWithRole{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.UserWithRole{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return models.UserWithRole{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) GetAllProfiles(ctx context.Context) ([]models.UserWithRole, error) {
	const op = "storage.repo.GetAllProfiles"

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("%s: rollback failed: %v; original error: %w", op, rollbackErr, err)
			}
		}
	}()

	rows, err := tx.Query(ctx,
		`SELECT
  			u.id,
  			u.username,
  			u.email,
  			u.fio,
  			u.phone_number,
  			r.name AS role
			FROM users u
			JOIN user_roles ur ON ur.user_id = u.id
			JOIN roles r ON r.id = ur.role_id
			ORDER BY u.id, r.name;
`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []models.UserWithRole
	for rows.Next() {
		var user models.UserWithRole
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FIO, &user.PhoneNumber, &user.Role); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func (s *Storage) ChangeRole(ctx context.Context, userID int64, role string) error {
	const op = "storage.repo.ChangeRole"

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("%s: rollback failed: %v; original error: %w", op, rollbackErr, err)
			}
		}
	}()

	row := tx.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, strings.ToLower(role))
	var roleID int
	if err := row.Scan(&roleID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, storage.ErrRoleNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx,
		"UPDATE user_roles SET role_id = $1 WHERE user_id = $2",
		roleID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
