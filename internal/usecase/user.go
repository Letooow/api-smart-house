package usecase

import (
	"context"
	"homework/internal/domain"
)

type User struct {
	userRepository        UserRepository
	sensorOwnerRepository SensorOwnerRepository
	sensorRepository      SensorRepository
}

func NewUser(ur UserRepository, sor SensorOwnerRepository, sr SensorRepository) *User {
	return &User{
		userRepository:        ur,
		sensorOwnerRepository: sor,
		sensorRepository:      sr,
	}
}

func (u *User) RegisterUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if u.userRepository == nil || user.Name == "" {
		return nil, ErrInvalidUserName
	}
	err := u.userRepository.SaveUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) AttachSensorToUser(ctx context.Context, userID, sensorID int64) error {
	if _, err := u.userRepository.GetUserByID(ctx, userID); err != nil {
		return err
	}

	if u.sensorRepository == nil {
		return ErrUserNotFound
	}
	if _, err := u.sensorRepository.GetSensorByID(ctx, sensorID); err != nil {
		return err
	}
	err := u.sensorOwnerRepository.SaveSensorOwner(ctx, domain.SensorOwner{UserID: userID, SensorID: sensorID})
	if err != nil {
		return err
	}
	return nil
}

func (u *User) GetUserSensors(ctx context.Context, userID int64) ([]domain.Sensor, error) {
	if u.userRepository == nil {
		return nil, ErrInvalidUserName
	}
	if _, err := u.userRepository.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	if u.sensorOwnerRepository == nil {
		return nil, ErrUserNotFound
	}
	sensorsOwnerByUserID, err := u.sensorOwnerRepository.GetSensorsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	sensors := make([]domain.Sensor, 0, len(sensorsOwnerByUserID))
	for _, sensorOwner := range sensorsOwnerByUserID {
		sensor, err := u.sensorRepository.GetSensorByID(ctx, sensorOwner.SensorID)
		if err != nil {
			return nil, err
		}
		sensors = append(sensors, *sensor)
	}
	return sensors, nil
}
