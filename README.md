# Oniti Slack Proxy

Proxy for slack notifications.

## Options

 - p : port to listen on, default 8080
 - c : server config file

## Signals

 - USR1 : redifine configuration
 - USR2 : display all clients hooks

 ## Usage
Send a post to ```/notify```

fields :
 - text: message to send
