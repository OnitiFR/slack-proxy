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
	port              int
	configFileName    string
	Secret            string
	BaseUrl           string
	selfClient        *Client
	Clients           map[string]*Client
	SlackChannels     map[string]*SlackChannel
	mutex             *sync.Mutex
	nbmessages        int
	nberrors          int
	lastSelfNotify    time.Time
	notifyRoute       string
	selfnotifychannel string
	selfCheckDuration time.Duration
}

type ServerToml struct {
	Secret               string
	BaseUrl              string
	NotifyRoute          string
	SelfNotifyChannel    string
	SelfCheckEveryXHours int
	Clients              []*Client
	SlackChannels        []*SlackChannel
}

// start the server and listen for incoming requests
func (s *Server) Start() {
	fmt.Println("Starting server on port", s.port)

	http.HandleFunc(s.notifyRoute, s.NotifyChannel)

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
	if status == http.StatusOK {
		s.nbmessages++
	} else {
		s.nberrors++
	}

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

	if s.NotifyRoute == "" {
		return fmt.Errorf("notify route must be specified")
	}

	if s.SelfNotifyChannel == "" {
		return fmt.Errorf("self notify channel must be specified")
	}

	if s.SelfCheckEveryXHours <= 0 {
		return fmt.Errorf("self check every x hours must be specified")
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

	// create self client
	server.selfClient = &Client{
		Name: "self",
	}

	s.Clients = append(s.Clients, server.selfClient)

	if len(meta.Undecoded()) > 0 {
		return fmt.Errorf("unknown fields in config file: %v", meta.Undecoded())
	}

	server.Secret = s.Secret
	server.BaseUrl = s.BaseUrl
	server.Clients = make(map[string]*Client)
	server.SlackChannels = make(map[string]*SlackChannel)
	server.notifyRoute = fmt.Sprintf("/%s/", s.NotifyRoute)
	server.selfnotifychannel = s.SelfNotifyChannel
	server.selfCheckDuration = time.Duration(s.SelfCheckEveryXHours) * time.Hour

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
				webhookUrl := fmt.Sprintf("%s%s%s/%s", s.BaseUrl, server.notifyRoute, client.Token, channel.Token)
				client.Webhooks[channel.Name] = webhookUrl
			}
		}

	}

	for _, client := range s.Clients {
		server.Clients[client.Token] = client

	}

	selfChannelExists := false
	for _, channel := range s.SlackChannels {
		server.SlackChannels[channel.Token] = channel

		if channel.Name == server.selfnotifychannel {
			selfChannelExists = true
		}
	}

	if !selfChannelExists {
		return fmt.Errorf("self notify channel does not exist")
	}

	return nil
}

// Notify self to inform the server is still running
func (s *Server) SelfNotify() {
	payload := map[string]string{
		"text": fmt.Sprintf("I'm still running. %d message(s) sent :tada:, %d error(s) :doh: since %s", s.nbmessages, s.nberrors, time.Duration(time.Since(s.lastSelfNotify))),
	}

	jsonRequest, err := json.Marshal(payload)

	if err != nil {
		fmt.Println("Error while marshalling json request", err)
		return
	}
	fmt.Println(s.selfClient.Webhooks[s.selfnotifychannel])

	res, err := http.Post(s.selfClient.Webhooks[s.selfnotifychannel], "application/json", bytes.NewBuffer(jsonRequest))

	if err != nil {
		fmt.Println("Error creating request", err)
		return
	}

	if res.StatusCode != http.StatusOK {
		fmt.Println("Error while sending self notify", res.StatusCode)
		fmt.Println(res.Body)
		return
	}

	s.lastSelfNotify = time.Now()
	s.nberrors = 0
	s.nbmessages = 0
}

// check if the server is still running
func (s *Server) SelfCheck() {
	for {
		time.Sleep(s.selfCheckDuration)
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
