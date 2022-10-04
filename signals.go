package main

import (
	"os"
	"os/signal"
	"syscall"
)

func listenSignals(server *Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		signal := <-c
		switch signal {
		case syscall.SIGUSR1:
			server.LoadConfig()
			server.DisplayClientsRoutes()

		case syscall.SIGUSR2:
			server.DisplayClientsRoutes()

		}
	}
}
