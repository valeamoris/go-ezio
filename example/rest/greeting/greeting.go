package main

import (
	"flag"
	"github.com/tal-tech/go-zero/core/conf"
	"github.com/valeamoris/go-ezio/example/rest/greeting/internal/config"
	"github.com/valeamoris/go-ezio/example/rest/greeting/internal/handler"
	"github.com/valeamoris/go-ezio/example/rest/greeting/internal/svc"
	"github.com/valeamoris/go-ezio/rest"
)

var configFile = flag.String("f", "etc/greet-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)
	server.Start()
}
