# XSSHunter-go

![](https://github.com/adamjsturge/xsshunter-go/blob/main/.github/setup-xsshunter-go.webp?raw=true)

XSSHunter-go is a self-hosted XSS hunter that allows you to create a custom XSS payload and track when it is triggered. It is a based off the original [XSSHunter-express](https://github.com/mandatoryprogrammer/xsshunter-express) but written in Go.

<!-- Table of content -->
<summary>Table of content</summary>
<ol>
    <li><a href="#setup">Setup</a></li>
    <li><a href="#volumes">Volumes</a></li>
    <li><a href="#notifications">Notifications</a></li>
    <li><a href="#using-traefik-for-ssl">Using Traefik for SSL</a></li>
    <li><a href=>
</ol>

The idea of why I decided to code this in Go is because I wanted this to be a maintained project that is stable. The original is based of Node 12 which is end of life. I also wanted to add some features that I thought would be useful (mostly expanding the notification system).

## Setup

```yml
version: '3.8'
services:
  xsshunter-go:
    image: adamjsturge/xsshunter-go:latest
    volumes:
      - ./db/:/app/db/
      - ./screenshots/:/app/screenshots/
    ports:
      - "1449:1449"
    environment:
      - CONTROL_PANEL_ENABLED=true
      - NOTIFY=discord://...,slack://...
```

Add the `S` to `HTTPS` with the <a href="#using-traefik-for-ssl">traefik guide</a> below

Please Reference https://github.com/adamjsturge/xsshunter-go/tree/main/docker-examples
The one wihtout a folder https://github.com/adamjsturge/xsshunter-go/blob/main/docker-examples/docker-compose.yml is for sqlite

For Postgres and cache (cache currently not implemented) https://github.com/adamjsturge/xsshunter-go/blob/main/docker-examples/example-postgres-cache/docker-compose.yml

https://github.com/adamjsturge/xsshunter-go/wiki/FAQ#why-use-postgres-over-sqlite

## Environment Variables

| Name | Description | Default |
| --- | --- | --- |
| CONTROL_PANEL_ENABLED | Enable the control panel | false |
| NOTIFY | Comma separated list of notification URLs |  |
| SCREENSHOTS_REQUIRE_AUTH | Require authentication to view screenshots | false |
| DOMAIN | Domain put into script | Based off URL (Defaults to HTTPS) |
| DATABASE_URL | Postgres Database URL | (Uses sqlite if no postgres db is present) |
| GO_ENV | Go environment `development` |  |
| ENFORCE_CERT_FROM_GOLANG | Force Golang to serve a cert (Only do if you know what you're doing) | false |

## Volumes

| Name | Description |
| --- | --- |
| /app/db/ | Database storage if using SQLite |
| /app/screenshots/ | Screenshot storage |

## Notifications

For notifications xsshunter-go uses shoutrrr.
To make your own notification URL, go to https://containrrr.dev/shoutrrr/v0.8/services/overview/ or see these examples:

| Service | URL |
| --- | --- |
| Discord | discord://`token`@`id` |
| Slack | slack://\[`botname`@\]`token-a`/`token-b`/`token-c` |
| Telegram | telegram://... |

## Using Traefik for SSL

```yml
version: '3.9'
services:
    xsshunter-go:
        image: adamjsturge/xsshunter-go:latest
        volumes:
            - ./db/:/app/db/
            - ./screenshots/:/app/screenshots/
        environment:
            - CONTROL_PANEL_ENABLED=true
            - NOTIFY=discord://...,slack://...
            - DOMAIN=https://xsshunter.example.com
        labels:
            - "traefik.enable=true"
            - "traefik.http.routers.xsshunter-go.entrypoints=web, websecure"
            - "traefik.http.routers.xsshunter-go.rule=Host(`xsshunter.example.com`)"
            - "traefik.http.routers.xsshunter-go.tls.certresolver=myresolver"
            - "traefik.http.routers.xsshunter-go.tls.domains[0].main=xsshunter.example.com"
            - "traefik.http.services.xsshunter-go.loadbalancer.server.port=1449"
    traefik:
        image: "traefik:v2.8"
        container_name: "traefik"
        command:
            - "--providers.docker=true"
            - "--providers.docker.exposedbydefault=false"
            - "--entrypoints.web.address=:80"
            - "--entrypoints.websecure.address=:443"
            - "--certificatesresolvers.myresolver.acme.httpchallenge=true"
            - "--certificatesresolvers.myresolver.acme.httpchallenge.entrypoint=web"
            - "--certificatesresolvers.myresolver.acme.email=xsshunter@example.com"
            - "--certificatesresolvers.myresolver.acme.storage=/shared/acme.json"
        ports:
            - "80:80"
            - "443:443"
        volumes:
            - "/var/run/docker.sock:/var/run/docker.sock:ro"
            - "./shared:/shared"
```

## Contributing

Thanks for looking to contribute.

If you have a massive feature change I recommend putting as an issue before starting anything! That way we can discuss it before you start working on it.

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for more information on how to contribute and everything you need to know.