# Oniti Slack Proxy

Proxy for slack notifications.

## Options

 - p : port to listen on, default 8080

## Signals

 - USR1 : redifine configuration

 ## Usage
Send a post to ```/notify```
add Authorisation Header
fields :
 - channel : Channel targeted
 - message: message to send
