package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kaeman-dev/kaeman-public-api/response"
)

type BSIQueue struct {
	Kind    string    `json:"kind"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
	Lobby   string    `json:"lobby"`
	LobbyID int       `json:"lobbyID"`
}

func (h Handler) Send(c *gin.Context) error {
	input := BSIQueue{}
	if err := response.Decode(c, &input); err != nil {
		return response.New("Invalid request body").WithStatus(http.StatusBadRequest).Write(c)
	}
	switch input.Kind {
	case "chat":
		if strings.TrimSpace(input.Content) == "" {
			return response.New("Message must not be blank").WithStatus(http.StatusBadRequest).Write(c)
		}
	case "splash":
		if input.Time.IsZero() || strings.TrimSpace(input.Lobby) == "" || input.LobbyID < 1 {
			return response.New("Valid time, lobby and lobbyID are required").WithStatus(http.StatusBadRequest).Write(c)
		}
	default:
		return response.New("Unknown message kind").WithStatus(http.StatusBadRequest).Write(c)
	}
	return response.NewData[any]("ok", nil).WithStatus(http.StatusAccepted).Write(c)
}
