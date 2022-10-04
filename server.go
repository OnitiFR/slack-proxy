package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type Server struct {
	port          int
	channelsDir   string
	clientsDir    string
	clients       map[string]*Client
	slackChannels map[string]*SlackChannel
	mutex         *sync.Mutex
}

// load channels from the channels directory
func (s *Server) LoadChannels() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// reset all channels
	s.slackChannels = make(map[string]*SlackChannel)

	files, err := ioutil.ReadDir(s.channelsDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".toml") {
			continue
		}

		channel, err := NewSlackChannelFromToml(fmt.Sprintf("%s/%s", s.channelsDir, file.Name()))
		if err != nil {
			return err
		}

		// check if chan already exists
		if _, ok := s.slackChannels[channel.Name]; ok {
			return fmt.Errorf("channel %s already exists", channel.Name)
		}

		s.slackChannels[channel.Name] = channel
	}

	return nil
}

// load clients from the clients directory
func (s *Server) LoadClients() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// reset all clients
	s.clients = make(map[string]*Client)

	files, err := ioutil.ReadDir(s.clientsDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".toml") {
			continue
		}

		client, err := NewClientFromToml(fmt.Sprintf("%s/%s", s.clientsDir, file.Name()))

		if err != nil {
			return err
		}

		// check if client already exists
		if _, ok := s.clients[client.AuthorisationToken]; ok {
			return fmt.Errorf("client %s already exists", client.Name)
		}

		s.clients[client.AuthorisationToken] = client
	}

	return nil
}

// start the server and listen for incoming requests
func (s *Server) Start() {
	fmt.Println("Starting server on port", s.port)

	http.HandleFunc("/channels", s.DisplayChannels)
	http.HandleFunc("/clients", s.DisplayClients)
	http.HandleFunc("/notify", s.NotifyChannel)

	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
	if err != nil {
		panic(err)
	}

}

func NewServer(port int, channelsDir string, clientsDir string) *Server {
	return &Server{
		port:          port,
		channelsDir:   channelsDir,
		clientsDir:    clientsDir,
		clients:       make(map[string]*Client),
		slackChannels: make(map[string]*SlackChannel),
		mutex:         &sync.Mutex{},
	}
}
