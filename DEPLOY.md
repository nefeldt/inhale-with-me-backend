# Deploying to mittwald

The backend deploys to **mittwald Container Hosting** (inside an mStudio project on a vServer / Dedicated Server). CI builds a Docker image, pushes it to **GHCR**, and the official **`mittwald/deploy-container-action`** updates a container stack.

```
push to main/tag ──▶ GitHub Actions
                       ├─ docker build ──▶ ghcr.io/<owner>/<repo>:<tag>
                       └─ mittwald/deploy-container-action (updates the stack)
                             └─ mittwald pulls the image ──▶ runs the container
                                   (persistent volume /data holds the SQLite file)
```

## One-time setup

### 1. mittwald API token
mStudio → your profile → **API tokens** (<https://studio.mittwald.de/app/profile/api-tokens>). Create a token with the **`api_write`** role. Copy it once — it can't be retrieved later.

### 2. Create the stack (to obtain its ID)
The deploy action *updates* an existing stack, so create it once. Locally, with the `mw` CLI (`brew install mw` — already installed here) and `deploy/stack.yaml`:

```bash
export MITTWALD_API_TOKEN=<your-token>
mw login status                      # verify auth

# Pick the target project:
mw project list

# Create the stack from the compose-style definition (fill IMAGE_REF/JWT_SECRET
# with any placeholder for this first create; CI overwrites them on deploy):
mw stack deploy -p <project-id> -c deploy/stack.yaml \
  --env IMAGE_REF=ghcr.io/<owner>/<repo>:latest \
  --env JWT_SECRET=$(openssl rand -hex 32)

# Note the stack UUID it prints (or: mw stack list -p <project-id>).
```

> The exact `mw stack` flags and `deploy/stack.yaml` field names (`envs`, `ports`, `volumes`, `stack_file` vs `stack-file`) should be eyeballed against the current docs — <https://developer.mittwald.de/docs/v2/platform/workloads/containers/> and the action README <https://github.com/mittwald/deploy-container-action> — as the API evolves. The mechanism (build → GHCR → deploy-container-action with a stack UUID) is confirmed.

### 3. GitHub repository secrets
In the repo's **Settings → Secrets and variables → Actions → Secrets** tab (all three
are read via `secrets.` — do NOT put them under the *Variables* tab, or the deploy
action panics with `Missing required environment variable INPUT_API_TOKEN`):

| Name | Value |
|------|-------|
| `MITTWALD_API_TOKEN` | the `api_write` token from step 1 |
| `STACK_ID` | the stack UUID from step 2 |
| `JWT_SECRET` | `openssl rand -hex 32` (the production signing secret) |

`GITHUB_TOKEN` (for pushing to GHCR) is provided automatically.

### 3b. Push notifications (optional — APNs)
Add these secrets to enable "a friend just smoked" push. Leave them unset and the
server runs with push disabled (no-op). Create one **APNs Auth Key** in the Apple
Developer portal → Certificates, Identifiers & Profiles → **Keys** → **+** →
enable **Apple Push Notifications service (APNs)** → download `AuthKey_XXXXXXXXXX.p8`
(downloadable once).

| Secret | Where to get it |
|--------|-----------------|
| `APNS_KEY_P8_BASE64` | base64 of the downloaded key: `base64 -i AuthKey_XXXXXXXXXX.p8 \| pbcopy` |
| `APNS_KEY_ID` | the 10-char Key ID (shown on the key page / in the filename) |
| `APNS_TEAM_ID` | your Apple Developer Team ID (Membership page, 10 chars) |

`APNS_PRODUCTION=true` is set in the workflow (TestFlight/App Store use the
production APNs host). `APNS_BUNDLE_ID` defaults to `feldt.systems.Inhale-With-Me`.

### 4. GHCR image visibility
The first push creates the GHCR package. Either make it **public** (Packages → package → visibility), or keep it private and register pull credentials with the mittwald project so it can pull:

```bash
mw registry create --uri ghcr.io --username <github-user> --password <a GHCR read:packages PAT>
```

### 5. Expose it on a domain (one-time)
Map a hostname to the container's port 8080 (HTTP/S ingress; only HTTP(S) is publicly exposable):

```bash
mw domain virtualhost create --hostname api.your-domain.tld \
  --path-to-container "/:<container-uid>:8080/tcp"
```

## Deploying

Every push to `main` (or a `v*` tag) runs `.github/workflows/deploy.yml`: build → push to GHCR → deploy. Trigger manually from the Actions tab (`workflow_dispatch`) too.

## Local container check

```bash
docker build -t inhale-with-me-backend:local .
docker run --rm -p 8080:8080 \
  -e JWT_SECRET=local-secret \
  -v inhale-data:/data \
  inhale-with-me-backend:local
curl -s localhost:8080/healthz
```

## Notes & tradeoffs

- **SQLite on a volume** is fine for an MVP: the `inhale-data` volume persists across recreations and is included in mittwald project backups. For scale/concurrency, switch to **managed PostgreSQL** (a one-line GORM driver change) and pass its connection string via env — see <https://developer.mittwald.de/docs/v2/platform/databases/postgresql/>.
- Container Hosting requires a plan that includes the feature (included with vServer/Dedicated Server).
- Alternative zero-config path (no Dockerfile; Railpack auto-detects Go): `mittwald/zerodeploy-action` with `mittwald-project-id`.
