package middleware

import (
	"github.com/labstack/echo/v4"
)

type UserCheckMiddleware struct {
}

func NewUserCheckMiddleware() *UserCheckMiddleware {
	return &UserCheckMiddleware{}
}

func (m *UserCheckMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return next(ctx)
	}
}
