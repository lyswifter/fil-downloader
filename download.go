package main

import (
	"context"
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
	"github.com/qiniupd/qiniu-go-sdk/syncdata/operation"
	"github.com/urfave/cli"
	"golang.org/x/xerrors"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano() | int64(os.Getpid())))
var MAXCHECKING = 8
var sem chan struct{}

var uploader operation.Uploader

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
			Name:  "download-cfg-path",
			Usage: "Giving cruster download config information path: s.json",
			Value: path.Join(RepoDir, "s.json"),
		},
		&cli.StringFlag{
			Name:  "upload-cfg-path",
			Usage: "Giving cruster upload config information path: u.json",
			Value: path.Join(RepoDir, "u.json"),
		},
		&cli.StringFlag{
			Name:  "sector-path",
			Usage: "Giving sectors to download information path",
			Value: path.Join(RepoDir, "sectors.txt"),
		},
		&cli.StringFlag{
			Name:  "target-path",
			Usage: "Giving the path to store sectors temp",
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

		err := DataStores()
		if err != nil {
			return err
		}

		log.Infof("db: %+v", InfoDB)

		maxqueue := cctx.Int64("max-queue")
		if maxqueue <= 0 {
			return xerrors.Errorf("max queue must greater than zero")
		}

		sem = make(chan struct{}, maxqueue)

		downloadCfgpath := cctx.String("download-cfg-path")
		if downloadCfgpath == "" {
			return xerrors.Errorf("ruster config file must provide")
		}

		upCfgpath := cctx.String("upload-cfg-path")
		if upCfgpath == "" {
			return xerrors.Errorf("ruster config file must provide")
		}

		sectorpath := cctx.String("sector-path")
		if sectorpath == "" {
			return xerrors.Errorf("sector infos config file must provide")
		}

		targetpath := cctx.String("target-path")
		if targetpath == "" {
			return xerrors.Errorf("sector temp location path must provide")
		}

		uid := cctx.String("uid")
		if uid == "" {
			return xerrors.Errorf("storage user identity must provide")
		}

		minerAddr := cctx.String("miner")
		if minerAddr == "" {
			return xerrors.Errorf("miner address must provide")
		}

		dcfgFilepath, _ := homedir.Expand(downloadCfgpath)
		sFilePath, _ := homedir.Expand(sectorpath)

		file, err := os.Open(dcfgFilepath)
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

		upcfgFilepath, _ := homedir.Expand(upCfgpath)
		x, err := operation.Load(upcfgFilepath)
		if err != nil {
			return err
		}

		uploader = *operation.NewUploader(x)

		sectornumbers := readline(sFilePath)
		if len(sectornumbers) == 0 {
			return xerrors.New("sector numbers must not be empty")
		}

		log.Infof("Need to download sectors: %d %v", len(sectornumbers), sectornumbers)

		ssize := cctx.String("sector-size")

		var wg sync.WaitGroup

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

				sectorDir := path.Join(targetpath, fmt.Sprintf("f0%s", minerAddr), "sectors", snum)
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

				//
				log.Infof("------  WILL TIGGER UPLOAD FOR: %s ------", snum)

				fs, err := ioutil.ReadDir(sectorDir)
				if err != nil {
					return err
				}

				if len(fs) == 0 {
					log.Warnf("nothing need to upload for: %s", snum)
					return nil
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

					if state == "already uploaded" || state == "already removed" {
						log.Warnf("File %s is %s", keyName, state)
						continue
					}

					log.Infof("upload start: %s", filename)

					start := time.Now()

					err = uploader.Upload(filename, keyName)
					if err != nil {
						return err
					}

					err = MarkAs(context.TODO(), keyName, "already uploaded")
					if err != nil {
						return err
					}

					log.Infof("upload finished: %s took: %s", filename, time.Since(start).String())

					log.Infof("remove start: %s", filename)

					rstart := time.Now()

					err = os.Remove(filename)
					if err != nil {
						return err
					}

					err = MarkAs(context.TODO(), keyName, "already removed")
					if err != nil {
						return err
					}

					log.Infof("remove finished: %s took: %s", filename, time.Since(rstart).String())
				}

				return nil
			}(snum)
		}

		wg.Wait()

		return nil
	},
}
