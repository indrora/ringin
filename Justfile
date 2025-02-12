
default:
    @just --list

build:
    go build -o bin/ringin ./cmd

bootstrap:
    go mod download
