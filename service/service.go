package service

import (
	"context"
	"log"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/db"
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

	var err error
	context.DB, err = db.StartDB(context)
	if err != nil {
		log.Println("StartDB error:", err)
		return 1
	}

	err = server.RunServer(context)
	if err != nil {
		log.Println("RunServer error:", err)
		return 1
	}

	<-ctx.Done()
	log.Println("shutting down...")
	svc.cancel()

	return 0
}
