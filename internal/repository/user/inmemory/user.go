package inmemory

import (
	"context"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
)

type UserRepository struct {
	users map[int64]*domain.User
	rw    *sync.RWMutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[int64]*domain.User),
		rw:    new(sync.RWMutex),
	}
}

var updateID int64 = 1

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if user == nil {
			return usecase.ErrInvalidUserName
		}
		r.rw.Lock()
		defer r.rw.Unlock()
		if _, ok := r.users[user.ID]; !ok {
			if user.ID == 0 {
				user.ID = updateID
				r.users[updateID] = user
				updateID++
			} else {
				r.users[user.ID] = user
			}
		}
		return nil
	}
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.rw.RLock()
		user, ok := r.users[id]
		r.rw.RUnlock()
		if !ok {
			return nil, usecase.ErrUserNotFound
		}
		return user, nil
	}
}
