# Thrempp
[![pipeline status](https://dev.sum7.eu/sum7/thrempp/badges/master/pipeline.svg)](https://dev.sum7.eu/genofire/thrempp/pipelines)
[![coverage report](https://dev.sum7.eu/sum7/thrempp/badges/master/coverage.svg)](https://dev.sum7.eu/genofire/thrempp/pipelines)
[![Go Report Card](https://goreportcard.com/badge/dev.sum7.eu/sum7/thrempp)](https://goreportcard.com/report/dev.sum7.eu/genofire/thrempp)
[![GoDoc](https://godoc.org/dev.sum7.eu/sum7/thrempp?status.svg)](https://godoc.org/dev.sum7.eu/genofire/thrempp)
 [![](https://inverse.chat/badge.svg?room=thrempp@conference.chat.sum7.eu)](https://conversations.im/j/thrempp@conference.chat.sum7.eu)

XMPP - Transport

ATM planned support for Threema only

## Get thrempp

#### Download

Latest Build binary from ci here:

[Download All](https://dev.sum7.eu/sum7/thrempp/-/jobs/artifacts/master/download/?job=build-my-project) (with config example)

[Download Binary](https://dev.sum7.eu/sum7/thrempp/-/jobs/artifacts/master/raw/bin/thrempp?inline=false&job=build-my-project)

#### Build

```bash
go get -u dev.sum7.eu/sum7/thrempp
```

## Configure

see `config_example.toml`

## Start / Boot

_/lib/systemd/system/thrempp.service_ :
```
[Unit]
Description=thrempp
After=network.target
# After=ejabberd.service
# After=prosody.service

[Service]
Type=simple
# User=notRoot
ExecStart=/opt/go/bin/thrempp serve --config /etc/thrempp.conf
Restart=always
RestartSec=5sec

[Install]
WantedBy=multi-user.target
```

Start: `systemctl start thrempp`
Autostart: `systemctl enable thrempp`


