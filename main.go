package main

import (
	"go-postgres/cli"
	"go-postgres/database"
)

func main() {
	//test database working
    database.Script()

	// test CLI working
	cli.CLI()
}