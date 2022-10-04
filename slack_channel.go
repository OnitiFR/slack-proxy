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

	payload_json := map[string]string{
		"text": prefixMessage + request.Text,
	}

	payload, err := json.Marshal(payload_json)
	if err != nil {
		return err
	}

	res, err := http.Post(c.WebhookUrl, "application/json", bytes.NewBuffer(payload))
	if res != nil {
		defer res.Body.Close()
	}

	return err
}
