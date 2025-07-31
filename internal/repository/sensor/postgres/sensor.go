package postgres

import (
	"context"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SensorRepository struct {
	pool *pgxpool.Pool
}

func NewSensorRepository(pool *pgxpool.Pool) *SensorRepository {
	return &SensorRepository{
		pool: pool,
	}
}

var updateSensorID int64 = 1

const saveSensorQuery = `INSERT INTO sensors (id, serial_number, type, current_state, description, is_active, registered_at, last_activity) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

const updateSensorQuery = `UPDATE sensors SET current_state = $2, description = $3, is_active = $4, last_activity = $5 WHERE id = $1`

const getSensorsQuery = `SELECT * FROM sensors`

const getSensorByID = `SELECT * FROM sensors WHERE id = $1`

const getSensorBySerialNumber = `SELECT * FROM sensors WHERE serial_number = $1`

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	if sensor.ID == 0 {
		sensor.ID = updateSensorID
	}

	if _, err := r.GetSensorByID(ctx, sensor.ID); err == nil {
		_, err = r.pool.Exec(ctx, updateSensorQuery, sensor.ID, sensor.CurrentState, sensor.Description, sensor.IsActive, sensor.LastActivity)
		if err != nil {
			return err
		}
		return nil
	}
	sensor.IsActive = false
	atomic.AddInt64(&updateSensorID, 1)
	_, err := r.pool.Exec(ctx, saveSensorQuery, sensor.ID, sensor.SerialNumber, sensor.Type, sensor.CurrentState, sensor.Description, sensor.IsActive, time.Now().Truncate(time.Microsecond), sensor.LastActivity)
	if err != nil {
		atomic.AddInt64(&updateSensorID, -1)
		return err
	}
	return nil
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	row, err := r.pool.Query(ctx, getSensorsQuery)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	var sensors []domain.Sensor
	for row.Next() {
		var sensor domain.Sensor
		err = row.Scan(&sensor.ID, &sensor.SerialNumber, &sensor.Type, &sensor.CurrentState, &sensor.Description, &sensor.IsActive, &sensor.RegisteredAt, &sensor.LastActivity)
		if err != nil {
			return nil, err
		}

		sensors = append(sensors, sensor)
	}
	return sensors, nil
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	row := r.pool.QueryRow(ctx, getSensorByID, id)
	var sensor domain.Sensor
	err := row.Scan(&sensor.ID, &sensor.SerialNumber, &sensor.Type, &sensor.CurrentState, &sensor.Description, &sensor.IsActive, &sensor.RegisteredAt, &sensor.LastActivity)
	if err != nil {
		return nil, err
	}
	return &sensor, nil
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	row := r.pool.QueryRow(ctx, getSensorBySerialNumber, sn)
	var sensor domain.Sensor
	err := row.Scan(&sensor.ID, &sensor.SerialNumber, &sensor.Type, &sensor.CurrentState, &sensor.Description, &sensor.IsActive, &sensor.RegisteredAt, &sensor.LastActivity)
	if err != nil {
		return nil, usecase.ErrSensorNotFound
	}
	return &sensor, nil
}
