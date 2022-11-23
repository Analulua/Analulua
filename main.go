package main

import (
	"log"
	"newdemo1/application"
	"newdemo1/infrastructure"
	"newdemo1/resource"
	"newdemo1/transport"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	zone := time.FixedZone("CST", 7*3600)
	time.Local = zone

	resource, err := resource.NewResource("config.yaml", "credential.yaml")
	if err != nil {
		panic(err)
	}
	defer resource.Flush()

	infra, err := infrastructure.NewInfrastructure(resource)
	if err != nil {
		panic(err)
	}

	application, err := application.NewApplication(resource, infra)
	if err != nil {
		panic(err)
	}

	tp, err := transport.NewTransport(resource, application)
	if err != nil {
		panic(err)
	}
	tp.Run()

	graceful := make(chan os.Signal, 1)
	signal.Notify(graceful, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-graceful
		tp.Stop()
	}()

	// DONE
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	log.Println("All server stopped!")
}
