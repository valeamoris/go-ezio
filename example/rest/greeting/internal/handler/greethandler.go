package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/valeamoris/go-ezio/example/rest/greeting/internal/logic"
	"github.com/valeamoris/go-ezio/example/rest/greeting/internal/svc"
	"github.com/valeamoris/go-ezio/example/rest/greeting/internal/types"
	"net/http"
)

func GreetHandler(svcCtx *svc.ServiceContext) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		req := new(types.Request)
		if err := ctx.Bind(req); err != nil {
			return err
		}
		//if err := ctx.Validate(req); err != nil {
		//	return err
		//}

		l := logic.NewGreetLogic(ctx, svcCtx)
		resp, err := l.Greet(req)
		if err != nil {
			return err
		} else {
			return ctx.JSON(http.StatusOK, resp)
		}
	}
}
