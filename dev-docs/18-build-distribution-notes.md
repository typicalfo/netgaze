# Addendum 18 – Build & Distribution Notes

**Build Strategy**: Simple GitHub distribution with go install support.

**Build Requirements**
- Go 1.21+ (for modern stdlib features)
- No CGO dependency for cross-platform static binaries
- Single binary distribution
- Version information embedded at build time

**Build Commands**
```bash
# Development build
make build

# Production build with version info
go build -ldflags "-X main.version=v1.0.0 -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o ng .

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=v1.0.0" -o ng-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=v1.0.0" -o ng-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=v1.0.0" -o ng-windows-amd64.exe .
```

**Version Information**
```go
// main.go
package main

var (
    version   = "dev"
    buildDate = "unknown"
)

func init() {
    cmd.Version = version
    cmd.BuildDate = buildDate
}
```

**GitHub Distribution**
- Public repository at github.com/username/netgaze
- Releases with pre-built binaries for major platforms
- Source distribution via go install

**User Installation Methods**

**Method 1: go install (recommended)**
```bash
go install github.com/username/netgaze@latest
```

**Method 2: Download binary from releases**
```bash
# Linux
curl -L https://github.com/username/netgaze/releases/latest/download/ng-linux-amd64 -o ng
chmod +x ng

# macOS
curl -L https://github.com/username/netgaze/releases/latest/download/ng-darwin-amd64 -o ng
chmod +x ng

# Windows
curl -L https://github.com/username/netgaze/releases/latest/download/ng-windows-amd64.exe -o ng.exe
```

**Method 3: Build from source**
```bash
git clone https://github.com/username/netgaze.git
cd netgaze
make build
```

**Dependencies Management**
- Use Go modules (go.mod already exists)
- Pin dependency versions for reproducible builds
- Regular dependency updates for security
- Minimal external dependencies (only essential packages)

**Static Binary Considerations**
- Avoid CGO where possible for broader compatibility
- Use pure Go implementations for networking
- Test builds on target platforms
- Consider upx compression for smaller binaries (optional)

**Release Process**
1. Update version in go.mod tags
2. Run full test suite
3. Build cross-platform binaries
4. Create GitHub release with binaries
5. Update go.mod version tag

**Future Distribution Options**
- Homebrew formula (user community contribution)
- Docker image for containerized usage
- Package manager repos (apt, yum, etc.)
- Snap or Flatpak for Linux distributions

**No Complex Build System**
- Makefile not required for basic usage
- Simple shell script for cross-platform builds
- GitHub Actions for automated releases (optional)
- No external build tools or dependencies

**Binary Size Optimization**
- Use upx for compression (optional, ~50% reduction)
- Strip debug symbols in production builds
- Optimize Go build flags for size
- Monitor binary size as dependencies grow

**Security Considerations**
- Sign releases with GPG keys (optional)
- Provide checksums for binary verification
- Regular security audits of dependencies
- Reproducible builds for verification