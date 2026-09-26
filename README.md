# Defense Matrix

Defense Matrix is a Go HTTP proxy that routes traffic through Docker-managed containers and scans requests for configured signatures.

The gateway uses Coraza with the OWASP Core Rule Set to block recognized HTTP attack requests before they reach a boat. This covers common web attack patterns, but it is not a network firewall or IDS and cannot inspect non-HTTP traffic or guarantee detection of every tool or payload. Use network-level firewall/IDS controls separately.

## Inspiration

The name and threat-response idea are inspired by Wamai from *Tom Clancy's Rainbow Six Siege*. His Mag-NET System pulls incoming projectiles toward its device, where they detonate. This project adapts that idea to HTTP traffic by scanning requests and rotating isolated backend containers when a threat is detected. Defense Matrix is an independent project and is not affiliated with Ubisoft.

## Requirements

- Docker Desktop or Docker Engine with Docker Compose v2
- Linux containers enabled in Docker Desktop

The engine uses the Docker socket to manage the backend containers. Access to that socket grants control over the Docker host.

## Run with Docker Compose

Clone the repository, then enter its folder. Start Docker Desktop first on Windows.

```sh
git clone https://github.com/moredig/defensematrix.git
cd defensematrix
```

Build and start the stack from the repository root. This keeps live logs in the terminal:

```sh
docker compose -f deploy/docker-compose.yml up --build
```

The initial build downloads base images and may take a few minutes. On later runs, start the already-built images with:

```sh
docker compose -f deploy/docker-compose.yml up
```

Both commands stay attached and show live logs. Press `Ctrl+C` to stop the stack. Add `--build` again after changing source or Dockerfiles.

Published web traffic is restricted to this computer by default. To allow access from another device on your LAN, set this computer's LAN address before starting Compose:

```powershell
$env:WAMAI_BIND_ADDRESS = "192.168.1.25"
docker compose -f deploy/docker-compose.yml up
```

Replace the example address with this computer's LAN address.

HTTP request bodies larger than 1 MiB are rejected. Requests that match an active WAF blocking rule receive HTTP `403` and are not forwarded to a boat.

Stop Wamai and remove its containers and network with:

```sh
docker compose -f deploy/docker-compose.yml down
```

This keeps the built images and your project files, configuration, and logs. To also remove the images built for Wamai and its boats, run:

```sh
docker compose -f deploy/docker-compose.yml down --rmi all
```

Do not expose this setup to untrusted networks. The engine has access to the Docker socket and manages the backend containers.

## Disclaimer
This project was built using Artificial Intelligence. This is simply an idea I had and really eanted to put out there. This is not meant to be used as a massive, production grade tool. Thank you.

## Project Contents

- `cmd/wamai`: Go application entry point
- `internal/`: HTTP handling, container orchestration, fleet management, and request scanning
- `deploy/`: Dockerfiles and Compose configuration
- `config/`: application profiles and settings
- `cli/`: terminal log monitor
- `logs/`: engine log output

