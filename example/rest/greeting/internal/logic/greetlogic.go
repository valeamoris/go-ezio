package logic

import (
	"github.com/labstack/echo/v4"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/valeamoris/go-ezio/example/rest/greeting/internal/svc"
	"github.com/valeamoris/go-ezio/example/rest/greeting/internal/types"
)

type GreetLogic struct {
	logx.Logger
	ctx    echo.Context
	svcCtx *svc.ServiceContext
}

func NewGreetLogic(ctx echo.Context, svcCtx *svc.ServiceContext) *GreetLogic {
	return &GreetLogic{
		Logger: logx.WithContext(ctx.Request().Context()),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GreetLogic) Greet(req *types.Request) (*types.Response, error) {
	return &types.Response{Message: req.Name}, nil
}
