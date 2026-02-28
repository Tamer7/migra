# Contributing to Migra

Contributions are welcome. This guide covers how to report bugs, suggest features, and submit code changes.

## Code of Conduct

Be respectful and professional in all interactions.

## Reporting Bugs

Check existing issues first. Include:

- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Migra version and OS
- Relevant configuration (sanitized)

## Suggesting Features

Open an issue with:

- Description of the feature
- Use case explanation
- Examples of how it would work

## Pull Requests

Process:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make test`
6. Commit with clear messages
7. Push and open a PR

## Development Setup

Requirements:

- Go 1.24+
- Make
- Docker (for integration tests)

Setup:

```bash
git clone https://github.com/YOUR_USERNAME/migra.git
cd migra
go mod download
make build
make test
```

## Project Structure

```
migra/
├── cmd/migra/          # CLI entry point
├── internal/
│   ├── adapter/        # Framework adapters
│   ├── cli/            # CLI commands
│   ├── config/         # Configuration
│   ├── engine/         # Execution engines
│   ├── logger/         # Logging
│   ├── state/          # State management
│   └── tenant/         # Multi-tenant support
├── pkg/migra/          # Public API
├── examples/           # Example configurations
└── docs/              # Documentation
```

## Coding Standards

Follow standard Go conventions:

- Run `gofmt` before committing
- Write tests for new features
- Keep functions focused
- Comment exported functions
- Maintain 70%+ test coverage

## Commit Messages

Format:

```
type(scope): subject

body (optional)

footer (optional)
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example:

```
feat(adapter): add PostgreSQL direct adapter

Implements Adapter interface for direct SQL execution.

Closes #123
```

## Adding Framework Adapters

1. Create `internal/adapter/yourframework.go`
2. Implement the `Adapter` interface (Deploy, Rollback, Status)
3. Register in `internal/adapter/registry.go`
4. Add tests in `internal/adapter/yourframework_test.go`
5. Update documentation

## Release Process

Releases are automated via GitHub Actions. To release:

1. Update version
2. Update CHANGELOG.md
3. Create tag: `git tag v0.1.0`
4. Push: `git push origin v0.1.0`

## Questions

- [Discussions](https://github.com/migra/migra/discussions)
- [Issues](https://github.com/migra/migra/issues)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
