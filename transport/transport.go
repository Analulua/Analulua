package transport

import (
	"newdemo1/application"
	"newdemo1/infrastructure/mq"
	"newdemo1/resource"
	"newdemo1/transport/grpc"
	"newdemo1/transport/http"
)

type Transport struct {
	Grpc grpc.Grpc
	Http http.Http
	MQ   mq.PubSub
	//Task task.Task
}

func NewTransport(resource *resource.Resource, app *application.Application) (Transport, error) {
	grpcTransport, err := grpc.NewGrpc(resource)
	if err != nil {
		return Transport{}, err
	}

	httpTransport := http.NewHttp(resource, app)
	if err != nil {
		return Transport{}, err
	}
	m, err := mq.NewMQ(resource)
	if err != nil {
		return Transport{}, err
	}
	return Transport{
		Grpc: grpcTransport,
		Http: httpTransport,
		MQ:   m,
	}, nil
}

func (t *Transport) Run() {
	go func() {
		_ = t.Http.Serve()
	}()

	go t.MQ.PubSub()

	go t.Grpc.Serve()

}

func (t *Transport) Stop() {
	_ = t.Grpc.GracefulStop()
}
