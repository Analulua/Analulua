package mq

import (
	"newdemo1/infrastructure/mq/pubsub1"
	"newdemo1/resource"
)

type (
	PubSub interface {
		PubSub() pubsub1.Client
	}
	MQ struct {
		resource *resource.Resource
		pubsub   pubsub1.Client
	}
)

func NewMQ(resource *resource.Resource) (PubSub, error) {
	pubsub, err := pubsub1.New(resource)
	if err != nil {
		return nil, err
	}
	return &MQ{
		resource: resource,
		pubsub:   pubsub,
	}, nil
}

func (m *MQ) PubSub() pubsub1.Client {
	return m.pubsub
}
