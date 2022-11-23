package store

import (
	"newdemo1/infrastructure/client"
	"newdemo1/infrastructure/repository"
	"newdemo1/resource"
)

type Store struct {
	Repository *repository.Repository
}

func NewStore(resource *resource.Resource) (*Store, error) {
	storeClient, err := client.NewClient(resource)
	if err != nil {
		return nil, err
	}

	repo, err := repository.NewRepository(resource, storeClient)
	if err != nil {
		return nil, err
	}

	return &Store{Repository: repo}, nil
}
