package handler

import (
	"github.com/valeamoris/go-ezio/example/rest/greeting/internal/svc"
	"github.com/valeamoris/go-ezio/rest"
	"net/http"
)

func RegisterHandlers(engine *rest.Server, svcCtx *svc.ServiceContext) {
	g := rest.Group{
		Prefix: "",
		Routes: []rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/from/:name",
				Handler: GreetHandler(svcCtx),
			},
		},
	}
	engine.Group(g)
}
