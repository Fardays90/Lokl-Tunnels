# Reverse Proxy Tunnel

This project creates a lightweight tunnel service that lets you expose your local web server (like `localhost:5500`) to the internet through a public-facing Go server. It works similarly to services like **ngrok** but is built with Go and WebSockets.

## How It Works

1. **Client CLI** (in `client/`):
   - Sends a Dial request to the public server to initiate a tunnel.
   - Establishes a persistent **WebSocket** connection to the server.
   - Listens for incoming HTTP requests (proxied over the WebSocket).
   - Forwards them to your local server (e.g., `localhost:5500`) and sends the response back.

2. **Server** (in `server/`):
   - Handles external HTTP requests (e.g., `https://your-server.com/abc12/`).
   - Identifies the WebSocket tunnel based on the path ID.
   - Forwards the request to the connected client.
   - Sends the response back to the requester.

After running client.go do

```bash
http --port <port>
```
to expose the local port to the server. That's all.

---


