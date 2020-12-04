package logic

import (
	"context"

	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/svc"
	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/types"

	"github.com/tal-tech/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) RegisterLogic {
	return RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req types.RegisterRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
