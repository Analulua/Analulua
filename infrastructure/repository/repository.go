package repository

import (
	"newdemo1/infrastructure/client"
	"newdemo1/resource"
)

type Repository struct {
	c *client.Client
}

func NewRepository(resource *resource.Resource, clt *client.Client) (*Repository, error) {
	return &Repository{
		c: clt,
	}, nil
}
