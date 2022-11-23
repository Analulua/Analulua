package controller

import (
	"newdemo1/application"
	"newdemo1/resource"
	"newdemo1/transport/http/controller/subscription"
)

type Controller struct {
	Subscription subscription.Controller
}

func NewController(resource *resource.Resource, app *application.Application) *Controller {
	return &Controller{
		Subscription: subscription.NewController(resource, app),
	}
}
