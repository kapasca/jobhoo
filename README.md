# 🎯 JOBHOO — Smart Recruitment Platform

<div align="center">

![JOBHOO](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat-square&logo=postgresql)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)
![Status](https://img.shields.io/badge/Status-Production%20Ready-success?style=flat-square)

**A modern recruitment platform connecting job seekers with employers efficiently**

[Overview](#-overview) • [Features](#-features) • [Tech Stack](#-tech-stack) • [Quick Start](#-quick-start) • [Documentation](#-documentation)

</div>

---

## 📋 Overview

**JOBHOO** is a full-featured recruitment platform built with Go and modern web technologies. It streamlines the hiring process by connecting candidates with companies, managing applications, and providing intelligent matching powered by AI.

### The Problem We Solve
Traditional job boards are cluttered with irrelevant postings. JOBHOO focuses on **quality matches** — helping candidates find roles that fit their skills and helping companies find qualified candidates efficiently.

### Who Uses JOBHOO?
- 👥 **Job Seekers** — Find relevant opportunities, manage applications
- 🏢 **Recruiters** — Post jobs, review applicants, track hiring pipeline
- 🔑 **Admins** — Verify companies, ensure platform quality

---

## ✨ Features

### For Job Seekers 🙋
- ✅ User registration with email verification
- ✅ Professional profile management with resume upload
- ✅ Advanced job search with filters (location, category, employment type)
- ✅ Apply to jobs with one-click application
- ✅ Save jobs for later
- ✅ AI-powered job recommendations
- ✅ Application tracking dashboard
- ✅ Email notifications for updates

### For Recruiters 🏢
- ✅ Company registration & profile setup
- ✅ Post job listings with rich descriptions
- ✅ View and manage applications (ATS board)
- ✅ Rank candidates by quality
- ✅ Update application statuses
- ✅ Close, archive, reopen jobs
- ✅ Company approval workflow

### For Admins 🔑
- ✅ Company verification & approval queue
- ✅ User management dashboard
- ✅ Freeze/unfreeze user accounts
- ✅ Modal-based detail views for quick review
- ✅ Audit trails for all actions

### Platform Features 🌟
- ✅ Rate limiting on authentication endpoints (brute-force protection)
- ✅ CSRF protection on all state-changing operations
- ✅ Email audit logging (compliance & debugging)
- ✅ Responsive design for mobile/tablet/desktop
- ✅ Dark theme with professional styling
- ✅ Real-time notifications
- ✅ AI-powered resume analysis & job recommendations

---

## 🛠 Tech Stack

### Backend
- **Language:** Go 1.22
- **Framework:** Chi (v5) — lightweight HTTP router with middleware
- **Database:** PostgreSQL 16 with pgx (v5) driver
- **Connection Pool:** 20 max connections, 2 min connections
- **Authentication:** Session-based with bcrypt hashing
- **AI Integration:** Anthropic Claude API (with mock fallback)
- **Email:** SMTP with dev/testing mode support

### Frontend
- **Templating:** Go `html/template`
- **CSS:** Custom CSS with design tokens (dark theme)
- **Interactivity:** HTMX for async operations
- **Modal System:** Fetch-based modal loading for admin features
- **Styling:** Navy (#192132) + Orange (#d96600) brand colors

### DevOps & Deployment
- **Containerization:** Docker & Docker Compose
- **Database Migrations:** SQL-based versioning (13 migrations)
- **Environment Config:** `.env` with sensible defaults
- **Graceful Shutdown:** Context-based cleanup on SIGINT/SIGTERM

### Testing & Quality
- **Code Quality:** Follows Go conventions, proper error handling
- **Security:** CSRF tokens, rate limiting, parameterized queries
- **Logging:** Structured error logging throughout
- **No External Dependencies:** Minimal, focused dependency set

---

## 🚀 Quick Start

### Prerequisites
- **Go 1.22+** or **Docker & Docker Compose**
- **PostgreSQL 16+** (if running locally)
- **.env file** (copy from `.env.example`)

### Option 1: Using Docker (Recommended) 🐳

```bash
# Clone repository
git clone https://github.com/yourusername/jobhoo.git
cd jobhoo

# Copy environment template
cp .env.example .env

# Start all services (app + database)
docker compose up -d

# Database migrations run automatically
# App available at http://localhost:8070
```

**That's it!** Docker Compose handles:
- PostgreSQL database setup
- Database migrations
- Application server
- All networking

### Option 2: Local Development 💻

```bash
# Prerequisites: Go 1.22 and PostgreSQL 16

# Clone repository
git clone https://github.com/yourusername/jobhoo.git
cd jobhoo

# Copy and configure .env
cp .env.example .env
# Edit .env:
# - DATABASE_URL: Update to your local postgres connection
# - PORT: Change if 8070 is in use (default: 8070)
# - EMAIL_PROVIDER: Set to "dev" for console logging

# Download dependencies
go mod download

# Run database migrations
cd internal/database/migrations && bash init-migrations.sh

# Start application
go run ./cmd/server

# Visit http://localhost:8070
```

### Quick Verification

```bash
# Check application is running
curl http://localhost:8070/healthz
# Expected: {"status":"ok"}

# View application logs
docker compose logs -f app

# Access database
docker compose exec db psql -U jobhoo -d jobhoo
```

---

## 📁 Project Structure

```
jobhoo/
├── cmd/                          # Entry points
│   ├── server/main.go           # Application server
│   └── seed/                    # Database seeding
├── internal/                     # Core application (non-exported)
│   ├── ai/                      # AI provider abstraction (Claude, OpenAI, mock)
│   ├── auth/                    # Authentication utilities
│   ├── config/                  # Configuration loading
│   ├── database/                # Repository pattern
│   │   ├── *_repo.go           # Data access layer (11 repos)
│   │   ├── database.go         # Connection pool setup
│   │   └── migrations/         # SQL migrations (13 versions)
│   ├── email/                   # Email system with logging
│   ├── handlers/                # HTTP request handlers (15 files)
│   ├── middleware/              # Request processing
│   │   ├── auth.go             # User/session loading
│   │   ├── csrf.go             # CSRF protection
│   │   └── rate_limit.go       # Brute-force protection
│   ├── models/                  # Domain models
│   └── router/                  # HTTP routing & setup
├── web/                          # Frontend assets
│   ├── static/                  # CSS, JS, images
│   │   └── css/components.css   # Brand styling (dark theme)
│   └── templates/               # HTML templates
│       ├── components/          # Reusable blocks (modals, cards)
│       ├── pages/              # Full page templates (22 pages)
│       └── layouts/            # Base layout wrapper
├── .env.example                 # Configuration template
├── .gitignore                   # Git ignore patterns
├── docker-compose.yml           # Local development environment
├── Dockerfile                   # Production container
├── go.mod & go.sum             # Go dependency management
├── Makefile                     # Development commands
└── DOC-*.md                    # Project documentation
```

### Key Directories Explained

**`internal/database/`** — Repository pattern for clean data access
- `users_repo.go`, `jobs_repo.go`, `companies_repo.go`, etc.
- Each model has dedicated CRUD methods
- All queries parameterized (SQL injection protection)
- Connection pooling handled at startup

**`internal/handlers/`** — HTTP request handlers
- Thin handlers: parse request → call repo → render template
- Separation of concerns: business logic in repos/models
- Role-based middleware for authorization

**`internal/middleware/`** — Request processing pipeline
- `auth.go` — Load user from session on every request
- `csrf.go` — Double-submit CSRF token protection
- `rate_limit.go` — Sliding window rate limiter (5 attempts/15 min for auth)

**`web/templates/`** — Template organization
- `components/` — Reusable blocks rendered with `RenderBlock()` for AJAX
- `pages/` — Full page templates with full layout
- Modal system: fetch-based loading to `/admin/api/*` endpoints

---

## 🔧 Installation & Configuration

### Environment Variables

See [DOC-DEVELOPMENT-GUIDE.md](DOC-DEVELOPMENT-GUIDE.md#5-konfigurasi-environment-variable) for the complete environment variable reference.

```env
# Application
APP_ENV=development              # development | production
PORT=8070                       # HTTP server port

# Database
DATABASE_URL=postgres://jobhoo:jobhoo_dev_password@localhost:5432/jobhoo?sslmode=disable

# Security
SESSION_SECRET=change-me-in-production  # Required in production

# AI Configuration (OpenAI provider with optional gateway)
AI_API_KEY=                     # Your OpenAI or gateway API key
AI_MODEL=gpt-4o                 # Model identifier
AI_BASE_URL=                    # Optional: custom gateway URL

# Email
EMAIL_PROVIDER=dev              # dev (console) | smtp (real sending)
EMAIL_FROM=no-reply@jobhoo.local
SMTP_HOST=                      # Only for SMTP mode
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
```

### Database Setup

Migrations are **automatically applied** by Docker Compose. For local setup:

```bash
# Run migrations manually
cd internal/database/migrations
bash init-migrations.sh

# Verify tables
psql -U jobhoo -d jobhoo -c "\dt"
```

### Email Configuration

**Development (Default):**
```env
EMAIL_PROVIDER=dev
# Emails logged to console
```

**Testing with Mailtrap:**
```env
EMAIL_PROVIDER=smtp
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_USER=your_username
SMTP_PASS=your_password
# Emails appear in Mailtrap inbox
```

**Production:**
```env
EMAIL_PROVIDER=smtp
SMTP_HOST=your-email-service.com
SMTP_PORT=587
SMTP_USER=your_user
SMTP_PASS=your_password
```

See `.env.example` for detailed setup instructions.

---

## 👨‍💻 Development

### Building & Running

```bash
# Build Go binary
go build -o jobhoo ./cmd/server

# Run directly
go run ./cmd/server

# With Docker Compose (includes database)
docker compose up

# Development with auto-reload (requires air)
air
```

### Database Migrations

Create new migration:
```bash
# Create pair of SQL files in internal/database/migrations/
# Example: 0014_new_feature.up.sql and 0014_new_feature.down.sql

# Migrations run automatically on container startup
# Or manually: cd internal/database/migrations && bash init-migrations.sh
```

### Testing Email Features

1. Set `EMAIL_PROVIDER=dev` in `.env` to see emails in console
2. Sign up new account → verify email
3. Click "Forgot Password" → reset link
4. Check console for email content

### Code Organization

- **Handlers are thin:** Parse request, call repo, render template
- **Repos own queries:** All database access via repository pattern
- **Models hold logic:** Business rules in models/types
- **Middleware is composable:** Small, single-purpose middleware functions
- **No globals:** Dependency injection throughout

### Adding a New Feature

1. **Create database migration** in `internal/database/migrations/`
2. **Create repository** in `internal/database/*_repo.go` with CRUD methods
3. **Create handler** in `internal/handlers/*_handler.go`
4. **Add routes** in `internal/router/router.go`
5. **Create templates** in `web/templates/pages/` or `components/`
6. **Add middleware** if needed in `internal/middleware/`

---

## 🏗 Architecture

### Design Patterns

**Repository Pattern**
```
Handler → Repository → Database
  ↑           ↑
  └─── Models (business logic)
```

**Middleware Stack**
```
Request → Logger → RealIP → Recoverer → Timeout → User → CSRF → Handler
```

**Dependency Injection**
```
main() creates all repos and passes to handlers.New()
No globals, easy to test with mocks
```

### Database Schema

13 migrations creating:
- `users` — User accounts with role
- `sessions` — Authentication sessions
- `companies` — Company profiles
- `candidate_profiles` — Candidate details
- `jobs` — Job listings
- `applications` — Job applications with status
- `saved_jobs` — User job bookmarks
- `email_tokens` — Verification & password reset tokens
- `email_logs` — Audit trail for compliance

**All tables have:**
- Created/updated timestamps
- Proper indexes for common queries
- Foreign key constraints with cascade options
- UUID primary keys

### Security Features

1. **Authentication:** Session-based, bcrypt password hashing
2. **CSRF Protection:** Double-submit tokens on all forms
3. **Rate Limiting:** 5 attempts/15 min on login/signup
4. **SQL Injection:** Parameterized queries (pgx)
5. **XSS Protection:** HTML template auto-escaping
6. **Email Verification:** Single-use tokens with TTL
7. **Password Reset:** Time-limited reset tokens (2 hours)
8. **Audit Logging:** All emails logged to database

---

## 📚 Documentation

### Project Documents
- **[DOC-PRODUCT-OVERVIEW.md](DOC-PRODUCT-OVERVIEW.md)** — What is JOBHOO? (non-technical)
- **[DOC-DEVELOPMENT-GUIDE.md](DOC-DEVELOPMENT-GUIDE.md)** — Developer setup, architecture, AI configuration, and design system
- **[DOC-DEVELOPMENT-PHASE.md](DOC-DEVELOPMENT-PHASE.md)** — Current development status per phase
- **[DOC-AUDIT-REPORT.md](DOC-AUDIT-REPORT.md)** — Technical, security, performance, and UI/UX audit

### API Endpoints

**Public Routes**
- `GET /` — Home page
- `GET /signup` — Signup form
- `POST /signup` — Create account
- `GET /login` — Login form
- `POST /login` — Authenticate user
- `GET /jobs` — Browse jobs
- `GET /jobs/search` — Search with filters
- `GET /jobs/{id}` — Job detail

**Candidate Routes** (`/dashboard/candidate`)
- View recommendations
- Apply to jobs
- Save jobs
- View profile
- Update resume

**Recruiter Routes** (`/dashboard/recruiter`)
- Post jobs
- View applicants (ATS board)
- Manage company profile
- Close/archive jobs

**Admin Routes** (`/dashboard/admin`)
- Company approval queue
- User management
- Freeze accounts
- Modal detail views

---

## 🤝 Contributing

### Development Workflow

1. **Create feature branch**
   ```bash
   git checkout -b feature/new-feature
   ```

2. **Follow Go conventions**
   - Clear package names
   - Exported functions start with uppercase
   - Error handling with context
   - Comments for exported functions

3. **Test before commit**
   ```bash
   go test ./...
   go build ./cmd/server
   ```

4. **Commit with clear messages**
   ```bash
   git commit -m "feat: add new feature description"
   ```

5. **Push and create pull request**

### Code Review Checklist

- [ ] Follows Go conventions
- [ ] Proper error handling
- [ ] No SQL injection (parameterized queries)
- [ ] Database migrations included (if schema changes)
- [ ] Tests included (if applicable)
- [ ] Documentation updated
- [ ] No hardcoded secrets/credentials

### Reporting Issues

- **Bug Report:** Include steps to reproduce, expected vs actual behavior
- **Feature Request:** Describe use case and proposed solution
- **Security Issue:** Email security@jobhoo.com (don't open public issue)

---

## 🔒 Security & Best Practices

### Running in Production

1. **Environment Variables**
   ```bash
   # Generate strong secret
   openssl rand -base64 32
   
   # Set in production:
   SESSION_SECRET=<generated_value>
   EMAIL_PROVIDER=smtp
   ```

2. **Database**
   - Use PostgreSQL 16+ with SSL
   - Strong password for database user
   - Regular backups
   - Monitor slow queries

3. **Email Service**
   - Use production email provider (AWS SES, SendGrid, etc.)
   - Configure SPF/DKIM/DMARC
   - Monitor bounce rates

4. **Application**
   - Deploy behind HTTPS reverse proxy (nginx, cloudflare)
   - Set rate limiting appropriately for traffic
   - Monitor error logs
   - Enable structured logging

### Security Features Implemented

✅ CSRF protection on all forms  
✅ Rate limiting on auth endpoints (5 attempts/15 min)  
✅ Email verification (48-hour tokens)  
✅ Password reset (2-hour tokens)  
✅ Bcrypt password hashing  
✅ Session-based authentication  
✅ Email audit logging  
✅ Parameterized SQL queries  
✅ HTML template auto-escaping  
✅ Role-based access control  

---

## 📊 Code Quality

See [DOC-AUDIT-REPORT.md](DOC-AUDIT-REPORT.md) for the detailed technical, security, and performance audit.

---

## 🗺 Roadmap

### Phase 1: MVP ✅
- [x] User authentication (candidates & recruiters)
- [x] Job posting & application
- [x] Basic search & filtering
- [x] Admin approval workflow
- [x] Email notifications

### Phase 2: Enhanced Features 🚀
- [ ] Unit & integration tests
- [ ] API documentation (Swagger/OpenAPI)
- [ ] Advanced resume parsing
- [ ] Recommendation engine refinement
- [ ] Mobile app

### Phase 3: Scale & Optimize 📈
- [ ] Caching layer (Redis)
- [ ] Search engine (Elasticsearch)
- [ ] Background jobs (task queue)
- [ ] Analytics dashboard
- [ ] Performance monitoring

---

## 📄 License

JOBHOO is licensed under the **MIT License**. See [LICENSE](LICENSE) file for details.

---

## 📞 Support & Contact

- 📧 **Email:** support@jobhoo.com
- 🐛 **Issues:** [GitHub Issues](https://github.com/yourusername/jobhoo/issues)
- 💬 **Discussions:** [GitHub Discussions](https://github.com/yourusername/jobhoo/discussions)
- 📖 **Documentation:** See `DOC-*.md` files

---

## 🙏 Acknowledgments

Built with:
- **Go** — Fast, reliable, minimal dependencies
- **PostgreSQL** — Powerful, open-source database
- **Chi** — Lightweight HTTP router
- **HTMX** — Hypermedia as the engine of application state
- **Dark theme** — Modern, professional aesthetic

---

<div align="center">

**Made with ❤️ by the JOBHOO team**

[⬆ Back to top](#-jobhoo--smart-recruitment-platform)

</div>
