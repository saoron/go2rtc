#!/bin/bash


env GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-w -s"

upx -9 go2rtc