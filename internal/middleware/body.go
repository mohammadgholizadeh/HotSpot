package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type RequestBodyMiddleware struct {
	LimitBytes int64
}

func NewRequestBodyMiddleware(limitBytes int64) *RequestBodyMiddleware {
	return &RequestBodyMiddleware{LimitBytes: limitBytes}
}

func (m *RequestBodyMiddleware) LimitBodySize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if m.LimitBytes > 0 {
			r := c.Request()
			w := c.Response()
			r.Body = http.MaxBytesReader(w, r.Body, m.LimitBytes)
		}
		return next(c)
	}
}
