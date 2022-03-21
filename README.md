# 🦑 Inkpot Server

This repository contains the code for the inkpot server. Sorry for the bad documentation, better documentation will (hopefully) follow.

The code is written in go, most of what's going on is in `main.go`. `go run main.go` will spin up an http server.

## Deployment

`go build` creates an executable called `inkpot-server`. If you run this somewhere, make sure to place the `assets`, `migrations` and `templates` folder there along with their content. Otherwise the server will fail to run.

## Database Migrations

This repository uses an SQLite database and [`golang-migrate`](https://github.com/golang-migrate/migrate/blob/331a15d92a86e002ce5043e788cc2ca287ab0cc2/MIGRATIONS.md) for migrations. All database migrations can be found in the `migrations` folder at the root of this repository.
