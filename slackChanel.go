package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BurntSushi/toml"
)

type SlackChannel struct {
	Name       string
	WebhookUrl string
}

type SlackChannelJson struct {
	Name string
}

func NewSlackChannelJson(slackChannel *SlackChannel) *SlackChannelJson {
	return &SlackChannelJson{slackChannel.Name}
}

func NewSlackChannelFromToml(filename string) (*SlackChannel, error) {
	var channel *SlackChannel
	_, err := toml.DecodeFile(filename, &channel)

	return channel, err
}

func NewSlackChannel(name string, webhookUrl string) *SlackChannel {
	return &SlackChannel{name, webhookUrl}
}

// send a message to the slack channel
func (c *SlackChannel) SendMessage(message string, client *Client) error {

	prefixMessage := fmt.Sprintf("*%s* :\n", client.Name)

	payload_json := map[string]string{
		"text": prefixMessage + message,
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
