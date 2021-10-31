SHELL=/usr/bin/env bash

.PHONY: clean
clean:
	rm fil-downloader

.PHONY: all
all:
	go build -o fil-downloader *.go
