# 🦑 Inkpot Server

This repository contains the code for the inkpot server. Sorry for the bad documentation, better documentation will (hopefully) follow.

The code is written in go, most of what's going on is in `main.go`. `go run main.go` will spin up an http server.

## Deployment

`go build` creates an executable called `inkpot-server`. If you run this somewhere, make sure to place the `assets` and `templates` folder there as well.
