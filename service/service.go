package service

import (
	"context"
	"log"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/db"
	"github.com/mbertschler/foundation/server"
	"github.com/mbertschler/foundation/server/broadcast"
)

func RunApp(config *foundation.Config) int {
	ctx, cancel := context.WithCancel(context.Background())

	context := &foundation.Context{
		Context:   ctx,
		Config:    config,
		Broadcast: broadcast.New(),
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
	cancel()
	return 0
}
