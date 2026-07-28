# JOBHOO

A focused recruitment platform: candidates discover jobs, recruiters manage
hiring, AI assists (never replaces) the decisions on both sides.

## Stack

- **Backend:** Go, `net/http` + [chi](https://github.com/go-chi/chi) router
- **Frontend:** server-rendered `html/template` + [HTMX](https://htmx.org)
- **Database:** PostgreSQL (via `pgx`)
- **AI:** pluggable provider interface (`internal/ai`) — ships with a
  dependency-free mock provider; Anthropic/OpenAI providers are stubbed and
  ready to implement without touching any other file
- **Containerization:** Docker + Docker Compose

## Project layout

```
cmd/server/            entrypoint (config → db → ai → router → http.Server)
internal/
  config/               env-driven configuration, one struct, one place
  database/             pgx pool + repositories (SQL lives here only)
  database/migrations/   schema + seed SQL
  models/               shared domain types
  ai/                    Provider interface + mock/anthropic/openai implementations
  handlers/              HTTP handlers (thin: parse → call → render)
  router/                full route table in one file
web/
  templates/layouts/     base.html (nav, footer, atmosphere background)
  templates/components/  reusable partials (job-card, etc.)
  templates/pages/       one template per route
  static/css/            tokens.css (design tokens) + components.css
  static/img/            official JOBHOO logo assets
```

## Running locally

No extra tools needed beyond Docker Desktop — everything below is plain `docker compose`:

```bash
cp .env.example .env
docker compose up --build       # builds the app image and starts app + Postgres
docker compose run --rm app ./jobhoo-seed   # populates 10 companies/recruiters + 100 jobs
```

Then open http://localhost:8080.

Schema is applied automatically the first time the `db` container initializes
its data volume (via `internal/database/migrations/*.sql` mounted into
`docker-entrypoint-initdb.d`). Demo **data** (companies, jobs) is a separate,
explicit, re-runnable step rather than baked into container init, so you can
reseed at will without recreating the database:

```bash
docker compose run --rm app ./jobhoo-seed   # safe to re-run: clears prior demo data first
```

Demo recruiter accounts use `recruiter1@jobhoo.demo` .. `recruiter10@jobhoo.demo`,
password `demo-password-123`.

To wipe the database entirely and start over:
```bash
docker compose down -v && docker compose up --build
```

A `Makefile` is included as an optional shortcut for the commands above
(`make up`, `make seed`, `make down`, ...) if you have `make` installed —
it's not required, and every `make` target is just a one-line alias for the
`docker compose` command shown next to it in the Makefile.

### Running without Docker

```bash
# requires a local Postgres; create the DB and run the migration SQL yourself
go run ./cmd/server
go run ./cmd/seed     # optional: populate demo data
```

## Design system

All color, spacing, radius, and type values live in
`web/static/css/tokens.css`. Every component in `components.css` is built
exclusively from those variables — there should never be a reason to add a
new hex code or magic number outside that one file. See the brand palette:

| Token | Value | Use |
|---|---|---|
| `--jh-navy-700` | `#1F2747` | dominant structural color |
| `--jh-orange-500` | `#FF7A00` | the only accent — CTAs, active states, key metrics |
| `--jh-white` | `#FFFFFF` | typography, content |
| `--jh-black` | `#0A0C16` | high-contrast needs only |

## AI architecture

`internal/ai.Provider` is the only interface handlers depend on. Swapping the
underlying model/vendor means writing one new `provider_<name>.go` file and
adding one line to `internal/ai/registry.go` — nothing else in the codebase
changes. Set `AI_PROVIDER=anthropic|openai|mock` in `.env` to select.

## What's implemented

The core product loop is now closed end-to-end against real Postgres data:

- **Discovery**: homepage, job search, category filter, pagination (all HTMX)
- **Auth**: signup/login/logout, bcrypt password hashing, revocable DB-backed sessions, role-based route guards
- **Candidate flow**: job detail page, apply (with cover note), save/bookmark jobs, profile (headline/resume/skills), AI resume improvement suggestions, AI job recommendations, applications + saved jobs dashboard
- **Recruiter flow**: company setup, post a job, recruiter dashboard (jobs + applicant counts), a full ATS Kanban board (Applied → Screening → Interview → Offer → Hired, plus a collapsed Rejected section), stage changes, and one-click AI candidate ranking (advisory only — never changes stage or hides anyone)
- **Admin**: simple platform-wide counts (users, candidates, recruiters, companies, jobs, applications) — deliberately no charts/enterprise complexity, per the brief
- **AI**: `internal/ai.Provider` interface; `MockProvider` (real keyword-overlap logic, zero external calls) and a fully implemented `AnthropicProvider` that calls the real Anthropic Messages API with structured-JSON prompts for ranking, match explanation, resume summarization, job recommendations, and resume advice. `OpenAIProvider` remains a documented stub — swap `AI_PROVIDER=openai` once implemented, following the same pattern.
- **Security**: CSRF protection (double-submit cookie) on every state-changing request, ownership checks on recruiter routes (a recruiter can't view/edit another company's pipeline by guessing a job ID), session tokens hashed before storage
- **Data**: `cmd/seed` generates 10 recruiter accounts/companies and 100 jobs across all 5 categories with varied locations, salaries, skills, and publish dates
- Custom 404 page matching the design system

## Known gaps / what I'd do before a real launch

Being direct about what's *not* here, since "production-ready" is the brief's stated bar:

- **No file upload** — resumes are plain text pasted into a textarea, not PDF/DOCX upload. Wiring that up needs an object storage decision (S3-compatible, etc.) I didn't want to make unilaterally.
- **No rate limiting** on login/signup/apply — worth adding before public launch to blunt credential-stuffing and spam applications.
- **CSRF is the standard double-submit-cookie baseline**, not a per-session token — a reasonable next hardening step, not a hole, but worth knowing.
- **No public company directory page** (`/companies` — removed from nav since it wasn't built; the route is still a TODO).
- **No email** anywhere (no verification, no notifications on stage changes, no password reset). Candidates/recruiters currently have no way to recover a lost password.
- **Accessibility**: keyboard focus states and semantic markup are in place from the design system, but there's been no full screen-reader pass.
- **I could not run a full `go build`/`go vet` in my own sandbox** — its network is limited to a small domain allowlist that excludes the Go module proxy and a couple of transitive dependencies' vanity import domains. What I *could* and did verify from inside the sandbox: `gofmt -e` syntax-checked every `.go` file with zero errors; `internal/models`, `internal/ai` (including the real Anthropic HTTP client), and `internal/config` — which only depend on the standard library — compiled cleanly with `go build`; every method a handler calls against a repository was cross-checked by name against what's actually defined; and, most importantly, every one of the 13 page templates was both parsed *and executed* through Go's real `html/template` engine with mock data shaped like the actual structs, which catches field/method-name mismatches that syntax checking alone would miss. That's meaningfully more verification than "it looks right," but it is not a substitute for `go build ./...` on a machine with normal internet access — please run that (and ideally `go vet ./...`) before deploying, and let me know what comes up.
