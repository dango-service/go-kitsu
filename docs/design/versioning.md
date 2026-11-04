# Versioning Policy and Lifecycle  -  `go-kitsu`

This document defines the versioning scheme, Conventional Commits relationship,
and backwards compatibility policy for `go-kitsu`.

---

## 1. SemVer 2.0.0

`go-kitsu` strictly follows **Semantic Versioning 2.0.0** (`vMAJOR.MINOR.PATCH`):

- **MAJOR**: Incompatible API breaking changes.
- **MINOR**: Backwards-compatible new features (new services, options, methods).
- **PATCH**: Backwards-compatible bug fixes or internal performance refactoring.

---

## 2. Initial Phase: `v0.x` Cycle

1. **Initial Release**: The first functional release tag is **`v0.1.0`**.
2. **Evolution in `v0.x`**:
   - In accordance with SemVer, during `v0.x` the API may evolve rapidly.
   - Breaking changes in this phase increment the `MINOR` number (e.g. `v0.1.0` -> `v0.2.0`).
   - Additions and fixes increment the `PATCH` number (e.g. `v0.1.0` -> `v0.1.1`).
3. **Transition to `v1.0.0`**:
   - `go-kitsu` will be tagged `v1.0.0` once `Dango` is operating reliably in production and client signatures are consolidated.

---

## 3. Conventional Commits and Releases

| Commit Type | Effect in `v0.x` | Effect in `v1.x+` |
|-------------|:----------------:|:-----------------:|
| `fix(...)` | `PATCH` (`v0.1.0` -> `v0.1.1`) | `PATCH` (`v1.0.0` -> `v1.0.1`) |
| `perf(...)` | `PATCH` (`v0.1.0` -> `v0.1.1`) | `PATCH` (`v1.0.0` -> `v1.0.1`) |
| `feat(...)` | `PATCH` or `MINOR` | `MINOR` (`v1.0.0` -> `v1.1.0`) |
| `feat(...)!` or `BREAKING CHANGE:` | `MINOR` (`v0.1.0` -> `v0.2.0`) | `MAJOR` (`v1.0.0` -> `v2.0.0`) |
| `docs(...)`, `style(...)`, `chore(...)`, `test(...)` | No version bump | No version bump |

---

## 4. Git Tagging and Go Modules

- Git tags must use the `v` prefix (e.g. `v0.1.0`, `v0.1.1`).
- The canonical module path in `go.mod` is:
  ```go
  module github.com/dango-service/go-kitsu
  ```
