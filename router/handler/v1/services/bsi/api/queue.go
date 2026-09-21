package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kaeman-dev/kaeman-public-api/jwt"
	"github.com/kaeman-dev/kaeman-public-api/model"
	"github.com/kaeman-dev/kaeman-public-api/response"
)

type BSIQueue struct {
	Kind string `json:"kind"`

	Content string `json:"content"`

	Time     time.Time `json:"time"`
	Lobby    string    `json:"lobby"`
	LobbyID  int       `json:"lobbyID"`
	ServerID string    `json:"serverID"`
	Location string    `json:"location"`
	Note     string    `json:"note"`
}

func (h Handler) Send(c *gin.Context) error {
	player := c.MustGet("minecraftJWT").(*jwt.Claims).Data.Minecraft
	var input BSIQueue
	if err := response.Decode(c, &input); err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	var data gin.H
	switch input.Kind {
	case "chat":
		if strings.TrimSpace(input.Content) == "" {
			return response.NewError(http.StatusBadRequest, "Message must not be blank")
		}
		data = gin.H{"type": "irc.recv", "kind": "chat", "rawContent": gin.H{
			"text": "BSIRC > ", "color": "light_purple",
			"extra": []gin.H{
				gin.H{
					"text": player.Name, "bold": true,
					"clickEvent": gin.H{"action": "suggest_command", "value": "/msg " + player.Name + " "},
					"hoverEvent": gin.H{"action": "show_text", "contents": gin.H{"text": "点击私聊", "color": "gray"}},
				},
				{"text": ": " + input.Content, "color": "white"},
			},
		}}
	case "splash":
		if _, ok := c.Get("publicJWT"); !ok {
			return response.NewError(http.StatusUnauthorized, "Public-API-Key header required")
		}
		if input.Time.IsZero() || strings.TrimSpace(input.Lobby) == "" || input.LobbyID < 1 {
			return response.NewError(http.StatusBadRequest, "Valid time, lobby and lobbyID are required")
		}
		id := uuid.NewString()
		splash := gin.H{"id": id, "time": input.Time, "lobby": input.Lobby, "lobbyID": input.LobbyID, "minecraft": player}
		if input.ServerID != "" {
			splash["serverID"] = input.ServerID
		}
		if input.Location != "" {
			splash["location"] = input.Location
		}
		if input.Note != "" {
			splash["note"] = input.Note
		}
		hover := input.Location
		if input.Note != "" {
			if hover != "" {
				hover += "\n"
			}
			hover += input.Note
		}
		destination := input.Lobby + "-" + strconv.Itoa(input.LobbyID)
		if input.ServerID != "" {
			destination += " (" + input.ServerID + ")"
		}
		text := " 将于 " + input.Time.Format("2006-01-02 15:04") + " 在 " + destination + " 喷药"
		if input.Location != "" {
			text += "，地点：" + input.Location
		}
		data = gin.H{"type": "irc.recv", "kind": "splash", "rawContent": gin.H{
			"text": "SPLASH > ", "color": "gold",
			"extra": []gin.H{
				{"text": player.Name, "bold": true, "hoverEvent": gin.H{"action": "show_text", "contents": gin.H{"text": hover, "color": "gray"}}},
				{"text": text, "color": "yellow"},
			},
		}, "splash": splash}
	default:
		return response.NewError(http.StatusBadRequest, "Unknown message kind")
	}
	payload, err := json.Marshal(gin.H{"msg": "ok", "data": data})
	if err != nil {
		return response.NewError(http.StatusInternalServerError, "Cannot encode message")
	}
	eventID := uuid.NewString()
	if id, ok := data["splash"]; ok {
		eventID = id.(gin.H)["id"].(string)
	}
	if err := h.Services.DB.WithContext(c.Request.Context()).Create(&model.Event{ID: eventID, Payload: payload}).Error; err != nil {
		return response.NewError(http.StatusServiceUnavailable, "Cannot enqueue broadcast")
	}
	return response.NewData[any]("ok", nil).WithStatus(http.StatusAccepted).Write(c)
}
