package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SlackChannel struct {
	Name       string
	WebhookUrl string
	Token      string
}

// send a message to the slack channel
func (c *SlackChannel) SendMessage(request *NotifyRequest, client *Client) error {

	prefixMessage := fmt.Sprintf("*%s* :\n", client.Name)

	request.Text = prefixMessage + request.Text

	// convert the request to json
	jsonRequest, err := json.Marshal(request)

	if err != nil {
		return err
	}
	res, err := http.Post(c.WebhookUrl, "application/json", bytes.NewBuffer(jsonRequest))
	if res != nil {
		defer res.Body.Close()
	}

	return err
}