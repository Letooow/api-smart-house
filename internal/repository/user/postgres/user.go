package postgres

import (
	"context"
	"fmt"
	"homework/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

var updateID int64 = 1

const saveUserQuery = `INSERT INTO users (id, name) VALUES ($1, $2)`

const getUserQuery = `SELECT * FROM users WHERE id = $1`

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	user.ID = updateID

	_, err := r.pool.Exec(ctx, saveUserQuery, user.ID, user.Name)
	if err != nil {
		return fmt.Errorf("can't save user: %w", err)
	}
	updateID++
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, getUserQuery, id)
	user := &domain.User{}
	err := row.Scan(&user.ID, &user.Name)
	if err != nil {
		return nil, fmt.Errorf("can't get user by id: %w", err)
	}
	return user, nil
}
