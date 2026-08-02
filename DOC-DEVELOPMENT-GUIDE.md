# JOBHOO — Complete Setup, Development Guide & Troubleshooting

Recruitment platform menghubungkan kandidat dan recruiter dengan AI sebagai asisten pengambilan keputusan — bukan pengganti.

**Status:** Platform berfungsi end-to-end dengan 100% fitur core tersedia.

---

## Table of Contents

1. [Overview](#overview)
2. [Stack & Architecture](#stack--architecture)
3. [Project Structure](#project-structure)
4. [Installation Guide (Windows)](#installation-guide-windows)
5. [Features Checklist](#features-checklist)
6. [What Makes JOBHOO Different](#what-makes-jobhoo-different)
7. [Running Locally](#running-locally)
8. [Troubleshooting & Common Issues](#troubleshooting--common-issues)
9. [Known Gaps Before Production](#known-gaps-before-production)

---

## Overview

JOBHOO adalah platform rekrutmen fokus (bukan HRIS, bukan social network) yang menghubungkan kandidat dengan role yang tepat dan membantu recruiter menemukan kandidat terbaik.

**Tech Stack:**
- **Backend:** Go (net/http + chi router)
- **Frontend:** Server-rendered Go templates + HTMX
- **Database:** PostgreSQL
- **AI:** Provider interface modular (mock/Anthropic/OpenAI)
- **Deployment:** Docker + Docker Compose

**Brand Identity:**
- **Warna:** Dark theme — navy #1F2747 dominan, orange #FF7A00 aksen saja
- **Tipografi:** Lexend
- **Geometri:** Rounded corners mengikuti bentuk logo

---

## Stack & Architecture

### Backend: Go + PostgreSQL + HTMX

Setiap halaman di-render server-side via `html/template`. HTMX menangani interaksi dinamis tanpa full page reload — filter kategori, pagination, board Kanban ATS, semuanya via AJAX.

Database schema di-migrate otomatis saat container pertama kali spin up, lewat `docker-entrypoint-initdb.d` mounted ke folder migration `.up.sql` files.

### AI Layer: Provider Interface

`internal/ai.Provider` adalah interface tunggal untuk semua AI features. JOBHOO menggunakan **OpenAI provider secara eksklusif**.

```go
type Provider interface {
  RankCandidates(ctx, profile, jobDesc) Ranking
  ExplainMatch(ctx, profile, jobDesc) string
  RecommendJobs(ctx, profileSkills, allJobs) []Job
  ImproveSuggestions(ctx, resumeText) []string
}
```

Configure OpenAI via `.env`:
- `AI_API_KEY` - Your OpenAI or gateway API key
- `AI_MODEL` - Model identifier (e.g., gpt-4o, openai/gpt-5-nano)
- `AI_BASE_URL` - Custom gateway (optional, defaults to OpenAI official API)

### Design System: Token-Driven

Semua warna, spacing, radius, tipe hidup di `web/static/css/tokens.css`. Komponen di `components.css` hanya referensi token — tidak ada magic number di luar satu file.

| Token | Value | Use |
|---|---|---|
| `--jh-navy-700` | `#1F2747` | Dominan struktur |
| `--jh-orange-500` | `#FF7A00` | Satu-satunya aksen |
| `--jh-white` | `#FFFFFF` | Typography |
| `--jh-black` | `#0A0C16` | High contrast |

---

## Project Structure

```
cmd/
  server/main.go          Entrypoint app (config → db → ai → router → server)
  seed/main.go            CLI untuk seed demo data (10 company, 100 job)

internal/
  config/                 Environment config (1 struct, 1 file)
  database/               pgx pool + repositories (SQL hanya di sini)
    migrations/           Schema & seeder SQL (*.up.sql & *.down.sql)
    applications_repo.go  
    candidate_profiles_repo.go
    companies_repo.go
    jobs_repo.go
    saved_jobs_repo.go
    sessions_repo.go
    users_repo.go
  models/                 Shared domain types (User, Job, Company, etc.)
  ai/                     OpenAI AI provider (exclusive)
    provider.go           Provider interface definition
    openai.go             OpenAI API implementation (all features)
    openai_test.go        Comprehensive unit & integration tests
    prompts.go            System prompts for AI consistency
  handlers/               HTTP handlers (thin: parse → repo call → render)
    pages.go              Homepage, jobs listing, job detail
    auth.go               Signup, login, logout
    recruiter.go          Post job, recruiter dashboard, ATS board
    profile.go            Candidate profile
    dashboard.go          Admin dashboard
    render.go             Template rendering helper
  router/                 Full route table (1 file)
  middleware/             Auth, CSRF, logging
  
web/
  templates/
    layouts/base.html     Nav, footer, layout utama
    components/           Reusable partials (job-card, etc.)
    pages/                1 template per route
  static/
    css/tokens.css        Design tokens
    css/components.css    Component styles
    img/                  Logo assets

docker-compose.yml        Service definition (app + db)
Dockerfile               Multi-stage build (Go compile + Alpine run)
go.mod, go.sum           Dependencies
.env.example             Template environment config
```

---

## Installation Guide (Windows)

### Prerequisites

**Docker Desktop for Windows**
- Download: https://www.docker.com/products/docker-desktop/
- Pilih "Use WSL 2 instead of Hyper-V" saat install (default biasanya)
- Docker akan minta restart jika WSL2 belum aktif — terima saja

**Catatan:** Tidak perlu install Go, PostgreSQL, atau tool lain — semuanya dalam container.

### Step-by-Step

**1. Clone/extract project:**
```powershell
cd C:\Users\YourName\jobhoo
```

**2. Setup environment:**
```powershell
copy .env.example .env
```

Dalam `.env`, perhatikan:
- `PORT=8070` — port aplikasi
- `DATABASE_URL=postgres://jobhoo:jobhoo_dev_password@db:5432/jobhoo?sslmode=disable` — **PENTING: `@db` bukan `@localhost`** (untuk Docker networking)
- `AI_PROVIDER=mock` — cukup untuk testing

**3. Pastikan Docker Desktop sudah running** — cek ikon paus di system tray

**4. Build & start:**
```powershell
docker compose up --build
```

Tunggu sampai muncul:
```
app-1  | 2026/07/28 03:06:01 JOBHOO listening on :8070 (env=development)
```

**5. Buka terminal baru, seed demo data:**
```powershell
docker compose run --rm -e DATABASE_URL="postgres://jobhoo:jobhoo_dev_password@db:5432/jobhoo?sslmode=disable" app ./jobhoo-seed
```

Expected output:
```
clearing previous demo data...
creating 10 recruiter accounts + companies...
creating 100 jobs across 5 categories...
done: 10 companies and 100 jobs across 5 categories seeded.
```

**6. Open browser:**
```
http://localhost:8070
```

Halaman harus menampilkan:
- Hero section
- 6 latest job opportunities cards dari database
- Menu "Browse Jobs" & "Post a Job" berfungsi

### Demo Accounts

**Admin:**
- Email: `admin@jobhoo.demo`
- Password: `demo-password-123`

**Recruiter (10 akun):**
- Email: `recruiter1@jobhoo.demo` s/d `recruiter10@jobhoo.demo`
- Password: `demo-password-123`
- `recruiter10` sengaja dibiarkan status company-nya **pending** untuk testing approval queue

**Candidate:**
- Signup baru lewat `/signup`

### Menghentikan

```powershell
# Tekan Ctrl+C di terminal yang menjalankan docker compose up
docker compose down
```

### Reset Database Penuh

```powershell
docker compose down -v
docker compose up --build
docker compose run --rm -e DATABASE_URL="postgres://jobhoo:jobhoo_dev_password@db:5432/jobhoo?sslmode=disable" app ./jobhoo-seed
```

---

## Features Checklist

### Candidate Features ✅
- ✅ Browse & cari lowongan (search, filter kategori, filter lokasi/negara, pagination)
- ✅ Lihat detail lowongan lengkap (termasuk salary dengan pemisah ribuan + mata uang)
- ✅ Apply ke lowongan dengan cover note
- ✅ Simpan/bookmark lowongan
- ✅ Edit profil (headline, resume file upload, resume text, skills chip-input, location)
- ✅ **Resume wajib** diupload saat registrasi atau melengkapi profil
- ✅ AI: saran perbaikan resume
- ✅ AI: rekomendasi lowongan berdasarkan skill
- ✅ Dashboard: riwayat lamaran + status, lowongan tersimpan

### Recruiter Features ✅
- ✅ Registrasi company, menunggu persetujuan admin sebelum bisa posting
- ✅ Upload logo company (live preview sebelum simpan)
- ✅ Wajib lengkapi profil company (industry + description) sebelum bisa posting job
- ✅ Post lowongan baru (dengan chip-input skill, scheduling opens_at/closes_at, HTML description)
- ✅ Kategori lowongan bebas: pilih dari 5 bawaan atau ketik custom
- ✅ Lokasi: country + state/province (9 negara ASEAN + Oceania), auto-fill currency
- ✅ Salary dengan currency sesuai negara (IDR, SGD, MYR, dll.)
- ✅ Dashboard (Job Management): daftar job + status + tanggal buka/tutup + jumlah pelamar
- ✅ Action dropdown per job: Pipeline, Edit, Close/Reopen, Archive
- ✅ ATS Board Kanban: Applied → Screening → Interview → Offer → Hired
- ✅ AI: ranking kandidat otomatis (saran saja, tidak mengubah stage)
- ✅ Company public page (auto-redirect dari menu "Public Page" di nav)

### Admin Features ✅
- ✅ Dashboard: statistik platform (total user, company, job, aplikasi)
- ✅ Company Approval Queue: approve/reject company baru dengan alasan penolakan
- ✅ Badge jumlah pending di tombol approval

### Security ✅
- ✅ Signup/login/logout dengan bcrypt password hashing
- ✅ Sesi tersimpan di DB (bisa di-revoke), bukan JWT stateless
- ✅ Proteksi route berbasis role (kandidat/recruiter/admin terpisah)
- ✅ CSRF protection di semua form
- ✅ Recruiter tidak bisa akses pipeline company lain (ownership check)
- ✅ Token session di-hash sebelum simpan DB
- ✅ File upload: MIME type validation dari header file (bukan ekstensi)
- ✅ HTML descriptions: sanitized via bluemonday (allowlist tag)

### Data & Infrastructure ✅
- ✅ PostgreSQL dengan schema terstruktur (migrations 0001–0009)
- ✅ Seeder demo data (1 admin, 10 company, 100 job ASEAN lokasi)
- ✅ Docker + Docker Compose
- ✅ Multi-stage Dockerfile (build di golang, run di alpine)
- ✅ Docker bind-mount untuk uploaded files (`web/static/uploads/`) agar persisten di host

---

## What Makes JOBHOO Different

### 1. AI Explains, Not Just Filters
Kebanyakan job portal hanya keyword-match. JOBHOO dirancang agar AI memberi **alasan** di balik tiap skor — "cocok karena X, kurang di Y" — baik untuk recruiter menilai kandidat maupun kandidat menilai job.

### 2. AI Never Takes Decisions
Ranking AI di ATS board JOBHOO **tidak pernah** mengubah stage kandidat atau menyembunyikan siapa pun — murni saran di atas data yang tetap terlihat utuh. Banyak platform lain membiarkan algoritma auto-reject tanpa recruiter sadar.

### 3. Provider-Agnostic Architecture
Kebanyakan startup menanam satu vendor AI di kode inti. JOBHOO pisah lewat satu interface — ganti dari Anthropic ke OpenAI tanpa ubah satu baris kode aplikasi.

### 4. Narrow Focus (Sengaja)
Brief awal eksplisit menolak jadi "LinkedIn kedua" — tidak ada feed sosial, tidak ada follow/connect. Platform besar tambah fitur sosial untuk retensi; JOBHOO sengaja tidak, supaya pengalaman tetap "cari kerja, kelola pelamar".

### 5. ATS Built-In, Bukan Add-On
Banyak job board hanya tempat posting — recruiter tetap pakai ATS terpisah (Greenhouse, Lever). JOBHOO menyatukan keduanya: satu platform, satu login, satu alur kerja.

### 6. Transparansi Status Lamaran
Banyak platform membuat kandidat "menghilang" setelah apply. Dashboard kandidat JOBHOO menampilkan stage lamaran real-time (Applied/Screening/Interview/dsb).

---

## Running Locally

### With Docker (Recommended)

```bash
cp .env.example .env
docker compose up --build              # Build & start app + Postgres
docker compose run --rm app ./jobhoo-seed    # Seed demo data
```

Open: http://localhost:8070

**Reseed tanpa recreate DB:**
```bash
docker compose run --rm app ./jobhoo-seed
```

**Wipe DB penuh:**
```bash
docker compose down -v
docker compose up --build
docker compose run --rm app ./jobhoo-seed
```

### Without Docker (Requires Local Postgres)

```bash
# Create database & apply migrations manually
createdb jobhoo
psql jobhoo < internal/database/migrations/0001_init.up.sql
psql jobhoo < internal/database/migrations/0003_sessions.up.sql
psql jobhoo < internal/database/migrations/0004_job_category.up.sql
psql jobhoo < internal/database/migrations/0005_fix_cascade.up.sql

# Run app
go run ./cmd/server

# Seed demo data (di terminal baru)
go run ./cmd/seed
```

---

## Troubleshooting & Common Issues

### Issue 1: Homepage Kosong (Hanya Navbar)

**Gejala:**
- Homepage load tapi hanya navbar, konten kosong
- "could not load jobs" error di browser

**Root Cause:**
Template rendering error (biasanya nil pointer evaluating `.CurrentUser.Role` saat `.CurrentUser` adalah nil).

**Solusi:**
1. Cek `web/templates/layouts/base.html` — gunakan nested `if` untuk nil-safe access:
   ```go
   {{if not .CurrentUser}}
     // Show public links
   {{else if eq .CurrentUser.Role "recruiter"}}
     // Show recruiter links
   {{else if eq .CurrentUser.Role "candidate"}}
     // Show candidate links
   {{end}}
   ```
2. Rebuild image: `docker compose build --no-cache && docker compose up`

### Issue 2: Seeder Hangs atau Tidak Eksekusi

**Gejala:**
- `docker compose run --rm app ./jobhoo-seed` seperti tidak ada output
- Atau: error "can't scan into dest[18]: cannot scan NULL into *string"

**Root Cause:**
- Database URL `@localhost:5432` tidak valid di dalam Docker container (inside container, nama service adalah `db`)
- Field struct tidak nullable tapi database return NULL

**Solusi:**
1. **Pass DATABASE_URL dengan `@db`:**
   ```powershell
   docker compose run --rm -e DATABASE_URL="postgres://jobhoo:jobhoo_dev_password@db:5432/jobhoo?sslmode=disable" app ./jobhoo-seed
   ```

2. **Nullable field di models jika DB return NULL:**
   ```go
   type Job struct {
     CompanyLogoURL *string  // Pointer, bukan string langsung
   }
   ```

### Issue 3: Foreign Key Constraint Error

**Gejala:**
```
ERROR: update or delete on table "users" violates foreign key constraint "jobs_created_by_fkey"
```

**Root Cause:**
Migration lupa menambah `ON DELETE CASCADE` ke foreign key.

**Solusi:**
1. Buat migration baru `0005_fix_cascade.up.sql`:
   ```sql
   ALTER TABLE jobs 
     DROP CONSTRAINT jobs_created_by_fkey,
     ADD CONSTRAINT jobs_created_by_fkey 
       FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;
   ```

2. Rebuild: `docker compose down -v && docker compose up --build`

### Issue 4: Port 8070 Sudah Dipakai

**Gejala:**
```
Bind for 0.0.0.0:8070 failed: port is already allocated
```

**Solusi:**
1. **Option A:** Ubah `.env`:
   ```
   PORT=8071
   ```
   
2. **Option B:** Ubah `docker-compose.yml`:
   ```yaml
   ports:
     - "8071:8070"
   ```

3. **Option C:** Kill proses yang pakai port:
   ```powershell
   netstat -ano | findstr :8070
   taskkill /PID <PID> /F
   ```

### Issue 5: Docker Image Cache (Kode Lama Masih Jalan)

**Gejala:**
- Ubah kode tapi aplikasi masih jalankan versi lama

**Solusi:**
```powershell
docker compose build --no-cache
docker compose down
docker compose up
```

### Issue 6: Database Sudah Ada, Migrasi Tidak Jalan

**Gejala:**
- `docker compose down` tapi volume `.down.sql` files dari migrasi lama tetap dijalankan

**Solusi:**
```powershell
docker compose down -v   # -v menghapus volume data
docker compose up --build
```

### Issue 7: Docker Desktop Gagal Start (WSL2 Error)

**Gejala:**
```
Docker Desktop requires WSL 2 backend
```

**Solusi:**
1. Buka PowerShell sebagai Administrator
2. Jalankan: `wsl --install`
3. Restart komputer
4. Coba Docker Desktop lagi

---

## Known Gaps Before Production

### No Email
Tidak ada email verification, password reset, atau notifikasi saat stage lamaran berubah.

### No Rate Limiting
Login/signup/apply tidak ada rate limiting. Perlu ditambah sebelum public launch untuk mencegah credential-stuffing dan spam.

### CSRF Baseline Only
Standard double-submit-cookie, bukan per-session token.

### File Storage Ephemeral in Production
Di dev, uploaded files disimpan di `web/static/uploads/` via Docker bind-mount. Di production, perlu migrasi ke object storage (S3-compatible) — lokasi path sudah di-abstract, tinggal swap handler.

### No Accessibility Pass
Keyboard focus & semantic markup sudah dari design system, tapi belum ada full screen-reader audit.
- Semua 13 page template di-parse & execute dengan mock data shaped seperti struct asli

**Sebelum deploy, jalankan di machine dengan internet normal:**
```bash
go build ./...
go vet ./...
go test ./...
```

---

## Quick Reference

| Task | Command |
|---|---|
| Start app + DB | `docker compose up --build` |
| Seed demo data | `docker compose run --rm -e DATABASE_URL="postgres://jobhoo:jobhoo_dev_password@db:5432/jobhoo?sslmode=disable" app ./jobhoo-seed` |
| Stop | `docker compose down` |
| Reset DB | `docker compose down -v && docker compose up --build` |
| View logs | `docker compose logs app` |
| Shell ke DB | `docker compose exec db psql -U jobhoo -d jobhoo` |
| Rebuild image | `docker compose build --no-cache` |

---

**Last Updated:** 2026-07-28  
**Status:** End-to-end platform berfungsi ✅
