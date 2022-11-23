package grpc

import (
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	err "newdemo1/resource/jaeger/common/error"
)

func TestErrorMetadata(t *testing.T) {
	attrs := make(map[string]string)
	attrs["max"] = "5"
	svcErr := err.ServiceError{
		Code:       "TEST_ERROR",
		Message:    "Test error",
		Attributes: attrs,
	}
	if got := errorMetadata(svcErr); got == nil {
		t.Fatalf("bad metadata: %v ", got)
	}
}

func TestGRPCError(t *testing.T) {
	attrs := make(map[string]string)
	attrs["max"] = "5"
	svcErr := err.ServiceError{
		Code:       "TEST_ERROR",
		Message:    "Test error",
		Attributes: attrs,
	}
	if got := grpcError(codes.Internal, svcErr); got == nil {
		t.Fatalf("bad error: %v ", got)
	}
}

func TestUnaryErrorInterceptor(t *testing.T) {
	server := grpc.NewServer(grpc.UnaryInterceptor(UnaryErrorInterceptor(nil)))
	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	lis, _ := net.Listen("tcp", "localhost:0")
	_ = server.Serve(lis)
}

func TestStreamErrorInterceptor(t *testing.T) {
	server := grpc.NewServer(grpc.StreamInterceptor(StreamErrorInterceptor(nil)))
	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	lis, _ := net.Listen("tcp", "localhost:0")
	_ = server.Serve(lis)
}

func TestUnaryAuthInterceptor(t *testing.T) {
	server := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor()))
	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	lis, _ := net.Listen("tcp", "localhost:0")
	_ = server.Serve(lis)
}

func TestRecoveryInterceptor(t *testing.T) {
	unary, stream := recoveryInterceptor()
	server := grpc.NewServer(grpc.UnaryInterceptor(unary), grpc.StreamInterceptor(stream))
	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	lis, _ := net.Listen("tcp", "localhost:0")
	_ = server.Serve(lis)
}

func TestWithRecovery(t *testing.T) {
	server := grpc.NewServer(WithRecovery()...)
	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	lis, _ := net.Listen("tcp", "localhost:0")
	_ = server.Serve(lis)
}

func TestWithErrorDetails(t *testing.T) {
	server := grpc.NewServer(WithErrorDetails(nil)...)
	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	lis, _ := net.Listen("tcp", "localhost:0")
	_ = server.Serve(lis)
}

func TestWithValidation(t *testing.T) {
	server := grpc.NewServer(WithValidation()...)
	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	lis, _ := net.Listen("tcp", "localhost:0")
	_ = server.Serve(lis)
}

func TestWithDefault(t *testing.T) {
	server := grpc.NewServer(WithDefault(nil)...)
	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	lis, _ := net.Listen("tcp", "localhost:0")
	_ = server.Serve(lis)
}

type ks struct{}

func (ks ks) GetSecretKey(_ string) []byte {
	return []byte("test-key")
}

var signInfo = SignatureInfo{
	Alg:             "HS256",
	HeaderFieldKeys: "partner_id device_id",
	BodyFieldKeys:   "trx_id trx_time",
	DefaultFields:   []string{"partner_id", "device_id", "timestamp"},
}

func TestUnarySignatureInterceptor(t *testing.T) {
	server := grpc.NewServer(grpc.UnaryInterceptor(UnarySignatureInterceptor(ks{}, signInfo)))
	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	lis, _ := net.Listen("tcp", "localhost:0")
	_ = server.Serve(lis)
}

func TestWithSignature(t *testing.T) {
	server := grpc.NewServer(WithSignature(ks{}, signInfo)...)
	go func() {
		time.Sleep(2 * time.Second)
		server.Stop()
	}()
	lis, _ := net.Listen("tcp", "localhost:0")
	_ = server.Serve(lis)
}
