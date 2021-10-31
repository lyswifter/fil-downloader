package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli"
)

var daemonCmd = cli.Command{
	Name:        "daemon",
	Description: "Start download daemon",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "file",
		},
	},
	Action: func(cctx *cli.Context) error {

		ticker := time.NewTicker(30 * time.Second)
		ctx := context.Background()
		sigChan := make(chan os.Signal, 2)

		log.Info("File downloader is running...")

		for {
			select {
			case <-ticker.C:
				log.Info("I am running")
			case <-ctx.Done():
				log.Warn("Shutting down..")
				log.Warn("Graceful shutdown successful")
				signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
				return nil
			}
		}

	},
}
