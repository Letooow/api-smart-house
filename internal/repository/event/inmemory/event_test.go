package inmemory

import (
	"context"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventRepository_SaveEvent(t *testing.T) {
	t.Run("err, event is nil", func(t *testing.T) {
		er := NewEventRepository()
		err := er.SaveEvent(context.Background(), nil)
		assert.Error(t, err)
	})

	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := er.SaveEvent(ctx, &domain.Event{})
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		err := er.SaveEvent(ctx, &domain.Event{})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("ok, save and get one", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		event := &domain.Event{
			Timestamp:          time.Now(),
			SensorSerialNumber: "0123456789",
			Payload:            0,
		}

		err := er.SaveEvent(ctx, event)
		assert.NoError(t, err)

		actualEvent, err := er.GetLastEventBySensorID(ctx, event.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, event.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, event.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, event.Payload, actualEvent.Payload)
	})

	t.Run("ok, collision test", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg := sync.WaitGroup{}
		var lastEvent domain.Event
		for i := 0; i < 1000; i++ {
			event := &domain.Event{
				Timestamp:          time.Now(),
				SensorSerialNumber: "0123456789",
				Payload:            0,
			}
			lastEvent = *event
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.NoError(t, er.SaveEvent(ctx, event))
			}()
		}

		wg.Wait()

		actualEvent, err := er.GetLastEventBySensorID(ctx, lastEvent.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, lastEvent.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, lastEvent.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, lastEvent.Payload, actualEvent.Payload)
	})
}

func TestEventRepository_GetLastEventBySensorID(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := er.GetLastEventBySensorID(ctx, 0)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		_, err := er.GetLastEventBySensorID(ctx, 0)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("fail, event not found", func(t *testing.T) {
		er := NewEventRepository()
		_, err := er.GetLastEventBySensorID(context.Background(), 234)
		assert.ErrorIs(t, err, usecase.ErrEventNotFound)
	})

	t.Run("ok, save and get one", func(t *testing.T) {
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sensorID := int64(12345)
		var lastEvent *domain.Event
		for i := 0; i < 10; i++ {
			lastEvent = &domain.Event{
				Timestamp: time.Now(),
				SensorID:  sensorID,
				Payload:   0,
			}
			time.Sleep(10 * time.Millisecond)
			assert.NoError(t, er.SaveEvent(ctx, lastEvent))
		}

		for i := 0; i < 10; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  54321,
				Payload:   0,
			}
			assert.NoError(t, er.SaveEvent(ctx, event))
		}

		actualEvent, err := er.GetLastEventBySensorID(ctx, lastEvent.SensorID)
		assert.NoError(t, err)
		assert.NotNil(t, actualEvent)
		assert.Equal(t, lastEvent.Timestamp, actualEvent.Timestamp)
		assert.Equal(t, lastEvent.SensorSerialNumber, actualEvent.SensorSerialNumber)
		assert.Equal(t, lastEvent.Payload, actualEvent.Payload)
	})
}

func TestEventRepository_GetEventsBySensorIDWithDate(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		t.Parallel()
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := er.GetEventsBySensorIDWithDate(ctx, 0, time.Now(), time.Now().Add(time.Hour))
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		t.Parallel()
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		_, err := er.GetEventsBySensorIDWithDate(ctx, 0, time.Now(), time.Now().Add(time.Hour))
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("fail, event not found", func(t *testing.T) {
		t.Parallel()
		er := NewEventRepository()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := er.GetEventsBySensorIDWithDate(ctx, 0, time.Now(), time.Now())
		assert.ErrorIs(t, err, usecase.ErrEventNotFound)
	})
	t.Run("ok, save and get one", func(t *testing.T) {
		t.Parallel()
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		now := time.Now()
		event := &domain.Event{
			Timestamp:          time.Now(),
			SensorSerialNumber: "0123456789",
			Payload:            0,
		}
		events := []domain.Event{*event}
		_ = er.SaveEvent(ctx, event)

		actualEvent, err := er.GetEventsBySensorIDWithDate(ctx, event.SensorID, now, time.Now())
		assert.NoError(t, err)
		assert.Equal(t, events, actualEvent)
	})
	t.Run("ok, save and get some", func(t *testing.T) {
		t.Parallel()
		er := NewEventRepository()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		now := time.Now()
		var events []domain.Event
		for i := 0; i < 100; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  int64(12345),
				Payload:   0,
			}
			_ = er.SaveEvent(ctx, event)
			events = append(events, *event)
		}
		res, err := er.GetEventsBySensorIDWithDate(ctx, 12345, now, time.Now())
		assert.NoError(t, err)
		assert.Equal(t, events, res)
	})
}

func FuzzTestGetEventsBySensorIDWithDate(f *testing.F) {
	for i := int64(0); i < 100; i++ {
		f.Add(i)
	}

	f.Fuzz(func(t *testing.T, b int64) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := NewEventRepository()
		for i := 0; i < 100; i++ {
			event := &domain.Event{
				Timestamp: time.Now(),
				SensorID:  int64(12345),
				Payload:   0,
			}
			_ = er.SaveEvent(ctx, event)
		}
		_, err := er.GetEventsBySensorIDWithDate(ctx, b, time.Now().Add(-1*time.Minute), time.Now())
		assert.Error(t, err)
	})
}
