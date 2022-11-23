package subscription

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"newdemo1/application"
	"newdemo1/resource"
)

type Controller interface {
	Create(g *gin.Context)
}

type controller struct {
	tracerOpsPrefix string
	resource        *resource.Resource
	app             *application.Application
}

func (c *controller) Create(g *gin.Context) {
	g.JSON(http.StatusOK, gin.H{"message": "success"})
}

func NewController(resource *resource.Resource, app *application.Application) Controller {
	return &controller{
		tracerOpsPrefix: "transport/http/controller/subscription/controller.go",
		resource:        resource,
		app:             app,
	}
}
