package logging

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
)

type Writer struct {
	Log   *slog.Logger
	Level slog.Level
}

var _ io.Writer = Writer{}

func (w Writer) Write(p []byte) (int, error) {
	for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
		if line == "" {
			continue
		}
		w.Log.Log(context.Background(), w.Level, line)
	}
	return len(p), nil
}

func RedirectGIN(l *slog.Logger) {
	log.SetFlags(0)
	log.SetOutput(Writer{Log: l, Level: slog.LevelInfo})

	gin.DebugPrintRouteFunc = func(method, path, handler string, handlers int) {
		l.Info("route registered", "method", method, "path", path, "handler", handler, "handlers", handlers)
	}
	gin.DebugPrintFunc = func(format string, values ...any) {
		message := strings.Join(strings.Fields(fmt.Sprintf(format, values...)), " ")
		level := slog.LevelInfo
		// gin 的 debugPrint 只会在前面加一个 [WARNING]
		// debug.go:83/87/93/101、gin.go:544、response_writer.go:70
		// [ERROR] 由 debugPrintError 直接写 DefaultErrorWriter（debug.go:112），不经过这里。
		if tag, rest, ok := strings.Cut(message, "]"); ok && strings.HasPrefix(tag, "[") {
			switch strings.TrimPrefix(strings.ToUpper(tag), "[") {
			case "WARNING":
				level = slog.LevelWarn
			}
			message = strings.TrimSpace(rest)
		}
		l.Log(context.Background(), level, message)
	}
	gin.DefaultWriter = Writer{Log: l, Level: slog.LevelInfo}
	gin.DefaultErrorWriter = Writer{Log: l, Level: slog.LevelError}
}
