package integration

import (
	"net/http"

	"github.com/go-kit/log/level"
	"github.com/kakkoyun/go-tool-dependency-demo/pkg/logger"
)

func LoggerMiddleware() func(http.Handler) http.Handler {
	logger := logger.NewLogger("info", logger.LogFormatLogfmt, "middleware")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			level.Info(logger).Log("msg", "request received", "method", r.Method, "url", r.URL.String())
			next.ServeHTTP(w, r)
		})
	}
}
