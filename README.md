# Turnify - Cloudflare TURN Server Proxy for Matrix

**THIS PROJECT IS HEAVILY INSPIRED BY [github.com/bpbradley/matrix-turnify](https://github.com/bpbradley/matrix-turnify).**

Turnify is a middleman proxy service that allows Matrix clients to use Cloudflare's TURN server for voice/video calling. It acts as a reverse proxy, ensuring proper authentication and providing valid TURN credentials from Cloudflare for authenticated users.

This project serves as an alternative to self-hosting your own TURN server or using other third-party providers.

## Background

Matrix, the decentralized communication protocol, supports the configuration of self-hosted TURN servers for voice and video calls. However, the [TURN server REST API](https://tools.ietf.org/html/draft-uberti-behave-turn-rest-00) provided by Matrix doesn't work directly with Cloudflare Calls, which generates short-lived credentials for TURN servers.

Turnify solves this issue by acting as an intermediary between Matrix clients and Cloudflare's TURN API. It authenticates the user, requests TURN credentials from Cloudflare, and proxies the response back to the client.

This setup has several advantages:

1. **Robust Security**: Cloudflare handles the security of your TURN server, minimizing your responsibilities.
2. **Global Scaling**: Cloudflare's infrastructure ensures users are connected to the nearest TURN server for optimal performance.
3. **Free Data Transfer**: Cloudflare offers 1,000 GB/month of free data transfer, making this a cost-effective solution for modest usage.

## How It Works

Turnify works by intercepting requests to paths ending in `/voip/turnServer` and handling both the `OPTIONS` and `GET` HTTP methods:

- **OPTIONS**: Always proxied as is. This is necessary for proper CORS handling.
- **GET**: Proxies the request to authenticate the user. If successful, it calls the Cloudflare API to retrieve TURN credentials and forwards them back to the client.

## Requirements

- A **reverse proxy** (e.g., Traefik, Nginx) that can route requests based on path (e.g., `^.*/voip/turnServer$`).
- A **valid API key** and **Application Token** for a TURN app in your Cloudflare account. You can create these in your Cloudflare dashboard under **Calls > Create > Turn App**.

## Setup

### Docker Compose Example

Here’s an example of how to set up Turnify using Docker Compose with Traefik as the reverse proxy.

```yaml
services:
  turnify:
    image: ghcr.io/mister-mr-matrix/turnify:latest
    restart: unless-stopped
    networks:
      - traefik
    environment:
      - TURNIFY_DEBUG=true
      - TURNIFY_MATRIX_HOMESERVER_URL=http://tuwunel:8008
      - TURNIFY_CF_TURN_TOKEN_ID=your_token_id
      - TURNIFY_CF_TURN_API_TOKEN=your_api_token
      - TURNIFY_TURN_CREDENTIAL_TTL_SECONDS=86400
    labels:
      - traefik.enable=true
      - traefik.docker.network=traefik
      - traefik.http.routers.turnify.entrypoints=https
      - traefik.http.services.turnify.loadbalancer.server.port=4499
      - traefik.http.routers.turnify.rule=Host(`matrix.example.com`) && PathRegexp(`^.*/voip/turnServer$`)
      - traefik.http.routers.turnify.priority=32
      - traefik.http.routers.turnify.tls=true
networks:
  traefik:
    external: true
```

### Environment Variables

| Variable                              | Description                                   | Default Value         | Required |
| ------------------------------------- | --------------------------------------------- | --------------------- | -------- |
| `TURNIFY_DEBUG`                       | Enables debug logging                         | `false`               | No       |
| `TURNIFY_PORT`                        | Port to run Turnify service on                | `4499`                | No       |
| `TURNIFY_MATRIX_HOMESERVER_URL`       | Matrix homeserver URL for user authentication | `http://tuwunel:8008` | No       |
| `TURNIFY_CF_TURN_TOKEN_ID`            | Cloudflare TURN Token ID                      | (None)                | Yes      |
| `TURNIFY_CF_TURN_API_TOKEN`           | Cloudflare TURN API Token                     | (None)                | Yes      |
| `TURNIFY_TURN_CREDENTIAL_TTL_SECONDS` | TTL for TURN credentials in seconds           | `86400`               | No       |

### Docker `.env` Example

```env
TURNIFY_DEBUG=true
TURNIFY_MATRIX_HOMESERVER_URL=http://tuwunel:8008
TURNIFY_CF_TURN_TOKEN_ID=your_turn_token_id
TURNIFY_CF_TURN_API_TOKEN=your_turn_api_token
TURNIFY_TURN_CREDENTIAL_TTL_SECONDS=86400
```

### Reverse Proxy Setup

Make sure your reverse proxy routes requests ending in `/voip/turnServer` to Turnify. Below is an example for Traefik, but other proxies like Nginx or Caddy should work similarly.

```
traefik.http.routers.turnify.rule=Host(`matrix.example.com`) && PathRegexp(`^.*/voip/turnServer$`)
```

Adjust the configuration for your own reverse proxy setup, ensuring it routes only requests to the `^.*/voip/turnServer$` endpoint to Turnify.

## Testing

You can test the functionality with a simple `curl` request. Ensure you provide a valid authorization token.

```sh
curl --header "Authorization: Bearer VALID_MATRIX_TOKEN" \
  -X GET https://matrix.example.com/_matrix/client/v3/voip/turnServer
```

If the request is authenticated, Turnify will return Cloudflare TURN credentials. If not, you’ll receive an unauthorized response from your Matrix homeserver.

To enable detailed logs, set the `TURNIFY_DEBUG=true` environment variable.

## Notes

- **Security**: This service doesn't store or handle user credentials directly. It only proxies the `GET` and `OPTIONS` requests to Matrix, retrieves Cloudflare TURN credentials, and returns them to the client.
- **Cloudflare Terms**: When using this service, you're subject to Cloudflare's terms of service for their TURN API. Ensure you comply with their usage policies.
- **Support**: This is a community-driven, open-source project. It’s not officially supported by the Matrix, Element, or Synapse teams.

## License

This project is open-source and available under the MIT License. See the [LICENSE](LICENSE) file for more information.
