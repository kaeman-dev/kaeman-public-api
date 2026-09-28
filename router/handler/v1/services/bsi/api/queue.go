package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/response"
	"github.com/kaeman-dev/kaeman-public-api/utils"
)

type BSIType string

const (
	BSITypeChat   BSIType = "chat"
	BSITypeSplash BSIType = "splash"
)

type BSIPayload interface {
	BSIChatRequest | BSISplashRequest |
		BSIChatResponse | BSISplashResponse
}

type BSIChatRequest struct {
	Content string `json:"content"`
}

// 可以复用
type BSIChatResponse BSIChatRequest

type BSISplashRequest struct {
	Splasher string    `json:"splasher"`
	Time     time.Time `json:"time"`
	Lobby    string    `json:"lobby"`
	LobbyID  int64     `json:"lobbyID"`
	Location string    `json:"location,omitempty"`
	ServerID string    `json:"serverID,omitempty"`
	Note     string    `json:"note,omitempty"`
}

// 可以复用
type BSISplashResponse BSISplashRequest

type BSIMessage[T BSIPayload] struct {
	ID   string  `json:"id,omitempty"`
	Type BSIType `json:"type"`
	Data T       `json:"data"`
}

type BSIEnvelope struct {
	Type BSIType         `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (h Handler) Send(c *gin.Context) error {
	var env BSIEnvelope
	if err := response.Decode(c, &env); err != nil {
		return response.New("Invalid request body").WithStatus(http.StatusBadRequest).Write(c)
	}

	decode := func(target any) error {
		d := json.NewDecoder(bytes.NewReader(env.Data))
		d.DisallowUnknownFields()
		return d.Decode(target)
	}
	broadcast := func(msg any) error {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		h.Services.BSI.Broadcast(data)
		return response.NewData[any]("ok", nil).WithStatus(http.StatusAccepted).Write(c)
	}

	switch env.Type {
	case BSITypeChat:
		req := BSIMessage[BSIChatRequest]{Type: BSITypeChat}
		if err := decode(&req.Data); err != nil {
			return response.New("Invalid request body").WithStatus(http.StatusBadRequest).Write(c)
		}
		if strings.TrimSpace(req.Data.Content) == "" {
			return response.New("Message must not be blank").WithStatus(http.StatusBadRequest).Write(c)
		}
		req.ID = utils.NewUUID()
		return broadcast(req)
	case BSITypeSplash:
		req := BSIMessage[BSISplashRequest]{Type: BSITypeSplash}
		if err := decode(&req.Data); err != nil {
			return response.New("Invalid request body").WithStatus(http.StatusBadRequest).Write(c)
		}
		if req.Data.Time.IsZero() || strings.TrimSpace(req.Data.Lobby) == "" || req.Data.LobbyID < 1 {
			return response.New("Valid time, lobby and lobbyID are required").WithStatus(http.StatusBadRequest).Write(c)
		}
		req.ID = utils.NewUUID()
		return broadcast(req)
	default:
		return response.New("Unknown message kind").WithStatus(http.StatusBadRequest).Write(c)
	}
}
