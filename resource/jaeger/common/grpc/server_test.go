package grpc

import (
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestServe(t *testing.T) {
	server := grpc.NewServer()
	go func() {
		Serve("localhost:0", server)
	}()
	time.Sleep(2 * time.Second)
	server.Stop()
}
