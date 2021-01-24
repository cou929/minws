package main

import (
	"log"

	"github.com/cou929/minws/server"
)

func main() {
	srv := server.NewServer(":5001")
	log.Fatal(srv.Serve())
}
