# Contributing to Claude Code Switcher

Thanks for your interest in contributing! This is a small, solo-maintained project, so contributions are welcome but may take time to review.

## Getting Started

### Prerequisites

- Go 1.26+ (check `go.mod` for exact version)
- Make (optional, but helpful)
- GoReleaser (optional, for testing releases locally)

### Setup

```bash
# Clone the repo
git clone https://github.com/virtuallytd/claude-code-switcher.git
cd claude-code-switcher

# Install dependencies
go mod download

# Build
make build

# Run tests
make test
```

## Making Changes

### Workflow

1. **Fork the repository**
2. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes**
4. **Add tests** if you're adding functionality
5. **Ensure tests pass**: `go test ./...`
6. **Ensure it builds**: `make build`
7. **Commit with clear messages**:
   ```bash
   git commit -m "feat: add support for X"
   ```
8. **Push to your fork**
9. **Open a Pull Request**

### Commit Message Format

Use conventional commit format (optional but appreciated):
- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation changes
- `test:` — Adding or updating tests
- `refactor:` — Code refactoring
- `chore:` — Build, CI, or tooling changes

### Code Style

- Follow standard Go conventions (`go fmt`, `go vet`)
- Keep functions small and focused
- Add comments for non-obvious logic
- Avoid external dependencies unless necessary

### Testing

- Add unit tests for new functionality
- Test files should be named `*_test.go`
- Run tests before submitting: `go test -v -race ./...`

## Pull Request Guidelines

- **Keep PRs focused** — One feature or fix per PR
- **Write clear descriptions** — Explain what and why, not just how
- **Link to issues** — Reference any related issues
- **Be patient** — This is a side project, reviews may take a few days

## Areas for Contribution

Some ideas if you're looking for where to help:

- **Windows support** — Add keyring backend for Windows
- **Tests** — We need more test coverage!
- **Documentation** — Improve docs, add examples
- **Bug fixes** — Check the issues page
- **MCP server templates** — Example configs for common MCP servers

## Questions?

Open an issue or email tony@virtuallytd.com.

## Code of Conduct

Be respectful, constructive, and professional. This is a small community project — let's keep it welcoming.
