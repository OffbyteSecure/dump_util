# Contributing to dump_util

Thank you for your interest in contributing to dump_util! We welcome contributions that improve the library, CLI, or documentation.

## How to Contribute

1. **Report Bugs or Request Features**: Open an issue on the [GitHub repository](https://github.com/OffbyteSecure/dump_util). Provide details like steps to reproduce, expected vs. actual behavior, and your environment (Go version, OS, DB types).

2. **Submit Pull Requests**:
   - Fork the repo and create a feature branch (`git checkout -b feature/my-feature`).
   - Make your changes and ensure they pass `go test ./...` and `go build ./...`.
   - Commit with clear messages (e.g., "Fix chunkSlice logic in sql_writer.go").
   - Push to your fork and open a PR against the `main` branch.
   - Reference any related issues in the PR description (e.g., "Fixes #123").

3. **Code Style**:
   - Follow Go conventions: Run `go fmt ./...` and `go vet ./...`.
   - Add tests for new features/bug fixes.
   - Update documentation if needed (e.g., README.md examples).

4. **Development Setup**:
   - Clone: `git clone https://github.com/OffbyteSecure/dump_util.git`
   - Install deps: `go mod tidy`
   - Build CLI: `go build -o dumper ./cmd/dumper`
   - Test: `go test ./...`

We review PRs promptly. For larger changes, discuss in an issue first.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to abide by it.
