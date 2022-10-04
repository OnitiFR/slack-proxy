package main

import "flag"

func main() {
	port := flag.Int("p", 8080, "port to listen on")
	channelsDir := flag.String("s", "channels", "directory to load channels from")
	clientsDir := flag.String("c", "clients", "directory to load clients from")

	flag.Parse()

	if *port == 0 {
		panic("Port must be specified")
	}

	if *channelsDir == "" {
		panic("Channel directory must be specified")
	}

	if *clientsDir == "" {
		panic("Clients directory must be specified")
	}

	s := NewServer(*port, *channelsDir, *clientsDir)

	err := s.LoadChannels()
	if err != nil {
		panic(err)
	}

	err = s.LoadClients()
	if err != nil {
		panic(err)
	}

	go listenSignals(s)

	s.Start()
}
