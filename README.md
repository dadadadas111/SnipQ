# SnipQ - Universal Snippet Expander

[![Core CI/CD](https://github.com/dadadadas111/SnipQ/actions/workflows/core.yml/badge.svg)](https://github.com/dadadadas111/SnipQ/actions/workflows/core.yml)
[![codecov](https://codecov.io/gh/dadadadas111/SnipQ/branch/main/graph/badge.svg)](https://codecov.io/gh/dadadadas111/SnipQ)
[![Go Report Card](https://goreportcard.com/badge/github.com/dadadadas111/SnipQ)](https://goreportcard.com/report/github.com/dadadadas111/SnipQ)

A cross-platform text snippet expander that works like IDE live templates, but system-wide on Windows, browsers, and mobile.

## ✨ Features

- **Query-style options**: `:date?format=YYYY-MM-DD`, `:ty?lang=vi&tone=casual`
- **Dynamic generators**: date, UUID, counters, clipboard, random
- **File-based vault**: portable, versionable, and syncable
- **Offline-first**: cloud sync for vault files and built-in packs
- **Cross-platform**: Windows, browsers, Android, iOS

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/dadadadas111/SnipQ.git
cd SnipQ/core

# Build the CLI tool
go build -o snipq ./cmd/cli

# Initialize a sample vault
./snipq init

# Try some expansions
./snipq expand ":ty"                          # → "Thank you."
./snipq expand ":ty?lang=vi&tone=casual"      # → "Cảm ơn bạn nha!"
./snipq expand ":date?format=Mon, 02 Jan 2006" # → "Thu, 28 Aug 2025"
./snipq expand ":uuid?upper=true"             # → "A977B672-8C53-4990-95B7-E5A2012BBD71"
```

### Example Snippets

**Multi-language Thanks**
```yaml
id: "snp_ty"
trigger: ":ty"
defaults:
  lang: "en"
  tone: "neutral"
template: |
  {{ if eq .lang "vi" }}{{ if eq .tone "casual" }}Cảm ơn bạn nha!{{ else }}Cảm ơn bạn.{{ end }}
  {{ else if eq .lang "ja" }}ありがとうございます。
  {{ else }}{{ if eq .tone "casual" }}Thanks!{{ else }}Thank you.{{ end }}{{ end }}
```

**Dynamic Date**
```yaml
id: "snp_date"
trigger: ":date"
defaults:
  format: "2006-01-02"
  tz: "Local"
template: "{{ date .format .tz }}"
```

## 🏗 Architecture

```
/core/           - Go core library (vault, parser, templates, API)
/sync-api/       - NestJS API for authentication and sync
/apps/
  /windows/      - Windows desktop app
  /browser-extension/ - Chrome/Firefox extension
  /mobile/       - Android/iOS apps
```

## 📦 Development Status

### ✅ Phase 1 - Core (Go) - **COMPLETE**
- [x] File-based vault (groups, snippets, settings, counters, history)
- [x] Query parser + precedence logic
- [x] Template rendering with functions
- [x] Expand/Preview API
- [x] CLI tool for testing

### 🔄 Phase 2 - Windows App MVP - **IN PROGRESS**
- [ ] Tray app + editor for vault
- [ ] Global keyboard hook + expansion
- [ ] Per-app exclusions + pause toggle

### 📋 Phase 3 - API + Sync
- [ ] Auth (Firebase + JWT exchange)
- [ ] Vault index/blobs endpoints
- [ ] Device registration
- [ ] Built-in packs listing & install
- [ ] Sync client in Windows app

### 📋 Phase 4 - Browser Extension
- [ ] Content script for textareas
- [ ] WASM core
- [ ] Chrome/Edge/Firefox support

### 📋 Phase 5 - Mobile Apps
- [ ] Android IME (gomobile)
- [ ] iOS Keyboard Extension (gomobile)
- [ ] Offline vault + sync

## 🛠 Development

### Prerequisites

- Go 1.22+
- Node.js 18+ (for sync-api)
- Git

### Building

```bash
# Core library and CLI
cd core
go build ./cmd/cli

# Run tests
go test ./...

# Run linter
golangci-lint run

# API (when ready)
cd sync-api
npm install
npm run build
```

### Project Structure

```
core/
├── pkg/
│   ├── types/        # Shared types (Group, Snippet, Settings, etc.)
│   ├── parser/       # Query parameter parsing
│   ├── template/     # Template engine with built-in functions
│   ├── vault/        # File-based storage management
│   └── core/         # Main engine implementation
├── cmd/cli/          # CLI tool for testing
└── internal/testdata/ # Sample vault for testing
```

### Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Add tests if applicable
5. Run tests: `go test ./...`
6. Run linter: `golangci-lint run`
7. Commit: `git commit -m 'Add amazing feature'`
8. Push: `git push origin feature/amazing-feature`
9. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Support

- 📧 Email: support@snipq.dev
- 🐛 Issues: [GitHub Issues](https://github.com/dadadadas111/SnipQ/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/dadadadas111/SnipQ/discussions)
