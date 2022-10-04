package main

import "flag"

func main() {
	port := flag.Int("p", 8080, "port to listen on")
	serverTomlFile := flag.String("c", "", "server config file")

	flag.Parse()

	if *port == 0 {
		panic("Port must be specified")
	}

	if *serverTomlFile == "" {
		panic("Server config file must be specified")
	}

	s := NewServer(*port, *serverTomlFile)

	go listenSignals(s)

	s.Start()
}
