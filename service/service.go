package service

import (
	"context"
	"log"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/server"
)

type service struct {
	context context.Context
	cancel  context.CancelFunc
	config  *foundation.Config
}

func RunApp(config *foundation.Config) int {
	ctx, cancel := context.WithCancel(context.Background())
	svc := &service{
		context: ctx,
		cancel:  cancel,
		config:  config,
	}
	_ = svc // todo, handle shutdown signals etc.

	context := &foundation.Context{
		Context: ctx,
		Config:  config,
	}

	err := server.StartServer(context)
	if err != nil {
		log.Println("StartServer error:", err)
		return 1
	}

	<-ctx.Done()
	log.Println("shutting down...")
	svc.cancel()

	return 0
}
