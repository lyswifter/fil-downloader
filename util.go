package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/xerrors"
)

var AlreadyErr = xerrors.New("Already Downloaded")

func MkDir(path string) error {
	if _, err := os.ReadDir(path); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, 0777)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}

	return true
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}

	return s.IsDir()
}

func IsFile(path string) bool {
	return !IsDir(path)
}

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

func download(urlstr string, filename string, maddr string, snum string) error {
	log.Infof("Start download: %s", urlstr)

	ctx := context.Background()

	parsedPath, err := url.Parse(filename)
	if err != nil {
		return err
	}

	fn := path.Base(parsedPath.Path)

	keyName := ""
	if fn == "sealed" {
		keyName = fmt.Sprintf("/sealed/s-t0%s-%s", maddr, snum)
	} else {
		keyName = fmt.Sprintf("/cache/s-t0%s-%s/%s", maddr, snum, fn)
	}

	state, _, err := QueryStatus(ctx, keyName)
	if err != nil {
		return err
	}

	if state == "already downloaded" {
		log.Warnf("File %s is already downloaded", keyName)
		return AlreadyErr
	}

	r, err := http.Get(urlstr)
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

	go func(ctx context.Context, keyName string) {
		ticker := time.NewTicker(30 * time.Second)

		for {
			select {
			case <-ticker.C:
				precent := float64(reader.Current*10000/reader.Total) / 100
				log.Infof("Download %s %.2f%%", urlstr, precent)

				if reader.Current == reader.Total {
					log.Infof("Finished download %s total: %d cur: %d", urlstr, reader.Total, reader.Current)
					err = MarkAs(ctx, keyName, "already downloaded")
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
	}(ctx, keyName)

	_, _ = io.Copy(f, reader)

	return nil
}

func RemoveContents(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}

		log.Infof("remove file: %s ok", file)
	}
	return nil
}
