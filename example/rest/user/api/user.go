package main

import (
	"flag"
	"fmt"
	"github.com/valeamoris/go-ezio/rest"

	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/config"
	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/handler"
	"github.com/valeamoris/go-ezio/example/rest/user/api/internal/svc"

	"github.com/tal-tech/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/user-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
