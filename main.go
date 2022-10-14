package main

import (
	"flag"
	"fmt"
)

var Version = "1.0.0"

func main() {
	port := flag.Int("p", 8080, "port to listen on")
	serverTomlFile := flag.String("c", "", "server config file")
	dumpVersion := flag.Bool("v", false, "dump version")

	flag.Parse()

	if *dumpVersion {
		println(Version)
		return
	}

	if *port == 0 {
		panic("Port must be specified")
	}

	if *serverTomlFile == "" {
		panic("Server config file must be specified")
	}

	s := NewServer(*port, *serverTomlFile)

	go listenSignals(s)

	fmt.Println("Version : ", Version)

	s.Start()
}
