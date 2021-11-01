package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/xerrors"
)

var AlreadyErr = xerrors.New("Already Downloaded")

func readline(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	rd := bufio.NewReader(f)

	var ret = []string{}
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行

		if err != nil || io.EOF == err {
			break
		}

		line = strings.Replace(line, "\n", "", -1)

		ret = append(ret, line)
	}

	return ret
}

// assembleDownloadUrl assembleDownloadUrl
func assembleDownloadTask(miner string, uid string, bucket BucketInfo, sectornumber string, ssize string) SectorInfo {

	//p_aux sealed
	paux := fmt.Sprintf("getfile/%s/f0%s//cache/s-t0%s-%s/p_aux", uid, miner, miner, sectornumber)
	sealed := fmt.Sprintf("getfile/%s/f0%s//sealed/s-t0%s-%s", uid, miner, miner, sectornumber)

	var cachepaths []string = []string{}
	if ssize == "32GiB" {
		for i := 0; i < 8; i++ {
			cache := fmt.Sprintf("getfile/%s/f0%s//cache/s-t0%s-%s/sc-02-data-tree-r-last-%d.dat", uid, miner, miner, sectornumber, i)
			cachepaths = append(cachepaths, cache)
		}
	} else {
		for i := 0; i < 16; i++ {
			cache := fmt.Sprintf("getfile/%s/f0%s//cache/s-t0%s-%s/sc-02-data-tree-r-last-%d.dat", uid, miner, miner, sectornumber, i)
			cachepaths = append(cachepaths, cache)
		}
	}

	return SectorInfo{
		SectorNumber: sectornumber,
		SectorSize:   ssize,
		Paux:         paux,
		Cache:        cachepaths,
		Sealed:       sealed,
	}
}

type Reader struct {
	io.Reader
	Total   int64
	Current int64
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.Current += int64(n)
	return
}

func download(url string, filename string) error {
	log.Infof("Start download: %s", url)

	ctx := context.Background()

	isalready, err := QueryStatus(ctx, url)
	if err != nil {
		return err
	}

	if isalready {
		log.Warnf("File %s is already downloaded", url)
		return AlreadyErr
	}

	r, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer func() { _ = r.Body.Close() }()

	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	reader := &Reader{
		Reader: r.Body,
		Total:  r.ContentLength,
	}

	go func(ctx context.Context) {
		ticker := time.NewTicker(30 * time.Second)

		for {
			select {
			case <-ticker.C:
				precent := float64(reader.Current*10000/reader.Total) / 100
				log.Infof("Download %s %.2f%%", url, precent)

				if reader.Current == reader.Total {
					log.Infof("Finished download %s total: %d total: %d cur: %d", url, reader.Total, reader.Current)
					err = MarkAsDownloaded(ctx, url)
					if err != nil {
						log.Errorf("mark as download err %s", err.Error())
						return
					}
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	_, _ = io.Copy(f, reader)

	return nil
}
