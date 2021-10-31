package main

import (
	"github.com/urfave/cli"
)

var initCmd = cli.Command{
	Name:        "init",
	Description: "Initial filcoin file download repo",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "path",
		},
	},
	Action: func(cctx *cli.Context) error {

		// RepoDir

		err := DataStores()
		if err != nil {
			return err
		}

		return nil
	},
}
