# zenroute

A minimal, local DNS/TLS proxy written in Go. Bypasses DPI firewalls and DNS poisoning using DoH and TLS ClientHello fragmentation. Supports macOS and Windows.

## Features

- **DoH Resolution**: Uses Cloudflare DoH to bypass poisoned local DNS.
- **TLS Fragmentation**: Splits the initial TLS ClientHello into configurable chunks to bypass DPI.
- **Auto Proxy Management**: Automatically toggles system proxy on start and graceful exit.
- **Configurable Bypass List**: Define which domains go through the bypass pipeline via `bypass-domains.txt`.

## Usage

Requires Go 1.21+ and macOS or Windows.
```bash
cp .env.example .env  # optional, defaults work fine
make run
```

The system proxy will automatically bind to `localhost:8080`. Hit `Ctrl+C` to unbind and exit.

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
- `SYSTEM_SERVICE` (default: `Wi-Fi`) macOS network interface name (macOS only)
- `FRAGMENT_SIZE` (default: `7`) TLS ClientHello chunk size in bytes
- `BYPASS_DOMAINS_FILE` (default: `./bypass-domains.txt`) path to the domains list
- `BYPASS_ALL` (default: `false`) bypass all domains

## Known Limitations

- Some ISPs block certain domains at the IP level rather than via DPI or DNS poisoning. In these cases, TLS fragmentation and DoH cannot help — a relay server outside the restricted region would be required.
- Windows system proxy settings may not apply to all applications. Some apps use their own network stack and ignore OS-level proxy configuration.