# XSSHunter-go

XSSHunter-go is a self-hosted XSS hunter that allows you to create a custom XSS payload and track when it is triggered. It is a based off the original [XSSHunter-express](https://github.com/mandatoryprogrammer/xsshunter-express) but written in Go.

| [Setup](#setup) | [Environment Variables](#environment-variables) | [Volumes](#volumes) | [Notifications](#notifications) |

The idea of why I decided to code this in Go is because I wanted this to be a maintained project that is stable. The original is based of Node 12 which is end of life. I also wanted to add some features that I thought would be useful (mostly expanding the notification system).

## Setup

```yml
version: '3'
services:
  xsshunter-go:
    image: adamjsturge/xsshunter-go:main
    volumes:
      - ./db/:/app/db/
      - ./screenshots/:/app/screenshots/
    ports:
      - "1449:1449"
    environment:
      - CONTROL_PANEL_ENABLED=true
      - NOTIFY=discord://...,slack://...
```

## Environment Variables

| Name | Description | Default |
| --- | --- | --- |
| CONTROL_PANEL_ENABLED | Enable the control panel | false |
| NOTIFY | Comma separated list of notification URLs |  |
| SCREENSHOTS_REQUIRE_AUTH | Require authentication to view screenshots | false |
| DOMAIN | Domain put into script | Based off URL |
| DATABASE_URL | Postgres Database URL | (Uses sqlite if no postgres db is present) |

## Volumes

| Name | Description |
| --- | --- |
| /app/db/ | Database storage if using SQLite |
| /app/screenshots/ | Screenshot storage |

## Notifications

Notifcations use from shoutrrr
To make your own notification URL, go to https://containrrr.dev/shoutrrr/v0.8/services/overview/ or see these examples:

| Service | URL |
| --- | --- |
| Discord | discord://`token`@`id` |
| Slack | slack://\[`botname`@\]`token-a`/`token-b`/`token-c` |
| Telegram | telegram://... |

