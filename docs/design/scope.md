# Library Scope and Boundaries  -  `go-kitsu`

This document defines the functional perimeter of `go-kitsu`.
The library does **not** attempt to cover the entirety of Kitsu's API,
focusing strictly on the needs of **Dango** and media tracking tools.

---

## 1. Boundary Principle

The official Kitsu JSON:API exposes extensive endpoints (social feeds, comments, drama, reaction posts, group memberships, notifications, and store merchandise).
Building an exhaustive client would:
- Multiply compiled binary size and build times.
- Increase runtime allocations with zero added value for scrobblers.
- Introduce an unmanageable maintenance burden.

Therefore, `go-kitsu` follows a **strict minimum viable scope**, focusing on
searching, querying, and updating anime watch progress.

---

## 2. In-Scope Components

| Component / Service | Methods & Resources | Technical Justification in Dango |
|---------------------|---------------------|-----------------------------------|
| **Auth / OAuth2** | `LoginPassword`, `ExchangeCode`, `RefreshToken` | Enables linking user Kitsu accounts via password or OAuth2 code grant. |
| **User (`UserService`)** | `CurrentUser(ctx)`, `GetByID(ctx, id)` | Identifies the authenticated user and user ID for library operations. |
| **Anime (`AnimeService`)** | `GetByID(ctx, id)`, `Search(ctx, query, opts...)` | Matches Plex titles (canonical, English, Romaji) and fetches episode count. |
| **Mapping (`MappingService`)** | `FindAnimeByExternalID(ctx, site, id)` | Resolves MyAnimeList / AniList / TVDB IDs to Kitsu Anime ID. |
| **LibraryEntry (`LibraryEntryService`)** | `GetEntry(ctx, userID, animeID)`, `GetCollection(ctx, userID, opts...)`, `SaveEntry(ctx, input)`, `DeleteEntry(ctx, id)` | Core scrobbling: checks watch status, mutates episode progress and completion (`current`, `completed`). |
| **Transport & Quotas** | `doRequest`, `X-RateLimit-*`, `Retry-After`, `io.LimitReader` | Enforces JSON:API compliant calls, rate-limit awareness, and DoS protection. |

---

## 3. Out-of-Scope Components

| Schema Area | Status | Reason for Exclusion |
|-------------|:------:|----------------------|
| **Social Feeds (`posts`, `comments`, `follows`)** | **Excluded** | Dango is an unattended daemon; does not consume social features. |
| **Drama & Manga (`dramas`, `manga`, `chapters`)** | **Deferred** | Dango focuses on video streams. Types can be extended later. |
| **Reviews & Reactions (`media-reactions`)** | **Excluded** | High payload overhead unused in media matching. |
| **Franchises & Characters (`characters`, `castings`)** | **Excluded** | Relational overhead unused in scrobbling. |
| **Notifications (`notifications`)** | **Excluded** | Does not participate in synchronization. |
| **Store & Merchandise** | **Excluded** | Completely outside synchronization context. |
