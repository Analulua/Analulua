package application

import (
	"newdemo1/infrastructure"
	"newdemo1/resource"
)

type Application struct {
}

func NewApplication(resource *resource.Resource, infrastructure *infrastructure.Infrastructure) (*Application, error) {
	return &Application{}, nil
}
