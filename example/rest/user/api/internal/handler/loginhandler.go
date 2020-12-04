package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/logic"
	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/svc"
	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/types"
)

func loginHandler(srvCtx *svc.ServiceContext) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var req types.LoginRequest
		if err := ctx.Bind(&req); err != nil {
			return err
		}

		l := logic.NewLoginLogic(ctx.Request().Context(), srvCtx)
		resp, err := l.Login(req)
		if err != nil {
			return err
		} else {
			return ctx.JSON(http.StatusOK, resp)
		}
	}
}
