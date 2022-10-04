# Oniti Slack Proxy

Proxy for slack notifications.

## Options

 - p : port to listen on, default 8080
 - s : directory to load channels from, default channels
 - c : directory to load clients from, default clients

## Signals

 - USR1 : redifine configuration

 ## Usage
Send a post to ```/notify```
add Authorisation Header
fields :
 - channel : Channel targeted
 - message: message to send
