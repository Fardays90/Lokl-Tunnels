# Lokl Tunnels

This project lets you expose local servers (e.g. `localhost:3000`) to the internet via a reverse proxy tunnel. Currently only supports http and https. It functions similarly to tools like **ngrok**, but is lightweight and requires no account creation. Made using **Go** and **WebSockets**

---

## Usage 
```
npx lokl-cli
```
Or directly download the binary from [here](client/bin/) lokl-cli for linux 64 bit arch and lokl-cli.exe for windows 64 bit arch

## How It Works

- **Client** (`client/`):
  - Sends request to `ws://tunnels.fardays.com/connect` for upgrading http to websocket connection via cli.
  - Authenticates request using a token.
  - Receives a unique subdomain (e.g. `abcde.tunnels.fardays.com`).
  - Listens for proxied requests via WebSocket and forwards them to your local service (e.g. `localhost:5173`).
  - Sends the local server’s response back to the tunnel server.

- **Server** (`server/`):
  - Accepts WebSocket connections at `/connect`.
  - Assigns a random subdomain (e.g., `randomid.tunnels.fardays.com`).
  - Proxies HTTP requests to the correct connected client based on subdomain.
  - Forwards the client's response to the original requester.

## Traffic Routing

- All traffic to `*.tunnels.fardays.com` is routed through cloudflare, a Nginx container and the websocket server running on my homelab.
- While all traffic passes through my server, I do not log anything beyond what’s needed for routing. The entire goal of this project was to learn something new and let me use my own infra to expose dev servers and other resources to people.

---
