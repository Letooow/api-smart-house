package usecase

import (
	"context"
	"homework/internal/domain"
	"time"
)

type Event struct {
	eventRepository  EventRepository
	sensorRepository SensorRepository
}

func NewEvent(er EventRepository, sr SensorRepository) *Event {
	return &Event{
		eventRepository:  er,
		sensorRepository: sr,
	}
}

func (e *Event) ReceiveEvent(ctx context.Context, event *domain.Event) error {
	if event == nil {
		return ErrInvalidEventTimestamp
	}
	if e.sensorRepository != nil {
		sensor, err := e.sensorRepository.GetSensorBySerialNumber(ctx, event.SensorSerialNumber)
		if err != nil {
			return ErrSensorNotFound
		}
		event.Timestamp = time.Now()
		event.SensorID = sensor.ID
		err = e.eventRepository.SaveEvent(ctx, event)
		if err != nil {
			return err
		}

		sensor.LastActivity = time.Now()
		sensor.CurrentState = event.Payload
		err = e.sensorRepository.SaveSensor(ctx, sensor)
		if err != nil {
			return err
		}
	}
	if e.eventRepository == nil {
		return ErrInvalidEventTimestamp
	}
	return nil
}

func (e *Event) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	sensorID, err := e.eventRepository.GetLastEventBySensorID(ctx, id)
	if err != nil {
		return nil, err
	}
	return sensorID, nil
}

func (e *Event) GetEventsBySensorIDWithDate(ctx context.Context, id int64, start, end time.Time) ([]domain.Event, error) {
	if start.After(end) {
		return nil, ErrInputDate
	}
	events, err := e.eventRepository.GetEventsBySensorIDWithDate(ctx, id, start, end)
	if err != nil {
		return nil, err
	}
	return events, nil
}
