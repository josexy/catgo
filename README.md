# catgo

A Cargo-like package manager for Go that provides a simplified, opinionated CLI interface for common Go development tasks.

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/josexy/catgo/releases).

### Build from Source

```bash
# Clone the repository
git clone https://github.com/josexy/catgo.git
cd catgo

# Build for your current platform
make build

# The binary will be available at ./bin/catgo
```

### Cross-platform Builds

```bash
# Build for specific platforms
make linux-amd64
make darwin-arm64
make windows-amd64

# Build all platforms
make all

# Create release packages (.zip files)
make -j releases
```

## Usage

### Creating a New Project

```bash
# Create a new project in a new directory
catgo new myproject

# Create with a specific module name
catgo new myproject --name github.com/username/myproject

# Initialize git repository automatically
catgo new myproject --git
```

### Initializing an Existing Directory

```bash
# Initialize in current directory
catgo init

# Initialize with a specific module name
catgo init --name github.com/username/myproject
```

### Building Your Project

```bash
# Build in dev mode (output to bin/catgo)
catgo build

# Build in release mode (optimized, stripped)
catgo build --release

# Build for a specific target
catgo build --target linux/amd64

# Build specific package
catgo build --package ./cmd/server
catgo build -p cmd/server/main.go
catgo build -p github.com/username/myproject/pkg/examples

# Build to current directory instead of bin/
catgo build --local

# Build to custom binary name instead of module package name
catgo build --output mybinary

# Disable CGO
catgo build --cgo-zero

# Use vendor directory
catgo build --vendor

# Set build variables (ldflags -X)
catgo build --set "main.Version=1.0.0" --set "main.BuildTime=$(date)"
```

### Running Your Project

```bash
# Build and run(same flags to the bulid sub-command)
catgo run

# Pass arguments to the binary
catgo run -- --port 8080 --debug

# Run with build options
catgo run --release -- --config prod.yml
```

### Managing Dependencies

```bash
# Add a single dependency
catgo add github.com/gin-gonic/gin

# Add multiple dependencies
catgo add github.com/spf13/cobra github.com/spf13/viper

# Add a specific version
catgo add github.com/gin-gonic/gin --rev v1.9.0

# Remove dependencies
catgo remove github.com/gin-gonic/gin
```

### Unit testing

```bash
# Test all packages
catgo test

# Test with verbose output
catgo test --verbose

# Test a specific package
catgo test --package ./cmd

# Run specific tests matching a pattern
catgo test --run ^TestBuild

# Run tests multiple times to catch flaky tests
catgo test --count 10

# Enable race detector
catgo test --race

# Set a timeout for tests
catgo test --timeout 30s

# Combine multiple options
catgo test --package ./internal/util --run ^TestExec --verbose --count 3

# Pass arguments to the test binary after --
catgo test -- -custom-flag value
```

### Other Commands

```bash
# Clean build artifacts
catgo clean

# Vendor dependencies
catgo vendor

# Show version information
catgo version
```

## Command Reference

### `catgo new <path>`

Create a new Go project in a new directory.

**Flags:**
- `--name <name>`: Set the module name
- `--git`: Initialize a git repository

### `catgo init [path]`

Initialize a Go module in an existing directory.

**Flags:**
- `--name <name>`: Set the module name
- `--git`: Initialize a git repository

### `catgo build`

Compile the local package to a binary.

**Flags:**
- `-r, --release`: Build in release mode with optimizations
- `-o, --output <name>`: Output binary name
- `-p, --package <path>`: Package to build
- `-t, --target <triple>`: Build for target (e.g., `linux/amd64`)
- `-l, --local`: Build to current directory
- `-z, --cgo-zero`: Disable CGO
- `--vendor`: Use vendor directory
- `-x, --set <var=value>`: Set build variables (ldflags -X)

### `catgo run`

Build and run the local package.

**Flags:** Same as `build`, plus:
- Use `--` to separate catgo flags from program arguments

### `catgo add <package>...`

Add dependencies to the project.

**Flags:**
- `--rev <version>`: Specify version/commit (only with single package)

### `catgo remove <package>...`

Remove dependencies from the project.

### `catgo test`

Run tests for the local package with enhanced output formatting.

**Flags:**
- `-r, --run <pattern>`: Run only tests matching the regular expression (default: `^Test`)
- `-p, --package <path>`: Package to test (default: `./...` for all packages)
- `-c, --count <n>`: Number of times to run each test (default: 1)
- `-t, --timeout <duration>`: Time limit for each test (e.g., `30s`, `5m`)
- `--race`: Enable race detector
- `-v, --verbose`: Show verbose output including test logs
- Use `--` to pass additional arguments to the test binary

**Note:** This command wraps `go test -json` and provides colorized, formatted output with test summaries.

### `catgo clean`

Remove all generated binaries for the local package.

### `catgo vendor`

Vendor dependencies into the vendor directory.

### `catgo version`

Display version information.

## Requirements

- Go 1.25.5 or later

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! This project is a work in progress.
