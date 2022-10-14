package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type AttachmentRequest struct {
	Fallback  string             `json:"fallback"`
	Text      string             `json:"text"`
	Color     string             `json:"color"`
	Mrkdwn_in []string           `json:"mrkdwn_in"`
	Ts        int64              `json:"ts"`
	Title     string             `json:"title"`
	Fields    []*AttachmentField `json:"fields"`
}

type NotifyRequest struct {
	Text        string               `json:"text"`
	Username    string               `json:"username"`
	Attachments []*AttachmentRequest `json:"attachments"`
	Icon_emoji  string               `json:"icon_emoji"`
}

// Main function, sends a message to a channel
// check if the client is allowed to send a message to the channel
func (s *Server) NotifyChannel(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Only accept POST
	if r.Method != "POST" {
		s.Reponse(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	routeParts := strings.Split(r.URL.Path, "/")

	if len(routeParts) != 4 {
		s.Reponse(w, http.StatusBadRequest, "Invalid route")
		return
	}
	clientToken := routeParts[2]
	channelName := routeParts[3]

	var request NotifyRequest

	// Read the request body
	text := r.FormValue("text")
	if text != "" {
		request.Text = text
	} else {
		contentType := r.Header.Get("Content-type")
		jsonString := ""

		if contentType == "application/x-www-form-urlencoded" {
			err := r.ParseForm()
			if err != nil {
				s.Reponse(w, http.StatusBadRequest, "Invalid request : "+err.Error())
				return
			}

			for _, value := range r.PostForm {
				jsonString = value[0]
				break
			}

		} else {
			// Try to parse the request as JSON
			b, errParse := ioutil.ReadAll(r.Body)
			if errParse != nil {
				s.Reponse(w, http.StatusBadRequest, "Invalid request : "+errParse.Error())
				return
			}
			jsonString = string(b)

		}

		if jsonString == "" {
			s.Reponse(w, http.StatusBadRequest, "Empty request")
			return
		}

		err := json.Unmarshal([]byte(jsonString), &request)
		if err != nil {
			s.Reponse(w, http.StatusBadRequest, "Error parsing JSON : "+err.Error())
			return
		}

	}

	// Get client from token
	client, okClient := s.Clients[clientToken]

	if !okClient {
		s.Reponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// check if channel exists
	channel, ok := s.SlackChannels[channelName]
	if !ok {
		s.Reponse(w, http.StatusNotFound, "Channel not found")
		return
	}

	// check if client is allowed to send a message to this channel
	chanAllowed := client.IsAllowedChannel(channel.Name)
	if !chanAllowed {
		message := fmt.Sprintf("channel %s not allowed for client %s", channelName, client.Name)
		s.Reponse(w, http.StatusUnauthorized, message)
		return
	}

	// send message
	err := channel.SendMessage(&request, client)
	if err != nil {
		s.Reponse(w, http.StatusInternalServerError, "Error sending message : "+err.Error())
		return
	}

	s.Reponse(w, http.StatusOK, "Message sent")
}
