# Registro #
Registro is a simple service registry written in Go.

It provides a client to be used inside any Go application and a REST
server that handles registration and service state management.

## Server ##
The server is responsible for keeping a list of running services and
managing the state of each one. It receives requests from the services
and keep them in the memory. There are no permanent storage. If the server
shutdown, all data is lost.

Services that are unresponsive for more that 10 minutes are deleted from the
list.

## Server Usage ##
There are two ways to run the server. The first one is by compiling and
running on the local machine and the second one is a lightweight docker
image.

### Local Machine ###
The following script will install dependencies, compile and run the application.

Please note that the *--addr :8000* sets the listening address for the server
socket and it's not required. Default is *:8080*.

	$ go get -d github.com/gorilla/mux
	$ cd $GOPATH/src/github.com/mscansian/registro
	$ go build -o registro .
	$ ./registro --addr :8080

### Docker ###
The following script will build a docker image and run with default options. It
should create two images *mscansian/registro:build* for the build environment
and a small *mscansian/registro:latest* image for the application itself.

Please note that the *--addr :8000* sets the listening address for the server
socket and it's not required. Default is *:8080*.

	$ ./build.sh
	$ docker run --rm -p 8000:8000 mscansian/registro --addr :8000

## Client ##
The client is a library that simplify the handling of request to the service
registry REST server.

Documentation available in:

	$ godoc github.com/mscansian/registro/client

## Client Usage ##
Clients need to create an application (if not exists), register itself as an
instance and send a heartbeat every 30 seconds.

	import (
		"log"
		"time"
		"github.com/mscansian/registro/client"
	)
	
	c := client.NewClient("http://localhost:8000/registro")
	app, inst, err := c.RegisterService("service-id", "app-name", "127.0.0.1", 8080)
	if err != nil {
		log.Fatal(err)
	}
	
	go func() {
		for {
			err := c.RenewInstance(app, inst)
			if err != nil {
				log.Printf("service renew error: %s", err)
			}
			<-time.After(30 * time.Second)
		}
	}()
	<-time.After(90 * time.Second)