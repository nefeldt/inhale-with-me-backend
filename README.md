# Inhale With Me — Backend

REST API for **Inhale With Me**, a social smoke-logging app (think "Beer With Me" / Untappd, but for cigarettes, joints, vapes, cigars and pipes). Log sessions, track stats, add friends and cheer each other's activity.

Written in **Go** with **chi** (routing), **GORM** (`glebarez/sqlite`, a pure-Go SQLite driver → static, cgo-free builds) and **JWT** auth. The companion iOS app lives in the sibling `Inhale With Me` repo.

## Requirements

- Go **1.27+**
- No C toolchain needed (pure-Go SQLite)

## Quick start

```bash
cp .env.example .env          # then set JWT_SECRET (openssl rand -hex 32)
go run ./cmd/server           # migrations run automatically; listens on :8080
```

`make run` / `make test` / `make build` are also available.

## Configuration (environment variables)

| Var | Default | Notes |
|-----|---------|-------|
| `PORT` | `8080` | HTTP port |
| `DB_PATH` | `./data/inhale.db` | SQLite file (dir auto-created) |
| `JWT_SECRET` | dev fallback | **required in production** |
| `JWT_TTL` | `720h` | access-token lifetime (Go duration) |
| `BCRYPT_COST` | `12` | password hashing cost |
| `CORS_ALLOWED_ORIGINS` | `*` | comma-separated |
| `ENVIRONMENT` | `development` | `development` \| `production` |

## Architecture

```
cmd/server            entrypoint (config → db → migrate → router → serve)
internal/config       env parsing + .env loader
internal/database     GORM/SQLite open + AutoMigrate
internal/model        entities (GORM tags) + API projection types + enums
internal/store        data access (one file per aggregate)
internal/auth         bcrypt, JWT manager, context, middleware
internal/api          chi router + handlers (auth, users, sessions, stats, friends, feed, reactions)
deploy/               mittwald Container Hosting stack definition
.github/workflows/    CI + build-and-deploy-to-mittwald pipeline
```

## Conventions

- All JSON keys are **snake_case**; timestamps are **RFC3339 UTC**; money is **integer cents**.
- Auth: send `Authorization: Bearer <token>` on every route except `POST /auth/*` and `GET /healthz`.
- Errors use one envelope: `{"error":{"code","message","fields?"}}`.
- Session visibility (`public`/`friends`/`private`) is enforced; a session you cannot see returns `404` (not `403`) so its existence isn't leaked.
- Auth model: a single stateless access token (default 30-day TTL); logout is client-side. See "future work" for refresh tokens.

## API reference (`/api/v1`)

🔓 public · 🔒 requires bearer token

| Method | Path | Auth | Purpose |
|--------|------|:----:|---------|
| GET | `/healthz` | 🔓 | liveness probe |
| POST | `/api/v1/auth/register` | 🔓 | `{email,username,password,display_name?}` → `{token,expires_at,user}` |
| POST | `/api/v1/auth/login` | 🔓 | `{login,password}` (login = email or username) |
| GET | `/api/v1/users/me` | 🔒 | own profile |
| PATCH | `/api/v1/users/me` | 🔒 | update `display_name/bio/avatar_url/username/currency` |
| GET | `/api/v1/users?query=&limit=` | 🔒 | search users by username/email prefix |
| GET | `/api/v1/users/{id}` | 🔒 | public profile + friend status |
| GET | `/api/v1/users/me/cost-settings` | 🔒 | per-type unit costs |
| PUT | `/api/v1/users/me/cost-settings` | 🔒 | replace cost settings |
| POST | `/api/v1/sessions` | 🔒 | log a session |
| GET | `/api/v1/sessions?limit=&before=&type=` | 🔒 | own sessions (cursor paginated) |
| GET | `/api/v1/sessions/{id}` | 🔒 | one session (if visible) |
| PATCH | `/api/v1/sessions/{id}` | 🔒 | update own session |
| DELETE | `/api/v1/sessions/{id}` | 🔒 | delete own session |
| GET | `/api/v1/stats/summary?tz=<IANA>` | 🔒 | today/week/month/all-time + streaks |
| POST | `/api/v1/friends/requests` | 🔒 | `{username}` or `{addressee_id}` |
| GET | `/api/v1/friends/requests?direction=incoming\|outgoing` | 🔒 | pending requests |
| POST | `/api/v1/friends/requests/{id}/accept` | 🔒 | accept (addressee only) |
| POST | `/api/v1/friends/requests/{id}/decline` | 🔒 | decline (addressee only) |
| DELETE | `/api/v1/friends/requests/{id}` | 🔒 | cancel own outgoing request |
| GET | `/api/v1/friends` | 🔒 | accepted friends |
| DELETE | `/api/v1/friends/{userId}` | 🔒 | unfriend |
| GET | `/api/v1/feed?limit=&before=` | 🔒 | friends' activity + reactions |
| GET | `/api/v1/sessions/{id}/reactions` | 🔒 | reactions on a session |
| POST | `/api/v1/sessions/{id}/reactions` | 🔒 | `{type?}` (default `cheers`) |
| DELETE | `/api/v1/sessions/{id}/reactions/{type}` | 🔒 | remove own reaction |

Session types: `cigarette joint vape cigar pipe other`. Moods: `great good neutral stressed bad`.

### Example

```bash
BASE=http://localhost:8080
TOKEN=$(curl -s -X POST $BASE/api/v1/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"noah@example.com","username":"noah","password":"hunter2hunter"}' \
  | python3 -c 'import sys,json;print(json.load(sys.stdin)["token"])')

curl -s $BASE/api/v1/users/me -H "Authorization: Bearer $TOKEN"
curl -s -X POST $BASE/api/v1/sessions -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"type":"cigarette","quantity":1}'
curl -s "$BASE/api/v1/stats/summary?tz=Europe/Berlin" -H "Authorization: Bearer $TOKEN"
```

## Deployment

Containerized deploy to **mittwald Container Hosting** via GitHub Actions (build → GHCR → `mittwald/deploy-container-action`). See **[DEPLOY.md](./DEPLOY.md)**.

## Future work

- Refresh tokens + short-lived access tokens + server-side revocation.
- Managed PostgreSQL (a one-line GORM driver swap) for horizontal scaling.
- Push notifications for friend requests and reactions.
