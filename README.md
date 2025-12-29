# Nightscout LibreLink Up Go

[![Docker Image Size](https://img.shields.io/docker/image-size/ghcr.io/mrcodeeu/nightscout-librelink-up-go?label=image%20size)](https://github.com/mrcodeeu/nightscout-librelink-up-go/pkgs/container/nightscout-librelink-up-go)
[![License](https://img.shields.io/github/license/mrcodeeu/nightscout-librelink-up-go)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mrcodeeu/nightscout-librelink-up-go)](go.mod)

A lightweight, high-performance Go application that fetches blood glucose data from Abbott's LibreLink Up service and automatically posts it to your Nightscout instance.

## Why This Implementation?

This Go implementation provides significant advantages over the original Node.js version:

| Feature | This (Go) | Node.js Alternative |
|---------|-----------|---------------------|
| **Image Size** | ~10 MB | ~500 MB |
| **Memory Usage** | 5-10 MB | 50-100 MB |
| **Startup Time** | < 1 second | 3-5 seconds |
| **CPU Usage** | Minimal | Higher |
| **Dependencies** | None (single binary) | npm + modules |

## Features

- 🔒 **Secure Authentication** - Direct integration with LibreLink Up API
- 🩸 **Automatic Polling** - Configurable intervals (default: 5 minutes)
- 📊 **Trend Arrows** - Posts glucose readings with trend indicators
- 🐳 **Minimal Docker Image** - Built on Alpine Linux (~10MB total)
- ⚡ **High Performance** - Native Go binary, extremely fast
- 🔄 **Graceful Shutdown** - Proper signal handling
- 📝 **Comprehensive Logging** - Detailed logs for troubleshooting
- 🌍 **Multi-Region Support** - Works with all LibreLink Up regions

## Quick Start

### Using Docker Compose (Recommended)

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  mongo:
    image: mongo:4.4
    restart: unless-stopped
    volumes:
      - mongo-data:/data/db
    networks:
      - nightscout

  nightscout:
    image: nightscout/cgm-remote-monitor:latest
    restart: unless-stopped
    depends_on:
      - mongo
    environment:
      MONGODB_URI: mongodb://mongo:27017/nightscout
      API_SECRET: your-secret-here-change-this
      BASE_URL: https://nightscout.yourdomain.com
      # Enable additional plugins as needed
      ENABLE: careportal basal
    ports:
      - "1337:1337"
    networks:
      - nightscout

  librelink-up:
    image: ghcr.io/mrcodeeu/nightscout-librelink-up-go:latest
    restart: unless-stopped
    depends_on:
      - nightscout
    environment:
      LINK_UP_USERNAME: your@email.com
      LINK_UP_PASSWORD: yourpassword
      LINK_UP_REGION: EU
      LINK_UP_TIME_INTERVAL: 5
      NIGHTSCOUT_URL: http://nightscout:1337
      NIGHTSCOUT_API_TOKEN: your-api-secret-here-change-this
    networks:
      - nightscout

networks:
  nightscout:
    driver: bridge

volumes:
  mongo-data:
```

Then start the stack:

```bash
docker compose up -d
```

### Using Docker Run

```bash
docker run -d \
  --name nightscout-librelink-up \
  --restart unless-stopped \
  -e LINK_UP_USERNAME="your@email.com" \
  -e LINK_UP_PASSWORD="yourpassword" \
  -e LINK_UP_REGION="EU" \
  -e LINK_UP_TIME_INTERVAL="5" \
  -e NIGHTSCOUT_URL="http://nightscout:1337" \
  -e NIGHTSCOUT_API_TOKEN="your-api-token" \
  ghcr.io/mrcodeeu/nightscout-librelink-up-go:latest
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `LINK_UP_USERNAME` | Your LibreLink Up email address | - | ✅ |
| `LINK_UP_PASSWORD` | Your LibreLink Up password | - | ✅ |
| `LINK_UP_REGION` | Your LibreLink Up region (see below) | `EU` | ❌ |
| `LINK_UP_TIME_INTERVAL` | Polling interval in minutes | `5` | ❌ |
| `NIGHTSCOUT_URL` | Your Nightscout URL (e.g., `http://nightscout:1337`) | - | ✅ |
| `NIGHTSCOUT_API_TOKEN` | Your Nightscout API secret (same as API_SECRET) | - | ✅ |

### Supported Regions

- `AE` - United Arab Emirates
- `AP` - Asia-Pacific
- `AU` - Australia
- `CA` - Canada
- `DE` - Germany
- `EU` - Europe (most European countries)
- `EU2` - Europe 2
- `FR` - France
- `JP` - Japan
- `US` - United States
- `LA` - Latin America
- `RU` - Russia
- `CN` - China

**Note**: Use the region where you registered your LibreLink Up account.

## Getting Your Nightscout API Token

The `NIGHTSCOUT_API_TOKEN` should be the **SHA1 hash** of your Nightscout `API_SECRET`.

You can generate it using:

```bash
echo -n "your-api-secret" | sha1sum | cut -d' ' -f1
```

Or use the same value as your `API_SECRET` in the Nightscout configuration.

## Architecture

```
┌─────────────────┐
│ LibreLink Up    │
│ (Abbott Cloud)  │
└────────┬────────┘
         │
         │ HTTPS (Poll every N minutes)
         │
┌────────▼────────┐
│  This App       │
│  (Go Binary)    │
│                 │
│ • Fetch data    │
│ • Transform     │
│ • Post to NS    │
└────────┬────────┘
         │
         │ HTTP POST
         │
┌────────▼────────┐
│  Nightscout     │
│  (Your CGM DB)  │
└─────────────────┘
```

## Building from Source

### Prerequisites

- Go 1.23 or later
- Docker (optional, for containerized builds)

### Local Build

```bash
git clone https://github.com/mrcodeeu/nightscout-librelink-up-go.git
cd nightscout-librelink-up-go
go build -o nightscout-librelink-up-go .
```

### Docker Build

```bash
docker build -t nightscout-librelink-up-go:local .
```

## Development

### Project Structure

```
.
├── main.go              # Application entry point
├── config/
│   └── config.go       # Configuration handling (env vars)
├── librelink/
│   └── client.go       # LibreLink Up API client
├── nightscout/
│   └── client.go       # Nightscout API client
├── Dockerfile          # Multi-stage Docker build
└── go.mod              # Go module definition
```

### Running Locally

```bash
# Set environment variables
export LINK_UP_USERNAME="your@email.com"
export LINK_UP_PASSWORD="yourpassword"
export LINK_UP_REGION="EU"
export NIGHTSCOUT_URL="http://localhost:1337"
export NIGHTSCOUT_API_TOKEN="your-token"

# Run
go run main.go
```

### Testing

```bash
go test ./...
```

## Troubleshooting

### Container keeps restarting

Check the logs:
```bash
docker logs nightscout-librelink-up
```

Common issues:
- **Authentication failed**: Check your LibreLink Up credentials
- **Wrong region**: Verify the region matches your LibreLink account
- **Nightscout unreachable**: Ensure the URL is correct (use container name if using Docker network)

### Data not appearing in Nightscout

1. **Verify API token**: Ensure `NIGHTSCOUT_API_TOKEN` matches your Nightscout `API_SECRET`
2. **Check Nightscout logs**: Look for incoming POST requests
3. **Test connectivity**: Try `docker exec nightscout-librelink-up ping nightscout`

### "No connection" or network errors

- If using Docker Compose, ensure both containers are on the same network
- Use container names for internal communication (e.g., `http://nightscout:1337`)
- For external Nightscout, use the full URL with `http://` or `https://`

### High CPU usage

This app uses minimal CPU. If you see high usage:
- Check the polling interval (`LINK_UP_TIME_INTERVAL`) - 5 minutes is recommended
- Verify no infinite retry loops in logs

## Comparison with Other Solutions

| Feature | This Project | timoschlueter/nightscout-librelink-up |
|---------|--------------|---------------------------------------|
| Language | Go | Node.js |
| Image Size | ~10 MB | ~500 MB |
| Memory | ~5-10 MB | ~50-100 MB |
| Startup Time | < 1 second | 3-5 seconds |
| Dependencies | 0 (static binary) | npm + many packages |
| Performance | Very high | Good |
| Build Time | Fast | Moderate |

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### How to Contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## FAQ

**Q: Does this work with FreeStyle Libre 2/3?**
A: Yes! As long as you use LibreLink Up to share your data.

**Q: Can I use this without Nightscout?**
A: No, this specifically syncs data to Nightscout. For other destinations, you'd need to modify the code.

**Q: How often does it poll for new data?**
A: Default is every 5 minutes. Configure via `LINK_UP_TIME_INTERVAL`.

**Q: Is my LibreLink password stored securely?**
A: Yes, it's only stored in environment variables and never logged or persisted.

**Q: Can I run multiple instances?**
A: Yes, but be aware of LibreLink API rate limits. One instance per user is recommended.

## Credits

- Inspired by [timoschlueter/nightscout-librelink-up](https://github.com/timoschlueter/nightscout-librelink-up)
- Built for the [Nightscout](https://nightscout.github.io/) community
- Uses Abbott's LibreLink Up service

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

This project is not affiliated with, endorsed by, or supported by Abbott or Nightscout Foundation. Use at your own risk. Always verify glucose readings with your actual device before making treatment decisions.

## Support

- **Issues**: [GitHub Issues](https://github.com/mrcodeeu/nightscout-librelink-up-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mrcodeeu/nightscout-librelink-up-go/discussions)
- **Nightscout Community**: [Nightscout Discord](https://discord.gg/zg7CvCQ)

---

Made with ❤️ for the Nightscout and T1D community
