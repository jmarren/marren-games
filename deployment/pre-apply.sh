#!/bin/bash
# build go app for linux/arm64 with musl to compile sqlite driver
echo "****** building binary ******"
env GOARCH=arm64 GOOS=linux CC=aarch64-linux-musl-gcc CGO_ENABLED=1 go build -x -tags "sqlite3" -o ./build ../cmd/server/main.go

echo "****** syncing ******** "
# copy all necessary files to s3 bucket
aws s3 sync ../ui s3://ask-away-s3-bucket/app/ui
aws s3 sync ../sql s3://ask-away-s3-bucket/app/sql
aws s3 cp ./build s3://ask-away-s3-bucket/app/build
aws s3 cp ./nginx.conf s3://ask-away-s3-bucket/app/nginx.conf
aws s3 sync ./letsencrypt/ s3://ask-away-s3-bucket/app/letsencrypt/
