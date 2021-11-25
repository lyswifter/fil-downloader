package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	"golang.org/x/xerrors"

	"github.com/qiniupd/qiniu-go-sdk/syncdata/operation"
)

var randomn = rand.New(rand.NewSource(time.Now().UnixNano() | int64(os.Getpid())))
var MAXQUEST = 8
var semu chan struct{}

var uploader operation.Uploader

var uploadCmd = cli.Command{
	Name:        "upload",
	Description: "Upload to cruster manually",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "uid",
			Usage: "Specify user identity",
		},
		&cli.StringFlag{
			Name:  "miner",
			Usage: "Specify miner address",
		},
		&cli.StringFlag{
			Name:  "config-path",
			Usage: "Giving cruster config information path",
			Value: path.Join(RepoDir, "u.json"),
		},
		&cli.StringFlag{
			Name:  "sector-path",
			Usage: "Giving sectors to download information path",
			Value: path.Join(RepoDir, "sectors.txt"),
		},
		&cli.Int64Flag{
			Name:  "max-queue",
			Usage: "The max queue number",
			Value: 1,
		},
		&cli.StringFlag{
			Name:  "sector-size",
			Usage: "sector size info, like: 32GiB, 64GiB",
			Value: "32GiB",
		},
	},
	Action: func(cctx *cli.Context) error {

		log.Infof("db: %+v", InfoDB)

		maxqueue := cctx.Int64("max-queue")
		if maxqueue <= 0 {
			return xerrors.Errorf("max queue must greater than zero")
		}

		semu = make(chan struct{}, maxqueue)

		cfgpath := cctx.String("config-path")
		if cfgpath == "" {
			return xerrors.Errorf("ruster config file must provide")
		}

		sectorpath := cctx.String("sector-path")
		if sectorpath == "" {
			return xerrors.Errorf("sector infos config file must provide")
		}

		uid := cctx.String("uid")
		if uid == "" {
			return xerrors.Errorf("storage user identity must provide")
		}

		minerAddr := cctx.String("miner")
		if minerAddr == "" {
			return xerrors.Errorf("miner address must provide")
		}

		cfgFilepath, _ := homedir.Expand(cfgpath)
		x, err := operation.Load(cfgFilepath)
		if err != nil {
			return err
		}

		uploader = *operation.NewUploader(x)

		sFilePath, _ := homedir.Expand(sectorpath)
		sectornumbers := readline(sFilePath)
		if len(sectornumbers) == 0 {
			return xerrors.New("sector numbers must not be empty")
		}

		log.Infof("Need to download sectors: %d %v", len(sectornumbers), sectornumbers)

		var wg sync.WaitGroup

		repo, err := homedir.Expand(RepoDir)
		if err != nil {
			return err
		}

		for _, snum := range sectornumbers {
			// if already uploaded, continue

			semu <- struct{}{}

			wg.Add(1)

			go func(snum string) error {

				defer wg.Done()
				defer func() {
					<-semu
				}()

				sectorDir := path.Join(repo, "sectors", snum)
				err = mkSectorsDir(sectorDir)
				if err != nil {
					return err
				}

				fs, err := ioutil.ReadDir(sectorDir)
				if err != nil {
					return err
				}

				for _, f := range fs {
					fn := f.Name()

					keyName := ""
					if fn == "sealed" {
						keyName = fmt.Sprintf("/sealed/s-t0%s-%s", minerAddr, snum)
					} else {
						keyName = fmt.Sprintf("/cache/s-t0%s-%s/%s", minerAddr, snum, fn)
					}

					if keyName == "" {
						return xerrors.New("key name must not be empty")
					}

					filename := path.Join(sectorDir, fn)

					state, _, err := QueryStatus(context.TODO(), keyName)
					if err != nil {
						return err
					}

					if state == "already uploaded" {
						log.Warnf("File %s is already uploaded", keyName)
						continue
					}

					log.Infof("upload start: %s", filename)
					start := time.Now()

					err = uploader.Upload(filename, keyName)
					if err != nil {
						return err
					}

					log.Infof("upload finished: %s took: %s", filename, time.Since(start).String())

					err = RemoveContents(sectorDir)
					if err != nil {
						return err
					}

					err = MarkAs(context.TODO(), keyName, "already uploaded")
					if err != nil {
						return err
					}
				}

				return nil
			}(snum)
		}

		wg.Wait()

		return nil
	},
}
