package main

import (
	"github.com/BurntSushi/toml"
)

type Client struct {
	Name               string
	AuthorisationToken string
	ChannelsAllowed    []string
}

type ClientJson struct {
	Name            string
	ChannelsAllowed []string
}

func NewClientJson(client *Client) *ClientJson {
	return &ClientJson{client.Name, client.ChannelsAllowed}
}

func NewClientFromToml(filename string) (*Client, error) {
	var client *Client
	_, err := toml.DecodeFile(filename, &client)

	return client, err
}

// check if the client is allowed to send a message to the channel
func (c *Client) IsAllowedChannel(channel string) bool {
	for _, allowed := range c.ChannelsAllowed {
		if allowed == channel {
			return true
		}
	}

	return false
}
