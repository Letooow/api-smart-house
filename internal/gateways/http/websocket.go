package http

import (
	"context"
	"encoding/json"
	"errors"
	"homework/internal/domain"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	useCases    UseCases
	connections sync.Map
}

func NewWebSocketHandler(useCases UseCases) *WebSocketHandler {
	return &WebSocketHandler{
		useCases:    useCases,
		connections: sync.Map{},
	}
}

func (h *WebSocketHandler) Handle(c *gin.Context, id int64) (err error) {
	w, r := c.Writer, c.Request
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		return err
	}

	defer func(conn *websocket.Conn, code websocket.StatusCode, reason string) {
		err = conn.Close(code, reason)
		if err != nil {
			log.Printf("failed to close websocket connection: %v", err)
		}
	}(conn, websocket.StatusNormalClosure, "closed")

	h.connections.Store(id, conn)

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	ticker := time.NewTicker(time.Millisecond * 4000)
	defer ticker.Stop()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.checkConnection(ctx, conn, id, cancel)()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var events *domain.Event
				events, err = h.useCases.Event.GetLastEventBySensorID(ctx, id)
				if err != nil {
					continue
				}
				err = marshallAndWrite(ctx, events, conn)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}
	}()
	wg.Wait()
	return nil
}

func marshallAndWrite(ctx context.Context, events *domain.Event, conn *websocket.Conn) error {
	jsonEvent, err := json.Marshal(events)
	if err != nil {
		return err
	}
	err = conn.Write(ctx, websocket.MessageText, jsonEvent)
	return err
}

func (h *WebSocketHandler) checkConnection(ctx context.Context, conn *websocket.Conn, id int64, cancel context.CancelFunc) func() {
	return func() {
		for {
			_, _, err := conn.Reader(ctx)
			if err != nil {
				_ = conn.Close(websocket.StatusNormalClosure, "client disconnected")
				h.connections.Delete(id)
				cancel()
			}
		}
	}
}

func (h *WebSocketHandler) Shutdown() error {
	var err error
	h.connections.Range(func(_, value interface{}) bool {
		conn, ok := value.(*websocket.Conn)
		if !ok {
			return false
		}
		nErr := conn.Ping(context.Background())
		if err != nil {
			err = errors.Join(err, nErr)
			return false
		}
		nErr = conn.Close(websocket.StatusNormalClosure, "server shutting down")
		if nErr != nil {
			err = errors.Join(err, nErr)
		}
		return true
	})
	if err != nil {
		return err
	}
	return nil
}
