package domain

import "time"

type SensorType string

const (
	SensorTypeContactClosure SensorType = "cc"
	SensorTypeADC            SensorType = "adc"
)

// Sensor - структура для хранения данных датчика
type Sensor struct {
	// ID - id датчика
	ID int64 `json:"sensor_id,omitempty" validate:"min=1"`
	// SerialNumber - серийный номер датчика
	SerialNumber string `json:"serial_number,omitempty" validate:"len:10"`
	// Type - тип датчика
	Type SensorType `json:"type,omitempty"`
	// CurrentState - текущее состояние датчика
	CurrentState int64 `json:"current_state,omitempty"`
	// Description - описание датчика
	Description string `json:"description,omitempty"`
	// IsActive - активен ли датчик
	IsActive bool `json:"is_active,omitempty"`
	// RegisteredAt - дата регистрации датчика
	RegisteredAt time.Time
	// LastActivity - дата последнего изменения состояния датчика
	LastActivity time.Time
}
