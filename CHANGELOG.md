# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release with core snippet expansion functionality
- Support for query-style options in triggers (`:trigger?param=value`)
- File-based vault system for storing snippets and settings
- Template engine with built-in functions (date, uuid, random, etc.)
- CLI tool for testing and managing snippets
- Cross-platform support (Windows, Linux, macOS)

### Features
- **Query Parser**: Parse triggers like `:ty?lang=vi&tone=casual`
- **Template Engine**: Go templates with custom functions
- **Vault Management**: YAML-based snippet storage
- **Built-in Functions**:
  - `date` - Format current date/time
  - `uuid` - Generate UUIDs
  - `random` - Generate random values
  - `counter` - Increment counters
  - String manipulation (upper, lower, etc.)
- **CLI Commands**:
  - `expand` - Expand a trigger
  - `preview` - Preview expansion without side effects
  - `list` - List all available snippets
  - `init` - Initialize a new vault

### Development
- Comprehensive test suite
- GitHub Actions CI/CD pipeline
- Code quality checks with golangci-lint
- Security scanning with Gosec
- Multi-platform builds and releases

## [0.1.0] - 2025-08-28

### Added
- Initial project structure
- Core Go library implementation
- Basic CLI tool for testing

### Technical Details
- Go 1.22+ support
- YAML configuration files
- Template-based snippet expansion
- Query parameter parsing
- File-based storage system
