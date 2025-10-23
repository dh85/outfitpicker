# Architecture Overview

## Project Structure

```
outfitpicker/
├── cmd/                    # CLI entry points
├── internal/               # Private application code
│   ├── app/               # Core business logic
│   ├── cli/               # CLI interface
│   ├── storage/           # Cache management
│   ├── interfaces/        # Abstractions
│   ├── mocks/            # Test mocks
│   ├── metrics/          # Performance monitoring
│   └── testutil/         # Shared test utilities
├── pkg/                   # Public packages
│   ├── config/           # Configuration management
│   └── version/          # Version information
└── test/                 # Integration tests
```

## Key Components

### Core Business Logic (`internal/app`)
- Category management and file operations
- Random selection algorithms
- User interaction flows

### Storage Layer (`internal/storage`)
- JSON-based cache persistence
- Thread-safe operations
- File system abstraction

### CLI Interface (`internal/cli`)
- First-run wizard
- User input handling
- Cross-platform compatibility

## Design Principles

1. **Separation of Concerns**: Clear boundaries between layers
2. **Dependency Injection**: Interfaces for testability
3. **Error Handling**: Comprehensive error propagation
4. **Thread Safety**: Concurrent operation support
5. **Cross-Platform**: Windows/Unix compatibility