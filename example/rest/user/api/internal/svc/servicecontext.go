package svc

import (
	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/config"
	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/middleware"
	"github.com/valeamoris/go-ezio/rest"
)

type ServiceContext struct {
	Config    config.Config
	UserCheck rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		UserCheck: middleware.NewUserCheckMiddleware().Handle,
	}
}
