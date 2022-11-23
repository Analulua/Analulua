package grpc

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

// Serve listen for client request
func Serve(address string, server *grpc.Server) {

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Print(err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		<-c
		log.Println("Shutting down server gracefully...")
		server.GracefulStop()
	}()

	err = server.Serve(lis)
	if err != nil {
		log.Print(err)
		return
	}
}
