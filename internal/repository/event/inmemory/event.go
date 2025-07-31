package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"math"
	"sync"
	"time"
)

type EventRepository struct {
	// key - SensorID, value - events slice
	events  map[int64][]*domain.Event
	rwMutex *sync.RWMutex
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		events:  make(map[int64][]*domain.Event),
		rwMutex: new(sync.RWMutex),
	}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if event == nil {
			return errors.New("event is nil")
		}
		r.rwMutex.Lock()
		r.events[event.SensorID] = append(r.events[event.SensorID], event)
		r.rwMutex.Unlock()
		return nil
	}
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.rwMutex.RLock()
		if _, ok := r.events[id]; !ok {
			return nil, usecase.ErrEventNotFound
		}
		events := r.events[id]
		r.rwMutex.RUnlock()

		var diffTime int64 = math.MaxInt64
		var resEvent *domain.Event
		for _, event := range events {
			timeCheck := time.Since(event.Timestamp).Nanoseconds()
			if timeCheck < diffTime {
				resEvent = event
				diffTime = timeCheck
			}
		}
		return resEvent, nil
	}
}

func (r *EventRepository) GetEventsBySensorIDWithDate(ctx context.Context, id int64, start, end time.Time) ([]domain.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.rwMutex.RLock()
		if _, ok := r.events[id]; !ok {
			return nil, usecase.ErrEventNotFound
		}
		var events []domain.Event
		for _, event := range r.events[id] {
			if event.Timestamp.After(start) && event.Timestamp.Before(end) {
				events = append(events, *event)
			}
		}
		r.rwMutex.RUnlock()
		return events, nil
	}
}
