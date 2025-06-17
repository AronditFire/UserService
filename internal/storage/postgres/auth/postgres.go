package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/AronditFire/User-Service/internal/domain/models"
	"github.com/AronditFire/User-Service/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(postgresDSN string) (*Storage, error) {
	const op = "storage.repo.New"

	pool, err := pgxpool.New(context.Background(), postgresDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

// SaveUser saves user to db
func (s *Storage) SaveUser(ctx context.Context, username string,
	email string, FIO string, phoneNumber string, passHash string,
) (int64, error) {
	const op = "storage.repo.SaveUser"

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				// Логируем или добавляем в ошибку
				err = fmt.Errorf("%s: rollback failed: %v; original error: %w", op, rollbackErr, err)
			}
		}
	}()

	var userID int64
	err = tx.QueryRow(ctx,
		"INSERT INTO users (username, email, fio, phone_number, password_hash) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		username, email, FIO, phoneNumber, passHash).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (s *Storage) User(ctx context.Context, username string) (models.User, error) {
	const op = "storage.repo.User"

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				// Логируем или добавляем в ошибку
				err = fmt.Errorf("%s: rollback failed: %v; original error: %w", op, rollbackErr, err)
			}
		}
	}()

	const sp = "sp_get_user"
	if _, err := tx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", sp)); err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	var user models.User
	for i := 0; i < 3; i++ {
		err = tx.QueryRow(ctx, "SELECT id, username, email, FIO, phone_number, password_hash FROM users WHERE username = $1", username).
			Scan(&user.ID, &user.Username, &user.Email, &user.FIO, &user.PhoneNumber, &user.PassHash)
		if err == nil {
			return user, nil
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		_, rbErr := tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", sp))
		if rbErr != nil {
			return models.User{}, fmt.Errorf("%s: %w", op, rbErr)
		}
		continue
	}
	return models.User{}, fmt.Errorf("%s: %w", op, errors.New("could not serialize read transaction"))
}

func (s *Storage) SetRole(ctx context.Context, userID int64, role string) error {
	const op = "storage.repo.SetRole"
	_, err := s.pool.Exec(ctx, `
        INSERT INTO user_roles (user_id, role_id)
        SELECT $1, r.id FROM roles r WHERE r.name = $2
        ON CONFLICT DO NOTHING
    `, userID, role)
	return fmt.Errorf("%s: %w", op, err)
}

func (s *Storage) Role(ctx context.Context, userID int64) (string, error) {
	const op = "storage.repo.Role"

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				// Логируем или добавляем в ошибку
				err = fmt.Errorf("%s: rollback failed: %v; original error: %w", op, rollbackErr, err)
			}
		}
	}()

	const sp = "sp_get_user_role"
	if _, err := tx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", sp)); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var role string
	for i := 0; i < 3; i++ {
		err = tx.QueryRow(ctx, `
            SELECT r.name
            FROM roles r
            JOIN user_roles ur ON ur.role_id = r.id
            WHERE ur.user_id = $1
            LIMIT 1
        `, userID).Scan(&role)
		if err == nil {
			return role, nil
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		_, rbErr := tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", sp))
		if rbErr != nil {
			return "", fmt.Errorf("%s: %w", op, rbErr)
		}
		continue
	}
	return "", fmt.Errorf("%s: %w", op, errors.New("could not serialize read transaction"))
}

func (s *Storage) SaveToken(ctx context.Context, refreshToken string, userID int64, expiresAt time.Time) error {
	const op = "storage.repo.SaveToken"

	_, err := s.pool.Exec(ctx, `
        INSERT INTO refresh_tokens (token, user_id, expires_at)
        VALUES ($1, $2, $3)
    `, refreshToken, userID, expiresAt)

	return fmt.Errorf("%s: %w", op, err)
}

func (s *Storage) GetToken(ctx context.Context, refreshToken string) (models.RefreshTokenClaims, error) {
	const op = "storage.repo.GetToken"

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return models.RefreshTokenClaims{}, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				// Логируем или добавляем в ошибку
				err = fmt.Errorf("%s: rollback failed: %v; original error: %w", op, rollbackErr, err)
			}
		}
	}()

	const sp = "sp_get_refresh_token"
	if _, err := tx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", sp)); err != nil {
		return models.RefreshTokenClaims{}, fmt.Errorf("%s: %w", op, err)
	}

	var rt models.RefreshTokenClaims
	for i := 0; i < 3; i++ {
		err = tx.QueryRow(ctx, `
            SELECT token, user_id, issued_at, expires_at
            FROM refresh_tokens WHERE token = $1
        `, refreshToken).Scan(&rt.Token, &rt.UserID, &rt.IssuedAt, &rt.ExpiresAt)
		if err == nil {
			return rt, nil
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RefreshTokenClaims{}, fmt.Errorf("%s: %w", op, pgx.ErrNoRows)
		}
		_, rbErr := tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", sp))
		if rbErr != nil {
			return models.RefreshTokenClaims{}, fmt.Errorf("%s: %w", op, rbErr)
		}
		continue
	}

	return models.RefreshTokenClaims{}, fmt.Errorf("%s: %w", op, errors.New("could not serialize read transaction"))
}

func (s *Storage) DeleteToken(ctx context.Context, refreshToken string) error {
	const op = "storage.repo.DeleteToken"

	_, err := s.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token = $1`, refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
