
# 🚀 Project Brief — Universal Snippet Expander with Query-Style Options - SNIPQ

## 1. Vision

A **cross-platform text snippet expander** that works like IDE live templates (e.g., `sout → System.out.println()`), but **system-wide** on Windows, browsers, and mobile.

* Users define **short triggers** (e.g., `:addr`, `:ty`) that expand into longer text.
* Supports **dynamic generators** (date, UUID, counters, clipboard, random).
* Special twist: **query-style options** (`:date?format=YYYY-MM-DD`, `:ty?lang=vi&tone=casual`) let users override defaults on the fly.
* Snippets are stored in **files/folders** (groups), making them portable, versionable, and syncable.
* Works **offline first**. A cloud API exists only to sync vault files and distribute built-in/add-on packs.

---

## 2. High-Level Architecture

### Core (Go)

* Developed **first**.
* Responsible for:

  * File-based vault (groups, snippets, settings, counters, history).
  * Query parser, precedence rules, template rendering.
  * Built-in functions (`date`, `uuid`, `counter`, `clipboard`, `random`).
* Exposes a simple API (`Expand`, `Preview`, CRUD).
* All apps (desktop, mobile, extension) embed/use this core.

### API / Sync Service (NestJS)

* Manages **authentication, authorization, cloud sync** of vault files.
* Users own their vaults, stored as files + blobs.
* Provides built-in snippet packs, add-ons, and (future) sharing/collaboration.
* **Stack:** NestJS + Postgres + Firebase Auth + AWS S3 + CloudFront.

### Apps

* **Windows Desktop:** System tray app + global keyboard hook (`SetWindowsHookEx`) + text injection (`SendInput`); YAML editor for snippets.
* **Browser Extension:** MV3 + TypeScript + WASM (Go core compiled); content script expands in `<textarea>` and `contenteditable`.
* **Mobile (Android/iOS):** Custom keyboards (IME/Keyboard Extension) call into the Go core (via gomobile). Store vault locally, sync with API.
* All apps use **the same vault file format**.

---

## 3. Vault (File/Folder Storage)

Example structure:

```
/vault/
  groups/
    10-personal/
      group.yaml
      snippets/
        ty.yaml
        date.yaml
    20-work/
      group.yaml
      snippets/
        addr.yaml
        inv.yaml
  settings.yaml
  counters.json
  history.jsonl
```

### Snippet File (`*.yaml`)

```yaml
id: "snp_ty"
name: "Thanks"
trigger: ":ty"
description: "Quick thanks in multiple languages"
strict: false
defaults:
  lang: "en"
  tone: "neutral"
template: |
  {{ if eq .lang "vi" }}{{ if eq .tone "casual" }}Cảm ơn bạn nha!{{ else }}Cảm ơn bạn.{{ end }}
  {{ else if eq .lang "ja" }}ありがとうございます。
  {{ else }}{{ if eq .tone "casual" }}Thanks!{{ else }}Thank you.{{ end }}{{ end }}
```

### Settings

```yaml
prefix: ":"
expandKey: "Tab"
strictBoundaries: true
excludedApps:
  - "code.visualstudio"
  - "jetbrains.intellij"
locale: "en-US"
defaultDateFormat: "2006-01-02"
timezone: "Local"
historyEnabled: true
historyLimit: 200
pinForSensitive: true
```

### Counters

```json
{
  "inv": {"value": 123, "step": 1, "start": 1, "updatedAt": "2025-08-27T14:12:00Z"}
}
```

---

## 4. Core (Go) — Public API

```go
type Core interface {
  OpenVault(path string) error
  Reload() error
  Save() error

  Expand(input TriggerInput) (Rendered, error)
  Preview(input TriggerInput) (string, error)

  ListGroups() ([]Group, error)
  ListSnippets(groupID string) ([]Snippet, error)
  UpsertSnippet(s Snippet) error
  DeleteSnippet(id string) error

  GetSettings() (Settings, error)
  SaveSettings(Settings) error
  NextCounter(name string, opts CounterOpts) (string, error)
}

type TriggerInput struct {
  RawTrigger string    // ":ty?lang=vi&tone=casual"
  AppID      string    // optional (per-app exclusions)
  Now        time.Time // testability
}

type Rendered struct {
  Output       string
  CursorOffset int
  UsedSnippet  string
  UsedParams   map[string]any
}
```

* Uses Go `text/template` with helpers (`date`, `uuid`, `counter`, etc.).
* Query params merged as: `query > snippet.defaults > global defaults`.
* Secure randomness via `crypto/rand`.

---

## 5. API / Sync Service (NestJS)

### Authentication

* Firebase Auth for identity (email/password, Google, GitHub, etc.).
* Clients exchange Firebase token → Nest issues short-lived API JWT.

### Vault & Files

* Vault = index of file paths → blob hashes.
* Blobs stored in **AWS S3** (immutable, content-addressed).
* Server only manages metadata (index, ACL, packs).
* Devices pull index + missing blobs; push mutations + new blobs.

### Endpoints

* `POST /auth/exchange` → { accessToken, refreshToken }
* `GET /v1/vault/index?since=ver` → changed files
* `POST /v1/vault/mutations` → apply changes, detect conflicts
* `POST /v1/blobs` → get signed S3 PUT URL
* `GET /v1/blobs/:sha256` → signed S3 GET URL
* `GET /v1/packs` → list built-in/add-on packs
* `POST /v1/packs/:id/install` → merges pack into user vault

### Sharing (future)

* ACL table (owner, editor, viewer).
* Allow sharing groups/snippets or whole vaults.
* Consumer can install shared pack (snapshot) or live-shared subtree (later).

---

## 6. Built-in & Add-on Snippets

* Distributed as **packs** (like npm modules).
* Example: “Official Utilities” pack (date, uuid, counter).
* Stored in S3, versioned.
* Installed by merging files into vault (`groups/official-utilities/...`).
* Updates create conflicts if user modified same snippets.

---

## 7. Tech Stack

### Core

* Language: **Go** (1.22+)
* Templates: `text/template` + helper funcs
* Secure RNG: `crypto/rand`
* Storage: local FS (YAML/JSON)

### API

* **NestJS** (TypeScript)
* **Postgres** (Prisma/TypeORM)
* **Firebase Auth** (Admin SDK)
* **AWS S3** (+ CloudFront for packs)
* **Redis** + BullMQ (background jobs)

### Apps

* **Windows:** Go (Wails) or C# shell with Go core; keyboard hook + SendInput.
* **Browser Extension:** TypeScript + WASM (Go core compiled).
* **Android:** Kotlin IME + Go core (gomobile).
* **iOS:** Swift keyboard extension + Go core (gomobile).

---

## 8. Roadmap / Milestones

### Phase 1 — Core (Go)

* Vault loader/saver, schema validation.
* Query parser + precedence logic.
* Template rendering with functions.
* Expand/Preview API.
* CLI tool for testing (`randx expand ':ty?lang=vi'`).

### Phase 2 — Windows App MVP

* Tray app + editor for vault.
* Global keyboard hook + expansion.
* Per-app exclusions + pause toggle.

### Phase 3 — API + Sync

* Auth (Firebase + JWT exchange).
* Vault index/blobs endpoints.
* Device registration.
* Built-in packs listing & install.
* Sync client in Windows app.

### Phase 4 — Browser Extension

* Content script for textareas.
* WASM core.
* Chrome/Edge/Firefox support.

### Phase 5 — Mobile Apps

* Android IME (gomobile).
* iOS Keyboard Extension (gomobile).
* Offline vault + sync.

### Phase 6 — Add-on Marketplace

* Search, install, update packs.
* Community submissions.

### Phase 7 — Sharing & Teams

* ACLs, shared vault subtrees.
* Viewer/editor roles.
* Real-time collab (v2+).

---

## 9. Example Snippets

**Date**

```yaml
id: "snp_date"
name: "Date"
trigger: ":date"
defaults:
  format: "2006-01-02"
  tz: "Local"
template: "{{ date .format .tz }}"
```

Usage: `:date?format=Mon, 02 Jan 2006&tz=UTC`

**UUID**

```yaml
id: "snp_uuid"
name: "UUID"
trigger: ":uuid"
defaults:
  upper: false
template: "{{ if .upper }}{{ upper (uuid false) }}{{ else }}{{ uuid false }}{{ end }}"
```

Usage: `:uuid?upper=1`

**Invoice Counter**

```yaml
id: "snp_inv"
name: "Invoice"
trigger: ":inv"
defaults: { pad: 5 }
template: "Invoice #{{ counter \"inv\" .pad }} — {{ date \"2006-01-02\" \"Local\" }}"
```

---

# ✅ Next Steps for Development Agent

1. **Initialize repo structure** with `/core`, `/sync-api`, `/apps/windows`.
2. **Implement Go core**:

   * Vault loader (YAML/JSON).
   * Query parser.
   * Template funcs.
   * `Expand()` logic + golden tests.
3. **Create Windows CLI** to validate expansions.
4. **Start NestJS API**:

   * Firebase Auth exchange.
   * Vault index/blobs endpoints.
   * Device register.
   * Built-in pack listing.
5. **Integrate Windows App** with sync client.
