package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	"golang.org/x/xerrors"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano() | int64(os.Getpid())))
var MAXCHECKING = 8
var sem chan struct{}

var downloadmd = cli.Command{
	Name:        "download",
	Description: "Download from cruster manually",
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
			Usage: "Giving cruster config information path: s.json",
			Value: path.Join(RepoDir, "s.json"),
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

		maxqueue := cctx.Int64("max-queue")
		if maxqueue <= 0 {
			return xerrors.Errorf("max queue must greater than zero")
		}

		sem = make(chan struct{}, maxqueue)

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
		sFilePath, _ := homedir.Expand(sectorpath)

		file, err := os.Open(cfgFilepath)
		if err != nil {
			return err
		}

		defer file.Close()

		val, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}

		var bucketinfo BucketInfo
		err = json.Unmarshal(val, &bucketinfo)
		if err != nil {
			return err
		}

		log.Infof("Bucket info: %+v", bucketinfo)

		sectornumbers := readline(sFilePath)
		if len(sectornumbers) == 0 {
			return xerrors.New("sector numbers must not be empty")
		}

		log.Infof("Need to download sectors: %d %v", len(sectornumbers), sectornumbers)

		ssize := cctx.String("sector-size")

		var wg sync.WaitGroup

		// var sectorinfos []SectorInfo
		for _, snum := range sectornumbers {
			// if already download, continue

			sem <- struct{}{}

			wg.Add(1)

			go func(snum string) error {

				defer wg.Done()
				defer func() {
					<-sem
				}()

				task := assembleDownloadTask(minerAddr, uid, bucketinfo, snum, ssize)

				//pick target host
				rsHost := bucketinfo.Rs_hosts[random.Intn(len(bucketinfo.Rs_hosts))]
				downloadHost := strings.Replace(rsHost, "9433", "5000", 1)

				pauxUrl := fmt.Sprintf("%s/%s", downloadHost, task.Paux)
				sealedUrl := fmt.Sprintf("%s/%s", downloadHost, task.Sealed)

				repo, err := homedir.Expand(RepoDir)
				if err != nil {
					return err
				}

				sectorDir := path.Join(repo, "sectors", snum)
				err = mkSectorsDir(sectorDir)
				if err != nil {
					return err
				}

				err = download(pauxUrl, path.Join(sectorDir, "p_aux"), minerAddr, snum)
				if err != nil {
					if err != AlreadyErr {
						return err
					}
				}

				for _, cachefile := range task.Cache {
					splitArr := strings.Split(cachefile, "/")
					length := len(strings.Split(cachefile, "/"))
					err = download(fmt.Sprintf("%s/%s", downloadHost, cachefile), path.Join(sectorDir, splitArr[length-1]), minerAddr, snum)
					if err != nil {
						if err != AlreadyErr {
							return err
						}
					}
				}

				err = download(sealedUrl, path.Join(sectorDir, "sealed"), minerAddr, snum)
				if err != nil {
					if err != AlreadyErr {
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
