# Contributing to SnipQ

First off, thank you for considering contributing to SnipQ! It's people like you that make SnipQ such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the issue list as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps which reproduce the problem**
- **Provide specific examples to demonstrate the steps**
- **Describe the behavior you observed after following the steps**
- **Explain which behavior you expected to see instead and why**
- **Include details about your configuration and environment**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

- **Use a clear and descriptive title**
- **Provide a step-by-step description of the suggested enhancement**
- **Provide specific examples to demonstrate the steps**
- **Describe the current behavior and explain which behavior you expected to see instead**
- **Explain why this enhancement would be useful**

### Pull Requests

1. **Fork the repo** and create your branch from `main`
2. **Make your changes** in your fork
3. **Add tests** if you've added code that should be tested
4. **Update documentation** if you've changed APIs
5. **Ensure the test suite passes**
6. **Make sure your code lints**
7. **Issue that pull request!**

## Development Process

### Setting Up Development Environment

```bash
# Clone your fork
git clone https://github.com/dadadadas111/SnipQ.git
cd SnipQ

# Set up Go environment
cd core
go mod download
go mod verify

# Run tests
go test ./...

# Run linter
golangci-lint run
```

### Core Development Guidelines

#### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `goimports`
- Write meaningful comments for exported functions
- Keep functions small and focused
- Use meaningful variable and function names

#### Testing

- Write tests for new functionality
- Maintain or improve test coverage
- Use table-driven tests where appropriate
- Test both success and error cases

#### Documentation

- Document all exported functions and types
- Keep README.md up to date
- Add examples for new features
- Update CHANGELOG.md for notable changes

### Git Workflow

#### Branch Naming

- `feature/description` - for new features
- `bugfix/description` - for bug fixes
- `docs/description` - for documentation updates
- `refactor/description` - for code refactoring

#### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
type(scope): description

feat(core): add support for nested snippet groups
fix(parser): handle edge case in query parameter parsing
docs(readme): update installation instructions
test(vault): add tests for snippet CRUD operations
```

Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

### Project Structure

```
SnipQ/
├── core/                    # Go core library
│   ├── pkg/
│   │   ├── types/          # Shared types
│   │   ├── parser/         # Query parsing
│   │   ├── template/       # Template engine
│   │   ├── vault/          # Storage management
│   │   └── core/           # Main engine
│   ├── cmd/cli/            # CLI application
│   └── internal/           # Internal packages
├── sync-api/               # NestJS sync API
├── apps/                   # Client applications
│   ├── windows/           # Windows desktop app
│   ├── browser-extension/ # Browser extension
│   └── mobile/            # Mobile apps
└── docs/                  # Documentation
```

### Component-Specific Guidelines

#### Core Library (`/core`)

- All public APIs must be documented
- Breaking changes require major version bump
- Must maintain backward compatibility within major versions
- Performance-critical code should be benchmarked

#### CLI Tool (`/core/cmd/cli`)

- Follow Unix command-line conventions
- Provide helpful error messages
- Support `--help` and `--version` flags
- Use consistent flag naming

#### Sync API (`/sync-api`)

- Follow RESTful API conventions
- Document all endpoints with OpenAPI/Swagger
- Include proper error handling and status codes
- Validate all inputs

### Review Process

1. **Automated Checks**: All PRs must pass CI checks (tests, linting, security scans)
2. **Code Review**: At least one maintainer must review and approve
3. **Testing**: New features must include appropriate tests
4. **Documentation**: User-facing changes must include documentation updates

### Release Process

1. Update version numbers
2. Update CHANGELOG.md
3. Create release PR
4. Merge to main after approval
5. Tag release
6. GitHub Actions will build and publish artifacts

## Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Email**: For security issues or private concerns

## Recognition

Contributors are recognized in:
- README.md contributor section
- Release notes
- Annual contributor appreciation posts

Thank you for contributing to SnipQ! 🚀
