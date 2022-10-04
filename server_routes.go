package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type NotifyRequest struct {
	Text string
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

	routePath := r.URL.Path[len("/notify/"):]
	routeParts := strings.Split(routePath, "/")

	if len(routeParts) != 2 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Invalid route"))
		return
	}
	clientToken := routeParts[0]
	channelName := routeParts[1]

	var request NotifyRequest

	// Read the request body
	text := r.FormValue("text")
	if text != "" {
		request.Text = text
	} else {
		// Try to parse the request as JSON
		b, errParse := ioutil.ReadAll(r.Body)
		if errParse != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid request"))
			return
		}

		json.Unmarshal(b, &request)

	}

	// Get client from token
	client, okClient := s.Clients[clientToken]

	if !okClient {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	// we need a text
	if request.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("text is required"))
		return
	}

	// check if channel exists
	channel, ok := s.SlackChannels[channelName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("channel not found"))
		return
	}

	// check if client is allowed to send a message to this channel
	chanAllowed := client.IsAllowedChannel(channel.Name)
	if !chanAllowed {
		w.WriteHeader(http.StatusUnauthorized)
		message := fmt.Sprintf("channel %s not allowed for client %s", channelName, client.Name)
		w.Write([]byte(message))
		return
	}

	// send message
	err := channel.SendMessage(&request, client)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("message sent"))
}
