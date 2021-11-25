package main

import (
	"os"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli"
)

var log = logging.Logger("fil-downloader")

var RepoDir = "~/.fil-downloader"

func main() {
	logging.SetLogLevel("*", "INFO")

	local := []cli.Command{
		initCmd,
		daemonCmd,
		downloadmd,
		// uploadCmd,
	}

	app := &cli.App{
		Name:    "fil-downloader",
		Usage:   "Used for download files from qiniu cruster",
		Version: "0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "sector-dir",
				Value: "~/.fil-downloader",
			},
		},

		Commands: local,
	}

	if err := app.Run(os.Args); err != nil {
		log.Warn(err)
		os.Exit(1)
	}
}
