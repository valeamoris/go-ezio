package handler

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/valeamoris/go-ezio/rest"
	"net/http"

	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/svc"
)

func RegisterHandlers(engine *rest.Server, serverCtx *svc.ServiceContext) {
	engine.Group(
		rest.Group{
			Prefix: "/",
			Routes: []rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "ping",
					Handler: pingHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "register",
					Handler: registerHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "login",
					Handler: loginHandler(serverCtx),
				},
			},
		},
	)

	engine.Group(
		rest.Group{
			Prefix: "/user/",
			Routes: []rest.Route{
				{
					Method:  http.MethodGet,
					Path:    "info",
					Handler: userInfoHandler(serverCtx),
				},
			},
		},
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.UserCheck}...,
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret, jwt.MapClaims{}),
	)
}
