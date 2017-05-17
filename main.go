/*
	Registro is a simple service registry written in Go.

	It provides a client to be used inside any Go application and a REST
	server that handles registration and service state management.
*/
package main

import (
	"flag"
	"log"

	"github.com/mscansian/registro/server"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	s := server.NewServer(*addr)
	log.Fatal(s.Serve())
}
