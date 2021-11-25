package main

import (
	"encoding/json"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/qiniupd/qiniu-go-sdk/syncdata/operation"
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

		repodir, err := homedir.Expand(RepoDir)
		if err != nil {
			return err
		}

		scfgFile := path.Join(repodir, "s.json")
		ucfgFile := path.Join(repodir, "u.json")

		log.Infof("s.json: %s u.json: %s", scfgFile, ucfgFile)

		isexist := Exists(scfgFile)
		if !isexist {
			f, err := os.Create(scfgFile)
			if err != nil {
				return err
			}

			binfo := BucketInfo{}
			info, err := json.Marshal(binfo)
			if err != nil {
				return err
			}

			_, err = f.Write(info)
			if err != nil {
				log.Infof("here: %s", scfgFile)
				return err
			}
		}

		isexist = Exists(ucfgFile)
		if !isexist {
			cfg := operation.Config{}

			f, err := os.Create(ucfgFile)
			if err != nil {
				return err
			}

			info, err := json.Marshal(cfg)
			if err != nil {
				return err
			}

			_, err = f.Write(info)
			if err != nil {
				return err
			}
		}

		return nil
	},
}
