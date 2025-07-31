package postgres

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEventNotFound = errors.New("event not found")

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{
		pool,
	}
}

const saveEventQuery = `INSERT INTO events (timestamp, sensor_serial_number, sensor_id, payload) VALUES ($1, $2, $3, $4)`

const getLastEventBySensorIDQuery = `SELECT * FROM events WHERE sensor_id = $1 ORDER BY timestamp DESC LIMIT 1`

const getEventsByIDWithDateQuery = `SELECT * FROM events WHERE sensor_id = $1 BETWEEN $2 AND $3`

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	_, err := r.pool.Exec(ctx, saveEventQuery, event.Timestamp, event.SensorSerialNumber, event.SensorID, event.Payload)
	if err != nil {
		return fmt.Errorf("can't save event: %w", err)
	}
	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	row := r.pool.QueryRow(ctx, getLastEventBySensorIDQuery, id)
	event := &domain.Event{}
	err := row.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload)
	if err != nil {
		return nil, ErrEventNotFound
	}
	return event, nil
}

func (r *EventRepository) GetEventsBySensorIDWithDate(ctx context.Context, id int64, start, end time.Time) ([]domain.Event, error) {
	rows, err := r.pool.Query(ctx, getEventsByIDWithDateQuery, id, start, end)
	if err != nil {
		return nil, ErrEventNotFound
	}
	defer rows.Close()
	var events []domain.Event
	for rows.Next() {
		event := domain.Event{}
		err = rows.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}
	return events, nil
}
