package domain

import "time"

// Event - структура события по датчику
type Event struct {
	// Timestamp - время события
	Timestamp time.Time
	// SensorSerialNumber - серийный номер датчика
	SensorSerialNumber string `json:"sensor_serial_number" validate:"len:10"`
	// SensorID - id датчика
	SensorID int64 `json:"sensor_id,omitempty"`
	// Payload - данные события
	Payload int64 `json:"payload"`
}
