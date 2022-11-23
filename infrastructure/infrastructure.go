package infrastructure

import (
	"newdemo1/infrastructure/mq"
	"newdemo1/infrastructure/store"
	"newdemo1/infrastructure/sync"
	"newdemo1/resource"
)

type Infrastructure struct {
	Store *store.Store
	Sync  sync.Sync
	MQ    mq.PubSub
}

func NewInfrastructure(resource *resource.Resource) (*Infrastructure, error) {
	infras, err := store.NewStore(resource)
	if err != nil {
		return nil, err
	}

	mq, err := mq.NewMQ(resource)
	if err != nil {
		return nil, err
	}
	sc := sync.New(resource)

	return &Infrastructure{
		Store: infras,
		MQ:    mq,
		Sync:  sc,
	}, nil
}
