package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Returns a list of all channels
func (s *Server) DisplayChannels(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	channels := make([]*SlackChannelJson, 0)
	for _, channel := range s.slackChannels {
		channels = append(channels, NewSlackChannelJson(channel))
	}

	json.NewEncoder(w).Encode(channels)
}

// Returns a list of all clients
func (s *Server) DisplayClients(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	clients := make([]*ClientJson, 0)
	for _, client := range s.clients {
		clients = append(clients, NewClientJson(client))
	}

	json.NewEncoder(w).Encode(clients)
}

// Main function, sends a message to a channel
// check if the client is allowed to send a message to the channel
func (s *Server) NotifyChannel(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Only accept POST
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	channelName := r.FormValue("channel")
	message := r.FormValue("message")

	// Get token from header
	authorisationToken := r.Header.Get("Authorization")

	// Get client from token
	client, okClient := s.clients[authorisationToken]

	if !okClient {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	// we need a channel and a message
	if channelName == "" || message == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("channel and message are required"))
		return
	}

	// check if channel exists
	channel, ok := s.slackChannels[channelName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("channel not found"))
		return
	}

	// check if client is allowed to send a message to this channel
	chanAllowed := client.IsAllowedChannel(channelName)
	if !chanAllowed {
		w.WriteHeader(http.StatusUnauthorized)
		message := fmt.Sprintf("channel %s not allowed for client %s", channelName, client.Name)
		w.Write([]byte(message))
		return
	}

	// send message
	err := channel.SendMessage(message, client)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("message sent"))
}
