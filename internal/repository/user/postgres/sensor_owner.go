package postgres

import (
	"context"
	"fmt"
	"homework/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SensorOwnerRepository struct {
	pool *pgxpool.Pool
}

func NewSensorOwnerRepository(pool *pgxpool.Pool) *SensorOwnerRepository {
	return &SensorOwnerRepository{
		pool,
	}
}

var updateSensorOwnerID int64 = 1

const saveSensorOwnerQuery = `INSERT INTO sensors_users (id, sensor_id, user_id) VALUES ($1, $2, $3)`

const getSensorsByUserID = `SELECT (sensor_id) FROM sensors_users WHERE user_id = $1`

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	_, err := r.pool.Exec(ctx, saveSensorOwnerQuery, updateSensorOwnerID, sensorOwner.SensorID, sensorOwner.UserID)
	if err != nil {
		return fmt.Errorf("can't save sensor owner: %w", err)
	}
	updateSensorOwnerID++
	return nil
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	rows, err := r.pool.Query(ctx, getSensorsByUserID, userID)
	if err != nil {
		return nil, fmt.Errorf("can't get sensors by user: %w", err)
	}
	defer rows.Close()
	var sensors []domain.SensorOwner
	for rows.Next() {
		sensor := domain.SensorOwner{}
		sensor.UserID = userID
		err = rows.Scan(&sensor.SensorID)
		if err != nil {
			return nil, fmt.Errorf("can't scan sensor owner: %w", err)
		}

		sensors = append(sensors, sensor)
	}
	return sensors, nil
}
