package svc

import (
	"github.com/valeamoris/go-ezio/example/rest/greeting/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
