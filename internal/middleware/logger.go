package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type LoggerMiddleware struct {
	logger *zap.Logger
}

func NewLoggerMiddleware(logger *zap.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{logger: logger}
}

func (m *LoggerMiddleware) Log(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		latency := time.Since(start)
		req := c.Request()
		res := c.Response()

		m.logger.Info("http request",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path),
			zap.Int("status", res.Status),
			zap.String("remote_ip", c.RealIP()),
			zap.Duration("latency", latency),
		)

		return err
	}
}
