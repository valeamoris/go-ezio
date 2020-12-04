package gogen

import (
	"bytes"
	"fmt"
	"github.com/valeamoris/go-ezio/tools/ezioctl/vars"
	"strings"
	"text/template"

	"github.com/tal-tech/go-zero/tools/goctl/api/spec"
	"github.com/tal-tech/go-zero/tools/goctl/api/util"
	"github.com/tal-tech/go-zero/tools/goctl/config"
	"github.com/tal-tech/go-zero/tools/goctl/util/format"
)

var middlewareImplementCode = `
package middleware

import (
	{{.ImportPackages}}
)

type {{.name}} struct {
}

func New{{.name}}() *{{.name}} {	
	return &{{.name}}{}
}

func (m *{{.name}})Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		// TODO generate middleware implement function, delete after code implementation

		// Passthrough to next handler if need 
		return next(ctx)
	}	
}
`

func genMiddleware(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	var middlewares = getMiddleware(api)
	for _, item := range middlewares {
		middlewareFilename := strings.TrimSuffix(strings.ToLower(item), "middleware") + "_middleware"
		formatName, err := format.FileNamingFormat(cfg.NamingFormat, middlewareFilename)
		if err != nil {
			return err
		}

		filename := formatName + ".go"
		fp, created, err := util.MaybeCreateFile(dir, middlewareDir, filename)
		if err != nil {
			return err
		}
		if !created {
			return nil
		}
		defer fp.Close()

		name := strings.TrimSuffix(item, "Middleware") + "Middleware"
		t := template.Must(template.New("contextTemplate").Parse(middlewareImplementCode))
		buffer := new(bytes.Buffer)
		err = t.Execute(buffer, map[string]string{
			"name":           strings.Title(name),
			"ImportPackages": fmt.Sprintf("\"%s\"", vars.EchoFrameworkUrl),
		})
		if err != nil {
			return err
		}

		formatCode := formatCode(buffer.String())
		_, err = fp.WriteString(formatCode)
		return err
	}
	return nil
}
