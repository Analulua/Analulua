package pubsub1

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/base64"
	"google.golang.org/api/option"
	"newdemo1/resource"
	"newdemo1/resource/jaeger/common/tracer"
)

type (
	Client interface {
		Publish(ctx context.Context, topic string, message *pubsub.Message) error
	}
	client struct {
		resource *resource.Resource
		client   *pubsub.Client
	}
)

func (c *client) Publish(ctx context.Context, topic string, message *pubsub.Message) error {
	tr := tracer.StartTrace(ctx, "messageQueue.pubSub.Publish")
	ctx = tr.Context()
	defer tr.Finish()

	topicData := c.client.Topic(topic)
	result := topicData.Publish(ctx, message)

	if _, err := result.Get(ctx); err != nil {
		return err
	}

	return nil
}
func New(resource *resource.Resource) (Client, error) {
	creadentialJSON, err := base64.RawStdEncoding.DecodeString(resource.Credential.PubSub.CredentialBase64)
	if err != nil {
		return nil, err
	}

	clientPubSub, err := pubsub.NewClient(context.Background(),
		resource.Credential.PubSub.ProjectID, option.WithCredentialsJSON(creadentialJSON))

	if err != nil {
		return nil, err
	}
	return &client{
		resource: resource,
		client:   clientPubSub,
	}, nil
}
