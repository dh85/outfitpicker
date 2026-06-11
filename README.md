# OutfitPicker - Go CLI

OutfitPicker is a terminal app for choosing outfits from a local wardrobe directory.
Each subdirectory under the configured wardrobe root is treated as a category, and
`.avatar` files inside those category directories are treated as outfits.

The app can:

- scan wardrobe categories and report whether they contain outfits
- pick a random outfit within a category or across all available categories
- avoid repeats during the current interactive session
- persist worn outfit rotation state between runs
- reset one category or all category rotations
- exclude categories from cross-category random selection
- recover from missing or invalid config during startup

## Project Structure

```text
outfitpicker-go/
├── cmd/outfitpicker/              # Production composition root and main()
├── internal/
│   ├── application/usecases/      # Application use cases and query services
│   ├── cli/                       # Console UI, menu flow, runtime/session orchestration
│   ├── domain/
│   │   ├── entities/              # Core data structures
│   │   ├── errors/                # Domain/application error mapping
│   │   ├── interfaces/            # Repository/service ports
│   │   ├── logic/                 # Business rules
│   │   └── validation/            # Input/path/language validation
│   └── infrastructure/
│       ├── persistence/           # Config/cache repositories
│       ├── services/              # Filesystem-backed category scanner
│       └── system/                # File services and OS adapters
├── demo-outfits*/                 # Sample wardrobe data
└── docs/                          # Architecture and migration notes
```

## Architecture

The app follows a layered design:

- `cmd/outfitpicker` wires production dependencies.
- `internal/domain` owns business concepts and validation.
- `internal/application/usecases` owns deterministic use cases and query services.
- `internal/infrastructure` owns filesystem and persistence adapters.
- `internal/cli` owns terminal interaction, menu transitions, session-only state, and random selection.

Important current decisions:

- `Application` is a thin CLI facade; it does not expose mutable config state.
- Config is accessed through `ConfigurationController`.
- Random outfit choice is centralized in `RuntimeSelectionService`.
- `PickOutfitUseCase` only loads candidate outfits and does not choose randomly.
- Config/cache writes use atomic temp-file-and-rename persistence with per-path in-process and PID-aware lock-file serialization.

## Development

```bash
make test
make coverage-check
make build
make run
```

Coverage is enforced by `make coverage-check` with `COVERAGE_MIN ?= 92.5`.

Current verified coverage is above the threshold, with domain and infrastructure packages near or at full coverage.

## Runtime Data

Config and cache are stored under the user config directory in an `outfitpicker`
subdirectory. `XDG_CONFIG_HOME` is honored when set.

The persisted files are:

- `config.json`
- `cache.json`

## Notes

This is a local CLI app. It includes atomic file replacement plus write locking for
normal single-user usage, concurrent goroutines, and overlapping app instances.
Path validation rejects traversal, restricted system directories, control
characters, and symlink components while allowing normal Unicode user paths. It is
not intended as a multi-user service or daemon.
