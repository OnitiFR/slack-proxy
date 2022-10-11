package main

type Client struct {
	Name               string
	Icon               string
	AuthorizedChannels []string
	Token              string
	Webhooks           map[string]string
}

// check if the client is allowed to send a message to the channel
func (c *Client) IsAllowedChannel(channel string) bool {
	if len(c.AuthorizedChannels) == 0 {
		return true
	}

	for _, allowed := range c.AuthorizedChannels {
		if allowed == channel {
			return true
		}
	}

	return false
}
