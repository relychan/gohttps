# gohttps

gohttps includes reusable functions to create HTTP servers in Go. The library aims to simplify HTTP server creation in Go by providing common patterns and middleware out of the box, with strong observability and configuration management features.

## Key Features

- Router creation with sensible defaults using Chi router.
- Built-in middlewares: compression, decompression, CORS, request timeout, max body size.
- Observability support: OpenTelemetry integration, Prometheus metrics
- TLS support for HTTPS servers
- Configuration-driven setup via YAML/JSON with JSON schema validation.
