package gogen

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/valeamoris/go-ezio/tools/ezioctl/vars"
	"path"
	"strings"
	"text/template"
	"unicode"

	"github.com/tal-tech/go-zero/tools/goctl/api/spec"
	apiutil "github.com/tal-tech/go-zero/tools/goctl/api/util"
	"github.com/tal-tech/go-zero/tools/goctl/config"
	"github.com/tal-tech/go-zero/tools/goctl/util"
	"github.com/tal-tech/go-zero/tools/goctl/util/format"
)

const handlerTemplate = `package handler

import (
	{{if .HasResp }}"net/http"{{end}}

	{{.ImportPackages}}
)

func {{.HandlerName}}(svcCtx *svc.ServiceContext) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := ctx.Bind(&req); err != nil {
			return err
		}{{end}}

		l := logic.New{{.LogicType}}(ctx.Request().Context(), svcCtx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}req{{end}})
		if err != nil {
			return err
		} else {
			{{if .HasResp}} return ctx.JSON(http.StatusOK, resp){{else}}return nil{{end}}
		}
	}
}
`

type Handler struct {
	ImportPackages string
	HandlerName    string
	RequestType    string
	LogicType      string
	Call           string
	HasResp        bool
	HasRequest     bool
}

func genHandler(dir string, cfg *config.Config, group spec.Group, route spec.Route) error {
	handler := getHandlerName(route)
	if getHandlerFolderPath(group, route) != handlerDir {
		handler = strings.Title(handler)
	}
	parentPkg, err := getParentPackage(dir)
	if err != nil {
		return err
	}

	return doGenToFile(dir, handler, cfg, group, route, Handler{
		ImportPackages: genHandlerImports(group, route, parentPkg),
		HandlerName:    handler,
		RequestType:    util.Title(route.RequestType.Name),
		LogicType:      strings.Title(getLogicName(route)),
		Call:           strings.Title(strings.TrimSuffix(handler, "Handler")),
		HasResp:        len(route.ResponseType.Name) > 0,
		HasRequest:     len(route.RequestType.Name) > 0,
	})
}

func doGenToFile(dir, handler string, cfg *config.Config, group spec.Group, route spec.Route, handleObj Handler) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, handler)
	if err != nil {
		return err
	}

	filename = filename + ".go"
	fp, created, err := apiutil.MaybeCreateFile(dir, getHandlerFolderPath(group, route), filename)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	text, err := util.LoadTemplate(category, handlerTemplateFile, handlerTemplate)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = template.Must(template.New("handlerTemplate").Parse(text)).Execute(buffer, handleObj)
	if err != nil {
		return err
	}

	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}

func genHandlers(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	for _, group := range api.Service.Groups {
		for _, route := range group.Routes {
			if err := genHandler(dir, cfg, group, route); err != nil {
				return err
			}
		}
	}

	return nil
}

func genHandlerImports(group spec.Group, route spec.Route, parentPkg string) string {
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s\"",
		util.JoinPackages(parentPkg, getLogicFolderPath(group, route))))
	imports = append(imports, fmt.Sprintf("\"%s\"", util.JoinPackages(parentPkg, contextDir)))
	if len(route.RequestType.Name) > 0 {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", util.JoinPackages(parentPkg, typesDir)))
	}
	imports = append(imports, fmt.Sprintf("\"%s\"", vars.EchoFrameworkUrl))

	return strings.Join(imports, "\n\t")
}

func getHandlerBaseName(route spec.Route) (string, error) {
	handler, ok := apiutil.GetAnnotationValue(route.Annotations, "server", "handler")
	if !ok {
		return "", fmt.Errorf("missing handler annotation for %q", route.Path)
	}

	for _, char := range handler {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) {
			return "", errors.New(fmt.Sprintf("route [%s] handler [%s] invalid, handler name should only contains letter or digit",
				route.Path, handler))
		}
	}

	handler = strings.TrimSpace(handler)
	handler = strings.TrimSuffix(handler, "handler")
	handler = strings.TrimSuffix(handler, "Handler")
	return handler, nil
}

func getHandlerFolderPath(group spec.Group, route spec.Route) string {
	folder, ok := apiutil.GetAnnotationValue(route.Annotations, "server", groupProperty)
	if !ok {
		folder, ok = apiutil.GetAnnotationValue(group.Annotations, "server", groupProperty)
		if !ok {
			return handlerDir
		}
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")
	return path.Join(handlerDir, folder)
}

func getHandlerName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	return handler + "Handler"
}

func getLogicName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	return handler + "Logic"
}
