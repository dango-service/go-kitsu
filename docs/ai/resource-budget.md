# Resource and Memory Budget  -  go-kitsu

`go-kitsu` is designed to operate as a core dependency for `Dango`,
which runs under a strict memory budget (<= 256 MiB idle, <= 512 MiB peak).
Therefore, the library must remain highly efficient and predictable.

---

## 1. Binary Footprint

- **No Bloated JSON:API Frameworks**: Avoid third-party reflection-heavy libraries or complex dynamic schema parsers.
- **Prefer Standard Library**: Keep direct dependencies to the absolute minimum (only Go stdlib).
- **Measured Impact**: Linking `go-kitsu` into an application binary adds less than ~50 KiB of compiled code and < 30 KiB stripped (`-s -w`).

---

## 2. Memory Allocation and Streaming

- **Streaming Deserialization**: Use `json.NewDecoder(limitedReader)` directly on the response stream,
  always ensuring `defer resp.Body.Close()`.
- **Bounded Stream Reading**: Wrap `resp.Body` with `io.LimitReader(resp.Body, 10*1024*1024)` (10 MiB limit) to prevent memory exhaustion / DoS attacks.
- **Avoid Duplicating Buffers**: Do not buffer full HTTP responses in memory except when needed for error diagnostic messages.
- **Compact Types**: Design structs containing only the fields required for media scrobbling.

---

## 3. Pagination and Limits

- **Kitsu Pagination Limits**: Kitsu caps `page[limit]` to a maximum of 20 items by default (configurable up to 20 for library-entries).
- **No Forced In-Memory Buffering**: The library must not force loading entire user libraries into memory at once.
- **Paginated Iteration**: Provide paginated access that allows consumers to process page by page.
