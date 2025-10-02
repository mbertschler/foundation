package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/db"
	"github.com/mbertschler/foundation/server"
	"github.com/mbertschler/foundation/server/broadcast"
)

func RunApp(config *foundation.Config) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	context := &foundation.Context{
		Context:   ctx,
		Config:    config,
		Broadcast: broadcast.New(),
	}

	var err error
	if context.Config.LitestreamYml != "" {
		err = restoreLitestreamIfNeeded(context)
		if err != nil {
			log.Println("restoreLitestreamIfNeeded error:", err)
			return 1
		}

		err = startLitestream(context)
		if err != nil {
			log.Println("startLitestream error:", err)
			return 1
		}
	}

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

	// ... rest of your app initialization

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("shutting down...")
	context.DB.Close()
	return 0
}

func restoreLitestreamIfNeeded(ctx *foundation.Context) error {
	if _, err := os.Stat(ctx.Config.DBPath); os.IsNotExist(err) {
		// Database doesn't exist, restore from replica
		cmd := exec.Command("litestream", "restore", "-config",
			ctx.Config.LitestreamYml, ctx.Config.DBPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restore database: %w", err)
		}
	}
	return nil
}

func startLitestream(ctx *foundation.Context) error {
	cmd := exec.CommandContext(ctx.Context, "litestream", "replicate", "-config", ctx.Config.LitestreamYml)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Println("litestream exited with error:", err)
		}
	}()

	return nil
}
