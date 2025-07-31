package inmemory

import (
	"context"
	"homework/internal/domain"
	"sync"
)

type SensorOwnerRepository struct {
	// key - UserID
	sensorsOwners map[int64][]domain.SensorOwner
	rw            *sync.RWMutex
}

func NewSensorOwnerRepository() *SensorOwnerRepository {
	return &SensorOwnerRepository{
		sensorsOwners: make(map[int64][]domain.SensorOwner),
		rw:            new(sync.RWMutex),
	}
}

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		r.rw.Lock()
		r.sensorsOwners[sensorOwner.UserID] = append(r.sensorsOwners[sensorOwner.UserID], sensorOwner)
		r.rw.Unlock()
		return nil
	}
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.rw.RLock()
		defer r.rw.RUnlock()
		if val, ok := r.sensorsOwners[userID]; ok {
			return val, nil
		} else {
			return []domain.SensorOwner{}, nil
		}
	}
}
