package main

import (
	"context"
	"os"
	"path"

	"github.com/ipfs/go-datastore"
	levelds "github.com/ipfs/go-ds-leveldb"
	"github.com/mitchellh/go-homedir"
	ldbopts "github.com/syndtr/goleveldb/leveldb/opt"
)

func setupLevelDs(path string, readonly bool) (datastore.Batching, error) {
	if _, err := os.ReadDir(path); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, 0777)
			if err != nil {
				return nil, err
			}
		}
	}

	db, err := levelds.NewDatastore(path, &levelds.Options{
		Compression: ldbopts.NoCompression,
		NoSync:      false,
		Strict:      ldbopts.StrictAll,
		ReadOnly:    readonly,
	})
	if err != nil {
		log.Errorf("NewDatastore: %s", err)
		return nil, err
	}
	return db, err
}

func mkSectorsDir(path string) error {
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

func DataStores() error {
	repodir, err := homedir.Expand(RepoDir)
	if err != nil {
		return err
	}

	ldb, err := setupLevelDs(path.Join(repodir, "loadinfo"), false)
	if err != nil {
		log.Errorf("setup beacondb: err %s", err)
		return err
	}

	InfoDB = ldb
	log.Infof("InfoDB: %+v", InfoDB)

	return nil
}

func MarkAsDownloaded(ctx context.Context, file string) error {
	key := datastore.NewKey(file)
	ishas, err := InfoDB.Has(ctx, key)
	if err != nil {
		return err
	}

	if !ishas {
		err := InfoDB.Put(ctx, key, []byte("already"))
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func QueryStatus(ctx context.Context, file string) (bool, error) {
	key := datastore.NewKey(file)
	ishas, err := InfoDB.Has(ctx, key)
	if err != nil {
		return false, err
	}

	if ishas {
		ret, err := InfoDB.Get(ctx, key)
		if err != nil {
			return false, err
		}

		if string(ret) == "already" {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, nil
}
