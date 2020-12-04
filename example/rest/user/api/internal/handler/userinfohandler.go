package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/logic"
	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/svc"
)

func userInfoHandler(srvCtx *svc.ServiceContext) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		l := logic.NewUserInfoLogic(ctx.Request().Context(), srvCtx)
		resp, err := l.UserInfo()
		if err != nil {
			return err
		} else {
			return ctx.JSON(http.StatusOK, resp)
		}
	}
}
