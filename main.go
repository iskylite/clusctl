package main

import (
	"context"
	log "myclush/logger"
	"os"
	"os/signal"
)

var (
	ctx    context.Context
	cancel context.CancelFunc
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	// signal handler
	ctx, cancel = context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(c chan os.Signal, cancel func()) {
		<-c
		log.Info("Get Cancel Signal")
		cancel()
	}(c, cancel)
	// Run Setup Service
	err := run(ctx, cancel)
	if err != nil {
		panic(err)
	}
}
