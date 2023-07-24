# revealer

The repository for manage a container image of HTTP server that returns the source IP, request headers, request body, and query string.
It is used to check request data and to isolate problems when you want to check whether a request is correctly sent to a server via VPN or Proxy, but cannot be checked by the server in question.

You can check the operation at [revealer.zuki.dev](https://revealer.zuki.dev) or [r.zuki.dev](https://r.zuki.dev).

## Run

Using docker:

```sh
docker run -p 8080:8080 --rm -it ghcr.io/corrupt952/revealer
```

## Configurations

### Environment variables

* PORT ... Start up on the specified por.t Set with environment variable. (Default: 8080)
