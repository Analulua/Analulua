package grpc

import (
	"google.golang.org/grpc"
	"log"
	"newdemo1/constant"
	"newdemo1/resource"
	commonGrpc "newdemo1/resource/jaeger/common/grpc"
	commonTelemetryGrpc "newdemo1/resource/jaeger/common/telemetry/instrumentation/grpc"
)

type Grpc struct {
	resource *resource.Resource
	server   *grpc.Server
}

func NewServer(resource *resource.Resource) (*grpc.Server, error) {
	commonIntercept := commonGrpc.WithDefault(constant.ServiceErrorCodeToGRPCErrorCode)
	chainedUnaryInterceptor := grpc.ChainUnaryInterceptor(
		commonTelemetryGrpc.UnaryServerInterceptor(resource.Jaeger.Tracer, constant.ServiceErrorCodeToGRPCErrorCode),
	)

	interceptors := append(commonIntercept, chainedUnaryInterceptor)
	return grpc.NewServer(interceptors...), nil
}

func NewGrpc(resource *resource.Resource) (Grpc, error) {
	server, err := NewServer(resource)
	if err != nil {
		return Grpc{}, err
	}
	return Grpc{
		resource: resource,
		server:   server,
	}, nil
}

func (g *Grpc) Serve() {
	log.Println("[Transaction Service GRPC] server started. Listening on port ", g.resource.Config.Service.GrpcPort)
	commonGrpc.Serve(g.resource.Config.Service.GrpcPort, g.server)
}

func (g *Grpc) GracefulStop() error {
	g.server.GracefulStop()

	return nil
}
