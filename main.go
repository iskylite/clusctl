package main

import (
	"context"
	"flag"
	log "myclush/logger"
	"os"
	"os/signal"
)

func main() {
	if debug {
		log.SetLevel(log.DEBUG)
		log.Debug("Logger Setup In DEBUG Mode")
	}
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	// signal handler
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(c chan os.Signal, cancel func()) {
		<-c
		log.Info("Get Cancel Signal")
		cancel()
	}(c, cancel)
	// Setup Service
	switch cc | ee | ss | pp | oo {
	case 1:
		putStreamClientServiceSetup(ctx, cancel)
	case 2:
		putStreamServerServiceSetup(ctx, cancel)
	case 4:
		RunCmdClientServiceSetup(ctx, cancel)
	case 8:
		PingClientServiceSetup(ctx)
	case 16:
		OoBwServiceSetup(ctx)
	default:
		flag.PrintDefaults()
		os.Exit(22)
	}
}
