# Contributing to Nightscout LibreLink Up Go

First off, thank you for considering contributing to this project! This application helps the Type 1 Diabetes and Nightscout community, and your contributions make a real difference.

## Code of Conduct

Be respectful, inclusive, and considerate. This project exists to help people manage their diabetes better.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Environment details**:
  - Docker version
  - LibreLink region
  - Nightscout version
  - Error logs (sanitize any personal data!)

**Example**:
```markdown
**Bug**: Application crashes on startup

**Environment**:
- Docker: 24.0.7
- Region: EU
- Nightscout: 15.0.2

**Steps to Reproduce**:
1. Start container with docker compose
2. Check logs

**Logs**:
```
Error: connection refused
```

**Expected**: Should connect to Nightscout

**Actual**: Crashes immediately
```

### Suggesting Enhancements

Enhancement suggestions are welcome! Please include:

- **Use case**: Why is this useful?
- **Proposed solution**: How should it work?
- **Alternatives considered**: What other approaches did you think about?

### Pull Requests

1. **Fork the repository**
2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**:
   - Write clear, idiomatic Go code
   - Follow the existing code style
   - Add comments for complex logic
   - Keep commits atomic and well-described

4. **Test your changes**:
   ```bash
   # Build
   go build -o nightscout-librelink-up-go .

   # Test
   go test ./...

   # Run locally
   ./nightscout-librelink-up-go
   ```

5. **Build Docker image**:
   ```bash
   docker build -t nightscout-librelink-up-go:test .
   docker run nightscout-librelink-up-go:test
   ```

6. **Commit your changes**:
   ```bash
   git commit -m "Add feature: brief description"
   ```

   Good commit message example:
   ```
   Add support for multiple Nightscout instances

   - Allow configuring multiple NIGHTSCOUT_URL values
   - Post readings to all configured instances
   - Add error handling for partial failures
   ```

7. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

8. **Open a Pull Request**:
   - Link to any related issues
   - Describe what changed and why
   - Include testing steps if applicable

## Development Setup

### Prerequisites

- **Go 1.23+**: [Download](https://go.dev/dl/)
- **Docker** (optional): For testing containerized builds
- **LibreLink Up account**: For testing (or use mock data)

### Local Development

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/nightscout-librelink-up-go.git
cd nightscout-librelink-up-go

# Install dependencies
go mod download

# Run with live reload (install air first)
go install github.com/cosmtrek/air@latest
air

# Or run directly
go run main.go
```

### Project Structure

```
.
├── main.go              # Entry point
├── config/
│   └── config.go       # Environment variable parsing
├── librelink/
│   └── client.go       # LibreLink Up API client
├── nightscout/
│   └── client.go       # Nightscout API client
├── Dockerfile          # Multi-stage build
└── go.mod              # Dependencies
```

### Code Style

- Follow standard Go formatting: `gofmt` or `goimports`
- Use meaningful variable names
- Keep functions small and focused
- Add comments for exported functions
- Error handling: Always check and handle errors appropriately

**Example**:
```go
// Good
func fetchGlucoseData(ctx context.Context, client *librelink.Client) (*GlucoseReading, error) {
    data, err := client.GetLatestReading(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch glucose data: %w", err)
    }
    return data, nil
}

// Avoid
func fetch(c *librelink.Client) *GlucoseReading {
    d, _ := c.GetLatestReading(context.Background()) // Never ignore errors!
    return d
}
```

### Testing

- Add tests for new functionality
- Run tests before submitting PR: `go test ./...`
- Test with different regions if applicable
- Verify Docker build works: `docker build .`

### Documentation

- Update README.md if adding features
- Add inline comments for complex logic
- Update environment variable table if adding new config

## What We're Looking For

### Priority Areas

- **Bug fixes**: Especially authentication or data sync issues
- **Additional regions**: Support for more LibreLink regions
- **Error handling**: Better retry logic, exponential backoff
- **Logging improvements**: More detailed debug information
- **Performance**: Reduce memory/CPU usage further
- **Tests**: Unit tests, integration tests
- **Documentation**: Clearer setup guides, troubleshooting

### Nice to Have

- Support for multiple CGM followers
- Configurable retry strategies
- Health check endpoint
- Prometheus metrics
- Configuration via file (in addition to env vars)

## Questions?

- **Open an issue**: For questions about contributing
- **Discussions**: For general questions about the project
- **Discord**: Join the [Nightscout Discord](https://discord.gg/zg7CvCQ)

## Recognition

Contributors will be recognized in the README and release notes. Thank you for helping make diabetes management better!

---

**Note**: This project is maintained in a private homelab automation repository, then synced to this public repo. If you're creating a PR, it will be reviewed and merged here, then synced back to the private repo.
