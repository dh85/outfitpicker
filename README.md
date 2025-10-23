# outfitpicker

Interactive CLI to pick outfits from category folders with per-category rotation, cached in your synced root.

## Installation

### Quick Install (Recommended)
```bash
curl -fsSL https://raw.githubusercontent.com/dh85/outfitpicker/main/install.sh | bash
```

### Package Managers

#### Homebrew (macOS/Linux)
```bash
brew tap dh85/tap
brew install outfitpicker
```

#### Go Install
```bash
go install github.com/dh85/outfitpicker/cmd/outfitpicker@latest
```

#### Windows

**PowerShell:**
```powershell
Invoke-WebRequest -Uri "https://github.com/dh85/outfitpicker/releases/latest/download/outfitpicker_Windows_x86_64.zip" -OutFile "outfitpicker.zip"
Expand-Archive -Path "outfitpicker.zip" -DestinationPath "."

# Add to PATH (choose one method):
# Method 1: Move to existing PATH directory
Move-Item outfitpicker.exe "$env:USERPROFILE\AppData\Local\Microsoft\WindowsApps\"

# Method 2: Add current directory to PATH
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
[Environment]::SetEnvironmentVariable("PATH", "$currentPath;$(Get-Location)", "User")
```

#### Linux Packages

**Debian/Ubuntu:**
```bash
wget https://github.com/dh85/outfitpicker/releases/latest/download/outfitpicker_linux_amd64.deb
sudo dpkg -i outfitpicker_linux_amd64.deb
```

**RHEL/CentOS/Fedora:**
```bash
wget https://github.com/dh85/outfitpicker/releases/latest/download/outfitpicker_linux_amd64.rpm
sudo rpm -i outfitpicker_linux_amd64.rpm
```

**Alpine Linux:**
```bash
wget https://github.com/dh85/outfitpicker/releases/latest/download/outfitpicker_linux_amd64.apk
sudo apk add --allow-untrusted outfitpicker_linux_amd64.apk
```

**Arch Linux:**
```bash
wget https://github.com/dh85/outfitpicker/releases/latest/download/outfitpicker_linux_amd64.pkg.tar.xz
sudo pacman -U outfitpicker_linux_amd64.pkg.tar.xz
```

### Manual Download
Download the latest binary for your platform from the [releases page](https://github.com/dh85/outfitpicker/releases).

## Updating

### Homebrew
```bash
brew update
brew upgrade outfitpicker
```

### Go Install
```bash
go install github.com/dh85/outfitpicker/cmd/outfitpicker@latest
```

### Linux Packages
Reinstall using the same method as installation with the latest version.

### Quick Install Script
```bash
curl -fsSL https://raw.githubusercontent.com/dh85/outfitpicker/main/install.sh | bash
```

## Usage

### New to Outfit Picker?

📖 **[Complete User Guide](USER_GUIDE.md)** - Detailed walkthrough for beginners  
⚡ **[Quick Start Guide](QUICK_START.md)** - Get up and running in 3 steps

### First run

```bash
outfitpicker
```

## Development

### Requirements

- Go 1.24.9 or later
- Git
- Make (optional, for convenience commands)
- golangci-lint (for linting)
- GoReleaser (for releases)

### Setup

#### Using Dev Container (Recommended)
The project includes a dev container configuration for consistent development environment:

1. Open in VS Code with Dev Containers extension
2. Click "Reopen in Container" when prompted
3. All tools (Go, golangci-lint, GoReleaser) are automatically installed

#### Manual Setup
```bash
# Clone the repository
git clone https://github.com/dh85/outfitpicker.git
cd outfitpicker

# Install dependencies
go mod tidy
go mod download

# Build the project
make build
# or
go build -o bin/outfitpicker ./cmd/outfitpicker
go build -o bin/outfitpicker-admin ./cmd/outfitpicker-admin
```

### Architecture

```
outfitpicker/
├── cmd/                    # CLI entry points
├── internal/               # Private application code
│   ├── app/               # Core business logic
│   ├── cli/               # CLI interface
│   ├── storage/           # Cache management
│   └── ui/                # User interface components
├── pkg/                   # Public packages
│   ├── config/           # Configuration management
│   └── version/          # Version information
└── test/                 # Integration tests
```

### Testing

```bash
# Run all tests
make test
# or
go test ./...

# Run tests with race detection
make test-race

# Run short tests only
make test-short

# Generate coverage report
make test-coverage

# Run benchmarks
make bench

# Run fuzz tests
make fuzz
```

### Code Quality

```bash
# Run linter
make lint
# or
golangci-lint run

# Format code
go fmt ./...
```

### Making a Release

1. **Prepare the release:**
   ```bash
   # Test release locally
   make release-test
   
   # Check release configuration
   make release-check
   ```

2. **Create and push a tag:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

3. **GitHub Actions will automatically:**
   - Run tests
   - Build binaries for all platforms
   - Create GitHub release with assets
   - Update Homebrew tap
   - Generate Linux packages (.deb, .rpm, .apk, .pkg.tar.xz)

### Development Workflow

1. Create a feature branch
2. Make changes and add tests
3. Run `make test lint` to ensure quality
4. Submit a pull request
5. After merge, tag for release if needed

### Project Structure

- **cmd/**: CLI applications (outfitpicker, outfitpicker-admin)
- **internal/app/**: Core business logic and algorithms
- **internal/storage/**: Cache management and persistence
- **internal/cli/**: Command-line interface and user interaction
- **pkg/config/**: Configuration file handling
- **pkg/version/**: Version information and build metadata