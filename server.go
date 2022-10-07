package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

type Server struct {
	port           int
	configFileName string
	Secret         string
	BaseUrl        string
	selfClient     *Client
	Clients        map[string]*Client
	SlackChannels  map[string]*SlackChannel
	mutex          *sync.Mutex
	nbmessages     int
	nberrors       int
	lastSelfNotify time.Time
}

type ServerToml struct {
	Secret        string
	BaseUrl       string
	Clients       []*Client
	SlackChannels []*SlackChannel
}

// start the server and listen for incoming requests
func (s *Server) Start() {
	fmt.Println("Starting server on port", s.port)

	http.HandleFunc("/notify/", s.NotifyChannel)

	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
	if err != nil {
		panic(err)
	}
}

// display all the clients and their webhooks
func (s *Server) DisplayClientsRoutes() {
	for _, client := range s.Clients {
		fmt.Println(client.Name)
		for channelName, webhookUrl := range client.Webhooks {
			fmt.Printf("  - %s : %s\n", channelName, webhookUrl)
		}
	}
}

// return message with status code
func (s *Server) Reponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	w.Write([]byte(message))
}

// load the server config from a toml file
func (server *Server) LoadConfig() error {
	var s ServerToml
	meta, err := toml.DecodeFile(server.configFileName, &s)

	if err != nil {
		return err
	}
	if s.BaseUrl == "" {
		return fmt.Errorf("base url must be specified")
	}

	if s.Secret == "" {
		return fmt.Errorf("secret must be specified")
	}

	// Check if the slack channels are valid
	for _, channel := range s.SlackChannels {
		if channel.Name == "" {
			return fmt.Errorf("slack channel name must be specified")
		}

		if channel.WebhookUrl == "" {
			return fmt.Errorf("slack channel webhook url must be specified")
		}

		// md5 hash channel name + secret
		channel.Token = fmt.Sprintf("%x", md5.Sum([]byte(channel.Name+s.Secret)))
	}

	// Check if the clients are valid
	for _, client := range s.Clients {
		if client.Name == "" {
			return fmt.Errorf("client name must be specified")
		}

		// md5 hash channel name + secret
		client.Token = fmt.Sprintf("%x", md5.Sum([]byte(client.Name+s.Secret)))

		client.Webhooks = make(map[string]string)
		for _, channel := range s.SlackChannels {
			if client.IsAllowedChannel(channel.Name) {
				webhookUrl := fmt.Sprintf("%s/notify/%s/%s", s.BaseUrl, client.Token, channel.Token)
				client.Webhooks[channel.Name] = webhookUrl
			}
		}

	}

	if len(meta.Undecoded()) > 0 {
		return fmt.Errorf("unknown fields in config file: %v", meta.Undecoded())
	}

	server.Secret = s.Secret
	server.BaseUrl = s.BaseUrl
	server.Clients = make(map[string]*Client)
	server.SlackChannels = make(map[string]*SlackChannel)

	for _, client := range s.Clients {
		if client.Name == "self" {
			server.selfClient = client
		}
		server.Clients[client.Token] = client

	}

	for _, channel := range s.SlackChannels {
		server.SlackChannels[channel.Token] = channel
	}

	return nil
}

// Notify self to inform the server is still running
func (s *Server) SelfNotify() {
	payload := map[string]string{
		"text": fmt.Sprintf("Oniti Proxy is still running. %d message(s) sent :tada:, %d error(s) :doh: since %s", s.nbmessages, s.nberrors, time.Duration(time.Since(s.lastSelfNotify))),
	}

	jsonRequest, err := json.Marshal(payload)

	if err != nil {
		fmt.Println("Error while marshalling json request", err)
		return
	}
	_, err = http.Post(s.selfClient.Webhooks["notifs"], "application/json", bytes.NewBuffer(jsonRequest))

	if err != nil {
		fmt.Println("Error creating request", err)
		return
	}

	s.lastSelfNotify = time.Now()
	s.nberrors = 0
	s.nbmessages = 0
}

// check if the server is still running
func (s *Server) SelfCheck() {
	for {
		time.Sleep(24 * time.Hour) // check every 24 hours
		s.SelfNotify()
	}
}

func NewServer(port int, serverTomlFile string) *Server {
	s := &Server{
		port:           port,
		configFileName: serverTomlFile,
		Clients:        make(map[string]*Client),
		SlackChannels:  make(map[string]*SlackChannel),
		mutex:          &sync.Mutex{},
		nbmessages:     0,
		nberrors:       0,
		lastSelfNotify: time.Now(),
	}

	err := s.LoadConfig()

	if err != nil {
		panic(err)
	}

	go s.SelfCheck()

	return s
}
