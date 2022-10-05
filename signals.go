package main

import (
	"fmt"
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
			err := server.LoadConfig()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Config reloaded")
			server.DisplayClientsRoutes()

		case syscall.SIGUSR2:
			server.DisplayClientsRoutes()

		}
	}
}
