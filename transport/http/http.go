package http

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"newdemo1/application"
	"newdemo1/resource"
	commonHttp "newdemo1/resource/jaeger/common/http"
	"newdemo1/transport/http/controller"
)

type Http struct {
	resource   *resource.Resource
	app        *application.Application
	controller *controller.Controller
}

func NewHttp(resource *resource.Resource, app *application.Application) Http {
	return Http{
		resource:   resource,
		app:        app,
		controller: controller.NewController(resource, app),
	}
}

func (h *Http) Serve() error {
	g := gin.Default()
	g.Any("/", func(context *gin.Context) {
		serviceName, _ := json.Marshal(h.resource.Config.Telemetry.Tracer.ServiceName)
		_, _ = context.Writer.Write(serviceName)
	})
	subscription := g.Group("/subscription")
	{
		subscription.GET("/:id", h.controller.Subscription.Create)
	}
	log.Println("[Recurring Service HTTP] server started. Listening on port ", h.resource.Config.Service.HttpPort)
	return http.ListenAndServe(h.resource.Config.Service.HttpPort, commonHttp.NewHandler(
		g,
		commonHttp.WithTelemetry(h.resource.Jaeger.Tracer),
		commonHttp.WithHealthCheckPath(
			"/health",
		),
		commonHttp.WithDefault(),
	))
}
