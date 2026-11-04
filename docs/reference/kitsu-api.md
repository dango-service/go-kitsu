# Kitsu.io JSON:API Reference  -  `go-kitsu`

Technical reference for interacting with Kitsu's Edge JSON:API.

---

## 1. Endpoints Overview

| Resource | HTTP Method | Path | Description |
|----------|-------------|------|-------------|
| **Anime** | `GET` | `/api/edge/anime` | Search and filter anime titles |
| **Anime** | `GET` | `/api/edge/anime/{id}` | Fetch individual anime by ID |
| **Mappings**| `GET` | `/api/edge/mappings` | Resolve external IDs (MAL, TVDB) to Kitsu ID |
| **Library** | `GET` | `/api/edge/library-entries` | Query user library entries |
| **Library** | `POST` | `/api/edge/library-entries` | Create new library entry |
| **Library** | `PATCH`| `/api/edge/library-entries/{id}` | Update watch progress or status |
| **Library** | `DELETE`| `/api/edge/library-entries/{id}` | Remove library entry |
| **Users** | `GET` | `/api/edge/users` | Query user profiles (filter by slug/self) |
| **OAuth2** | `POST` | `/api/oauth/token` | Obtain access and refresh tokens |

---

## 2. Note on `api-docs.kitsu.cloud` vs `kitsu.io`

- `https://api-docs.kitsu.cloud/` is **CGWire Zou**, an animation studio production tracking REST API (assets, shots, tasks, sequences for 3D/VFX studios).
- `https://kitsu.io/api/edge/` is the **Kitsu Anime Platform**, the official consumer API for anime scrobbling, episode tracking, and watch lists.
- This library targets the **Kitsu.io** API to fulfill the needs of **Dango** and anime catalog tools.
