package server

import (
	"log/slog"
	"net/http"
	"time"
)

func LoggerMiddleware(logger *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		receivedTime := time.Now()

		logger.Info("request received",
			"method", r.Method,
			"path", r.URL.Path,
		)

		next(w, r)

		logger.Info("request complete",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(receivedTime).Milliseconds(),
		)
	}
}
