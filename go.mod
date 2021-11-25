module github.com/lyswifter/fil-downloader

go 1.17

require (
	github.com/ipfs/go-datastore v0.5.0
	github.com/ipfs/go-ds-leveldb v0.5.0
	github.com/ipfs/go-log/v2 v2.3.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/qiniupd/qiniu-go-sdk v1.1.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/urfave/cli v1.22.5
	golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0-20190314233015-f79a8a8ca69d // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/jbenet/goprocess v0.1.4 // indirect
	github.com/kirsle/configdir v0.0.0-20170128060238-e45d2f54772f // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20200116001909-b77594299b42 // indirect
)

replace github.com/qiniupd/qiniu-go-sdk => github.com/lyswifter/qiniu-go-sdk v1.1.2
