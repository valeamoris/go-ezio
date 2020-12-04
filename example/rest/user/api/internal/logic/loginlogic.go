package logic

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"time"

	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/svc"
	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/types"

	"github.com/tal-tech/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) LoginLogic {
	return LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req types.LoginRequest) (*types.UserResponse, error) {
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	jwtToken, err := l.generateJwtToken(l.svcCtx.Config.Auth.AccessSecret, now, accessExpire)
	if err != nil {
		return nil, err
	}
	return &types.UserResponse{
		JwtToken: types.JwtToken{
			AccessToken:  jwtToken,
			AccessExpire: now + accessExpire,
			RefreshAfter: now + accessExpire/2,
		},
	}, nil
}

func (l *LoginLogic) generateJwtToken(secret string, iat, seconds int64) (string, error) {
	claims := jwt.StandardClaims{
		ExpiresAt: iat + seconds,
		IssuedAt:  iat,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
