# zenroute

A minimal, local DNS/TLS proxy for macOS written in Go. Bypasses elementary DPI firewalls and DNS poisoning (for now only Discord, planning making this dynamic later) using DoH and TLS ClientHello fragmentation.

## Features

- **DoH Resolution**: Uses Cloudflare DoH to bypass poisoned local DNS.
- **TLS Fragmentation**: To bypass DPI control, the initial TLS ClientHello packet is split into configurable chunks.
- **Auto Proxy Management**: Automatically toggles macOS `networksetup` on start and graceful exit.

## Usage

Requires Go 1.25+ and macOS. (I'll add windows support later with WinDivert)

```bash
make run
```

The system proxy will automatically bind to `localhost:8080` for the `Wi-Fi` interface. Hit `Ctrl+C` to unbind and exit.

## Config

Environment variables:

- `PROXY_ADDR` (default: `127.0.0.1`)
- `PROXY_PORT` (default: `8080`)
- `SYSTEM_SERVICE` (default: `Wi-Fi`)
