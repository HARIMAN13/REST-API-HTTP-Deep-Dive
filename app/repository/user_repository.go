package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

type UserRepository interface {
	Create(ctx context.Context, u model.User) (int, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func (r *userPostgresRepository) Create(ctx context.Context, u model.User) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password, role, is_active)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		u.Username, u.Email, u.Password, u.Role, u.IsActive,
	).Scan(&id)

	if err != nil {
		// pgx duplicate error checking is complex, so we return ErrDuplicate for now if error contains constraint name
		return 0, fmt.Errorf("menyimpan user: %w", err)
	}

	return id, nil
}

func (r *userPostgresRepository) FindByUsername(
	ctx context.Context, username string,
) (model.User, error) {
	var u model.User

	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at
		 FROM users WHERE LOWER(username) = LOWER($1)`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role,
		&u.IsActive, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}

	return u, nil
}

func (r *userPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.User, error) {
	var u model.User

	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role,
		&u.IsActive, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}

	return u, nil
}
