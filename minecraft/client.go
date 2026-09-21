package minecraft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kaeman-dev/kaeman-public-api/model"
)

type Client struct {
	HTTP *http.Client

	baseURL string
}

func New(baseURL string) *Client {
	return &Client{HTTP: &http.Client{Timeout: 10 * time.Second}, baseURL: strings.TrimRight(baseURL, "/")}
}

func (m *Client) Profile(ctx context.Context, id string) (model.MinecraftIdentity, error) {
	var identity model.MinecraftIdentity
	if !model.ValidMinecraftUUID(id) {
		return identity, errors.New("invalid Minecraft UUID")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+"/session/minecraft/profile/"+id, nil)
	if err != nil {
		return identity, err
	}
	resp, err := m.HTTP.Do(req)
	if err != nil {
		return identity, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return identity, fmt.Errorf("Mojang profile HTTP status: %d", resp.StatusCode)
	}
	var profile struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&profile); err != nil {
		return identity, err
	}
	if !strings.EqualFold(profile.ID, id) || !model.ValidMinecraftName(profile.Name) {
		return identity, errors.New("invalid Mojang profile identity")
	}
	return model.MinecraftIdentity{UUID: id, Name: profile.Name}, nil
}

func (m *Client) HasJoined(ctx context.Context, username, serverID, uuid string) (bool, error) {
	u, err := url.Parse(m.baseURL + "/session/minecraft/hasJoined")
	if err != nil {
		return false, fmt.Errorf("invalid session URL: %w", err)
	}
	query := u.Query()
	query.Set("username", username)
	query.Set("serverId", serverID)
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return false, err
	}
	resp, err := m.HTTP.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNoContent, http.StatusForbidden, http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("Mojang hasJoined HTTP status: %d", resp.StatusCode)
	}
	var profile struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&profile); err != nil {
		return false, fmt.Errorf("invalid Mojang response: %w", err)
	}
	if !model.ValidMinecraftName(profile.Name) || !strings.EqualFold(profile.ID, uuid) || !strings.EqualFold(profile.Name, username) {
		return false, nil
	}
	return true, nil
}
