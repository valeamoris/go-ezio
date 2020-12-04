package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/logic"
	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/svc"
)

func pingHandler(srvCtx *svc.ServiceContext) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		l := logic.NewPingLogic(ctx.Request().Context(), srvCtx)
		err := l.Ping()
		if err != nil {
			return err
		} else {
			return ctx.JSON(http.StatusOK, nil)
		}
	}
}
