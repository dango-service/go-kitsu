# AI Agent Workflow  -  go-kitsu

Methodology and rigorous workflow that all AI agents must follow when modifying this repository.

---

## 1. Core Principle: "Read Before Writing"

1. **Consult the Sources of Truth**:
   - `AGENTS.md` for project rules and invariants.
   - `docs/reference/kitsu-api.md` to validate Kitsu JSON:API specifications.
   - `docs/design/public-api.md` to verify public API contracts.
2. **Do Not Invent APIs or Fields**:
   - Every attribute in a struct must match official Kitsu JSON:API resource schemas.
   - Public signatures must follow established design decisions.

---

## 2. Managing Dependencies and Decisions

- **No Arbitrary Dependencies**: Any proposed addition to `go.mod` must be evaluated
  and formally justified in `docs/design/`.
- **Clarifying Ambiguities**: When facing ambiguities, consult official API specs and repository invariants before making assumptions.

---

## 3. Feature Implementation Cycle

1. **Define the JSON:API Endpoint and Payload**: Document exact path, headers, and attributes.
2. **Define Go Types**: Create compact, strongly typed Go structs.
3. **Write Tests with Fixtures**:
   - Add mock JSON responses in `testdata/`.
   - Write unit tests using `httptest.Server` verifying sent payloads, headers, and output.
4. **Implement Service Methods**: Connect JSON:API transport layer to the public service.
5. **Verify Quality**:
   - Run `go test -v ./...`.
   - Run `go vet ./...`.
6. **Atomic Commit**: Commit following Conventional Commits 1.0.0.
