package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kaeman-dev/kaeman-public-api/utils"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	prompt := func(label string) string {
		fmt.Print(label)
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return strings.TrimSpace(line)
	}

	accessToken := prompt("access token: ")
	serverID := prompt("server id: ")
	if accessToken == "" || serverID == "" {
		fmt.Fprintln(os.Stderr, "error: access token and server id are required")
		os.Exit(1)
	}

	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		fmt.Fprintln(os.Stderr, "error: access token is not a JWT")
		os.Exit(1)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	var claims struct {
		Profiles struct {
			MC string `json:"mc"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if claims.Profiles.MC == "" {
		fmt.Fprintln(os.Stderr, "error: access token carries no minecraft profile")
		os.Exit(1)
	}

	profile, err := utils.NormalizeUUID(claims.Profiles.MC)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	body, err := json.Marshal(map[string]string{
		"accessToken":     accessToken,
		"selectedProfile": profile,
		"serverId":        serverID,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	req, err := http.NewRequest(http.MethodPost, "https://sessionserver.mojang.com/session/minecraft/join", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "error: POST session/minecraft/join: HTTP %d %s\n", resp.StatusCode, strings.TrimSpace(string(response)))
		os.Exit(1)
	}

	fmt.Println("joined:", profile)
}
