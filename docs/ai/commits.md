# Commit Convention  -  go-kitsu

All commits in this repository must strictly follow the **Conventional Commits 1.0.0** specification.

---

## 1. Commit Structure

```
<type>(<scope>)<!>: <description>

[optional body: lines <= 72 characters]

[optional footer(s): BREAKING CHANGE:, Refs:]
```

### Mandatory Rules
- **type**: One of the allowed types below (in lowercase).
- **scope**: Mandatory whenever the change targets a specific subsystem.
- **description**: In English, imperative mood ("add", "fix", "update", never "added" or "adds"), lowercase start, no trailing period, maximum 72 characters.
- **breaking change**: Append `!` immediately after type/scope (e.g. `feat(client)!: change auth signature`) and include `BREAKING CHANGE: <desc>` in the footer.
- **atomicity**: One commit per logical unit of work. Never bundle unrelated changes into a single monolithic commit.

---

## 2. Allowed Types

| Type | Purpose | Example |
|------|---------|---------|
| `feat` | New public feature or method | Implement LibraryEntryService |
| `fix` | Bug fix or error handling correction | Fix JSON:API error array parsing |
| `docs` | Documentation-only changes | Add API specification |
| `style` | Formatting, whitespace (no code change) | Run gofmt on source files |
| `refactor` | Refactoring code without altering public API | Simplify internal JSON decoder |
| `perf` | Performance or memory allocation optimization | Limit reader buffer size |
| `test` | Adding or updating unit tests | Add httptest search fixtures |
| `build` | Build system or tooling dependencies | Adjust Go toolchain version |
| `ci` | Continuous integration workflows | Add GitHub Actions test matrix |
| `chore` | Maintenance tasks | Initialize .gitignore or licenses |
| `revert` | Reverting a previous commit | Revert previous commit |

---

## 3. Common Scopes in go-kitsu

- `client`: Base HTTP / JSON:API client core.
- `anime`: Anime lookup and search service.
- `library`: User library entry query and mutation service.
- `mapping`: External ID resolution service.
- `user`: User profile service.
- `auth`: OAuth2 token exchange, password grant, and refresh service.
- `jsonapi`: Internal JSON:API transport and error handling.
- `design`: Architecture and design documentation.
- `reference`: API specifications and reference analysis.
- `ai`: Guidelines for AI agents.
- `repo`: Repository metadata and configuration.
