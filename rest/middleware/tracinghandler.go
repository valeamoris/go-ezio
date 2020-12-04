package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
	"io"
	"log"
	"net/http"
)

const defaultComponentName = "echo/v4"

func TracingMiddleware(serviceName string) (echo.MiddlewareFunc, io.Closer) {
	// Recommended configuration for production.
	cfg := jaegercfg.Configuration{}

	// Zipkin shares span ID between client and server spans; it must be enabled via the following option.
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	// Create tracer and then initialize global tracer
	closer, err := cfg.InitGlobalTracer(
		serviceName,
		jaegercfg.Logger(jaeger.StdLogger),
		jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.ZipkinSharedRPCSpan(true),
	)

	if err != nil {
		log.Panicf("Could not initialize jaeger tracer: %s", err.Error())
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			opname := "HTTP " + req.Method + " URL: " + c.Path()
			var sp opentracing.Span
			if ctx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
				sp = opentracing.StartSpan(opname)
			} else {
				sp = opentracing.StartSpan(opname, ext.RPCServerOption(ctx), opentracing.ChildOf(ctx))
			}

			ext.HTTPMethod.Set(sp, req.Method)
			ext.HTTPUrl.Set(sp, req.URL.String())
			ext.Component.Set(sp, defaultComponentName)
			req = req.WithContext(opentracing.ContextWithSpan(req.Context(), sp))
			c.SetRequest(req)

			defer func() {
				status := c.Response().Status
				committed := c.Response().Committed
				ext.HTTPStatusCode.Set(sp, uint16(status))
				if status >= http.StatusInternalServerError || !committed {
					ext.Error.Set(sp, true)
				}
				sp.Finish()
			}()
			return next(c)
		}
	}, closer
}
