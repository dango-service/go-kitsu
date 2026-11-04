# AGENTS.md  -  go-kitsu

Single source of truth for AI agents working in this repository.
Read this entire file before modifying any code or documentation.

---

## 1. Project Purpose

**go-kitsu** is an idiomatic Go client library for the **Kitsu.io** JSON:API (v3 / edge).
Its goal is to provide a strongly typed, memory-efficient client with zero external
dependencies, primarily consumed by **Dango** and any Go application that requires
interaction with Kitsu's anime tracking database.

This library does not implement Dango domain logic, Plex webhook handling, or local
database operations; it is a standalone, agnostic client library.

---

## 2. Map of `docs/`

```
docs/
  ai/                        <- Rules for AI agents
    architecture-rules.md    <- Library architectural invariants
    commits.md               <- Conventional Commits 1.0.0 rules and examples
    go-style.md              <- Go style guide and error wrapping rules
    resource-budget.md       <- Resource efficiency and memory budget rules
    workflow.md              <- Step-by-step agent workflow
  reference/
    kitsu-api.md             <- Kitsu.io JSON:API technical specifications
  design/
    jsonapi.md               <- JSON:API standard specification compliance
    public-api.md            <- Exact Go public API contracts and signatures
    scope.md                 <- Scope and bounded context of supported features
    testing.md               <- Deterministic testing strategy with httptest
    versioning.md            <- SemVer policy and release lifecycle
```

---

## 3. Build / Test / Lint Commands

```bash
# Compile library
CGO_ENABLED=0 go build ./...

# Run unit tests
go test -v ./...

# Static analysis
go vet ./...
```

---

## 4. Go Style Rules

See full details in `docs/ai/go-style.md`. Summary:

- Errors must always be wrapped using `%w` (`fmt.Errorf("kitsu: %s: %w", op, err)`).
- `context.Context` must be the first parameter in all functions performing I/O or network calls.
- No mutable global variables and no `init()` functions.
- Client configured via functional options in `NewClient(httpClient *http.Client, opts ...ClientOption)`.
- Table-driven unit tests using `t.Run` and exclusive use of `httptest.Server` (zero real-network calls in tests).
- Strict `CGO_ENABLED=0` compatibility.
- Zero external dependencies beyond the Go standard library.
- Responses must be decoded using streaming `json.NewDecoder` bounded by `io.LimitReader(body, maxBytes)`.

---

## 5. Commit Convention

Format: `<type>(<scope>)<!>: <description>`

- Allowed types: `feat fix docs style refactor perf test build ci chore revert`
- Mandatory scope when applicable (e.g. `feat(client)`, `docs(design)`).
- Description in English, imperative mode, lowercase, no trailing period, <= 72 characters.
- Atomic commits per logical change. Never `--amend` or `push --force`.
- Never commit secrets, tokens, or personal credentials.

---

## 6. Resource and Memory Budget

- Ultra-lightweight: zero external dependencies (`net/http`, `encoding/json`, `context`).
- Streaming deserialization with `json.NewDecoder` bounded by `io.LimitReader`.
- Memory ceiling: peak RSS <= 5 MiB in background operations.
- Do not duplicate buffers or retain raw HTTP responses in memory unnecessarily.

---

## 7. Prohibited Actions

- Coupling `go-kitsu` to `Dango` concepts (Plex, webhooks, SQLite, sync loops).
- Importing `Dango` packages or logic into `go-kitsu`.
- Using mutable global variables or shared static state.
- Making real HTTP requests to Kitsu during unit tests (use `httptest.Server`).
- Hardcoding OAuth2 tokens, credentials, or personal secrets.
- Adding third-party external dependencies.
- Ignoring or discarding `context.Context` cancellation.
