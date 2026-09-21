package gateway

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/response"
)

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]chan []byte
	closed  bool
}
type upgradeWriter struct {
	gin.ResponseWriter
	context *gin.Context
	failed  bool
}

func (w *upgradeWriter) WriteHeader(status int) {
	if status >= 400 {
		w.failed = true
		_ = w.context.Error(response.Error(status, http.StatusText(status)))
		return
	}
	w.ResponseWriter.WriteHeader(status)
}
func (w *upgradeWriter) Write(data []byte) (int, error) {
	if w.failed {
		return len(data), nil
	}
	return w.ResponseWriter.Write(data)
}
func (w *upgradeWriter) WriteString(data string) (int, error) {
	if w.failed {
		return len(data), nil
	}
	return w.ResponseWriter.WriteString(data)
}

func (h *Hub) Listen(c *gin.Context) {
	conn, err := websocket.Accept(&upgradeWriter{ResponseWriter: c.Writer, context: c}, c.Request, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		_ = c.Error(err)
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(1024)
	ctx := conn.CloseRead(c.Request.Context())
	queue := make(chan []byte, 32)
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	if h.clients == nil {
		h.clients = make(map[*websocket.Conn]chan []byte)
	}
	h.clients[conn] = queue
	h.mu.Unlock()
	defer func() { h.mu.Lock(); delete(h.clients, conn); h.mu.Unlock() }()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-queue:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Write(writeCtx, websocket.MessageText, data)
			cancel()
			if err != nil {
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
func (h *Hub) Broadcast(data []byte) {
	h.mu.Lock()
	var slow []*websocket.Conn
	for conn, queue := range h.clients {
		select {
		case queue <- data:
		default:
			close(queue)
			delete(h.clients, conn)
			slow = append(slow, conn)
		}
	}
	h.mu.Unlock()
	for _, conn := range slow {
		conn.CloseNow()
	}
}
func (h *Hub) Close() {
	h.mu.Lock()
	h.closed = true
	connections := make([]*websocket.Conn, 0, len(h.clients))
	for conn, queue := range h.clients {
		close(queue)
		connections = append(connections, conn)
		delete(h.clients, conn)
	}
	h.mu.Unlock()
	for _, conn := range connections {
		conn.CloseNow()
	}
}
