# zenroute

A minimal, local DNS/TLS proxy for macOS written in Go. Bypasses DPI firewalls and DNS poisoning using DoH and TLS ClientHello fragmentation.

## Features

- **DoH Resolution**: Uses Cloudflare DoH to bypass poisoned local DNS.
- **TLS Fragmentation**: To bypass DPI control, the initial TLS ClientHello packet is split into configurable chunks.
- **Auto Proxy Management**: Automatically toggles macOS `networksetup` on start and graceful exit.
- **Configurable Bypass List**: Define which domains go through the bypass pipeline via `bypass-domains.txt`.

## Usage

Requires Go 1.25+ and macOS. (I'll add windows support later with WinDivert)

```bash
cp .env.example .env  # optional, defaults work fine
make run
```

The system proxy will automatically bind to `localhost:8080` for the `Wi-Fi` interface. Hit `Ctrl+C` to unbind and exit.

## Bypass Domains

Add domains to `bypass-domains.txt`, one per line. Lines starting with `#` are comments.

```
# messaging
discord.com
discord.gg

# other
example.com
```

Set `BYPASS_ALL=true` in your `.env` to bypass everything (sends all traffic through DoH + fragmentation).

## Config

Environment variables:

- `PROXY_ADDR` (default: `127.0.0.1`) listen address
- `PROXY_PORT` (default: `8080`) listen port
- `SYSTEM_SERVICE` (default: `Wi-Fi`) macOS network interface
- `FRAGMENT_SIZE` (default: `7`) TLS ClientHello chunk size in bytes
- `BYPASS_DOMAINS_FILE` (default: `./bypass-domains.txt`) path to the domains list
- `BYPASS_ALL` (default: `false`) bypass all domains
