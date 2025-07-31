package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"
)

type SensorRepository struct {
	sensorsByID           map[int64]*domain.Sensor
	sensorsBySerialNumber map[string]*domain.Sensor
	rwMutex               *sync.RWMutex
}

func NewSensorRepository() *SensorRepository {
	return &SensorRepository{
		sensorsByID:           make(map[int64]*domain.Sensor),
		sensorsBySerialNumber: make(map[string]*domain.Sensor),
		rwMutex:               new(sync.RWMutex),
	}
}

var updateID int64 = 1

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if sensor == nil {
			return errors.New("nil sensor")
		}
		r.rwMutex.Lock()
		sensor.RegisteredAt = time.Now()
		if sensor.ID == 0 {
			sensor.ID = updateID
			updateID++
		}
		r.sensorsByID[sensor.ID] = sensor
		r.sensorsBySerialNumber[sensor.SerialNumber] = sensor
		r.rwMutex.Unlock()
		return nil
	}
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.rwMutex.RLock()
		result := make([]domain.Sensor, 0, len(r.sensorsBySerialNumber))
		sensors := r.sensorsBySerialNumber
		r.rwMutex.RUnlock()
		for _, sensor := range sensors {
			r.rwMutex.Lock()
			result = append(result, *sensor)
			r.rwMutex.Unlock()
		}
		return result, nil
	}
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.rwMutex.RLock()
		sensor, ok := r.sensorsByID[id]
		r.rwMutex.RUnlock()
		if !ok {
			return nil, usecase.ErrSensorNotFound
		}
		return sensor, nil
	}
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.rwMutex.RLock()
		sensor, ok := r.sensorsBySerialNumber[sn]
		r.rwMutex.RUnlock()
		if !ok {
			return nil, usecase.ErrSensorNotFound
		}
		return sensor, nil
	}
}
