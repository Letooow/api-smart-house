package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
)

type void struct{}

type Sensor struct {
	sensorRepository SensorRepository
	sensorTypes      map[domain.SensorType]void
}

func NewSensor(sr SensorRepository) *Sensor {
	return &Sensor{
		sensorRepository: sr,
		sensorTypes: map[domain.SensorType]void{
			domain.SensorTypeContactClosure: {},
			domain.SensorTypeADC:            {},
		},
	}
}

func (s *Sensor) RegisterSensor(ctx context.Context, sensor *domain.Sensor) (*domain.Sensor, error) {
	if len(sensor.SerialNumber) != 10 {
		return nil, ErrWrongSensorSerialNumber
	}
	if _, ok := s.sensorTypes[sensor.Type]; !ok {
		return nil, ErrWrongSensorType
	}

	sensorBySerialNumber, err := s.sensorRepository.GetSensorBySerialNumber(ctx, sensor.SerialNumber)
	if err != nil {
		if errors.Is(err, ErrSensorNotFound) {
			err := s.sensorRepository.SaveSensor(ctx, sensor)
			if err != nil {
				return nil, err
			}
			return sensor, nil
		}
		return nil, err
	}
	return sensorBySerialNumber, nil
}

func (s *Sensor) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	return s.sensorRepository.GetSensors(ctx)
}

func (s *Sensor) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	sensor, err := s.sensorRepository.GetSensorByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return sensor, nil
}
