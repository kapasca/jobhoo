# JOBHOO Development Guide

Panduan teknis lengkap untuk menjalankan, mengembangkan, dan memahami arsitektur JOBHOO. Untuk gambaran produk non-teknis, lihat [DOC-PRODUCT-OVERVIEW.md](DOC-PRODUCT-OVERVIEW.md). Untuk status pengerjaan per fitur, lihat [DOC-DEVELOPMENT-PHASE.md](DOC-DEVELOPMENT-PHASE.md). Untuk temuan audit teknis/keamanan, lihat [DOC-AUDIT-REPORT.md](DOC-AUDIT-REPORT.md).

## 1. Ringkasan

JOBHOO adalah platform rekrutmen fokus (bukan HRIS, bukan jejaring sosial) yang menghubungkan kandidat dengan role yang tepat dan membantu recruiter menemukan kandidat terbaik. Platform berfungsi end-to-end dengan fitur inti (pencarian, lamar, ATS, verifikasi perusahaan, ranking AI) tersedia.

## 2. Tech Stack

1. Backend: Go, Chi v5 router, PostgreSQL dengan driver pgx v5.
2. Frontend: Go `html/template` (server-rendered) + HTMX untuk interaksi tanpa full page reload.
3. AI: Provider tunggal berbasis OpenAI (lihat Bagian 6).
4. Email: Dev logger dan SMTP sender, dengan audit logging.
5. Deployment: Docker + Docker Compose, multi-stage Dockerfile (build di image Go, run di Alpine).
6. Autentikasi: Sesi tersimpan di database (bukan JWT), token di-hash (SHA256) sebelum disimpan, password di-hash dengan bcrypt.

Dependensi langsung (`go.mod`): `chi/v5`, `pgx/v5`, `godotenv`, `bluemonday` (sanitasi HTML), `golang.org/x/crypto` (bcrypt).

## 3. Struktur Proyek

```
cmd/
  server/main.go          Entrypoint app (config -> db -> ai -> router -> server)
  seed/                   CLI untuk seed demo data

internal/
  config/                 Environment config (satu struct, satu file)
  database/                pgx pool + repositories (SQL hanya di sini)
    migrations/            Schema SQL (*.up.sql & *.down.sql), versi bertambah seiring waktu
    *_repo.go              Satu repository per entitas (users, jobs, companies, applications, dst.)
  models/                  Domain types (User, Job, Company, dll.)
  ai/                      Lapisan AI (provider interface + implementasi OpenAI)
    provider.go            Definisi interface Provider
    openai.go               Implementasi OpenAI (semua metode)
    openai_test.go          Unit & integration test
    prompts.go              System prompt bersama
    resume.go                Struktur hasil ekstraksi resume + parsing DOCX manual
  handlers/                HTTP handler (thin: parse request -> panggil repo -> render template)
  middleware/               Auth, CSRF, rate limiting
  router/                  Tabel routing lengkap
  auth/                    Utilitas hashing password & token
  email/                   Email sender (dev/smtp) + tipe pesan

web/
  templates/
    layouts/               Layout dasar (nav, footer)
    components/            Partial yang dapat digunakan ulang (job-card, modal admin, dll.)
    pages/                 Satu template per halaman
  static/
    css/                   Design tokens dan komponen CSS
    js/                    Script sisi klien (chips, confirm, ats-board)
    img/                   Aset logo
    uploads/               File yang diunggah pengguna (mounted volume di Docker)

docker-compose.yml        Definisi service (app + db)
Dockerfile                Multi-stage build (Go compile + Alpine run)
go.mod, go.sum             Dependency management
.env.example               Template konfigurasi environment
Makefile                   Perintah pengembangan singkat (run, build, up, down, seed, fmt, tidy)
```

## 4. Instalasi & Menjalankan Secara Lokal

### 4.1 Prasyarat

Docker Desktop (disarankan, mencakup WSL2 di Windows). Tidak perlu instalasi Go atau PostgreSQL terpisah karena semuanya berjalan dalam container.

### 4.2 Langkah Instalasi (Docker)

```powershell
cp .env.example .env
docker compose up --build
```

Tunggu hingga muncul log `JOBHOO listening on :8070 (env=development)`, lalu seed data demo di terminal baru:

```powershell
docker compose run --rm -e DATABASE_URL="postgres://jobhoo:jobhoo_dev_password@db:5432/jobhoo?sslmode=disable" app ./jobhoo-seed
```

Buka `http://localhost:8070`.

Catatan penting: di dalam Docker, gunakan `DATABASE_URL` dengan host `@db` (nama service), bukan `@localhost`.

### 4.3 Akun Demo

| Peran | Email | Password |
|---|---|---|
| Admin | admin@jobhoo.demo | demo-password-123 |
| Recruiter (1-10) | recruiter1@jobhoo.demo s/d recruiter10@jobhoo.demo | demo-password-123 |
| Candidate | Daftar baru lewat `/signup` | - |

`recruiter10` sengaja dibiarkan berstatus company pending untuk menguji antrian approval.

### 4.4 Menghentikan & Reset

```powershell
docker compose down            # Hentikan
docker compose down -v         # Hentikan + hapus volume data (reset penuh)
docker compose up --build      # Jalankan ulang
```

### 4.5 Tanpa Docker (Perlu PostgreSQL Lokal)

```bash
createdb jobhoo
cd internal/database/migrations && bash init-migrations.sh
go run ./cmd/server
go run ./cmd/seed          # Di terminal baru
```

## 5. Konfigurasi Environment Variable

Lihat `.env.example` untuk template lengkap. Ringkasan variabel:

| Variabel | Wajib | Keterangan |
|---|---|---|
| `APP_ENV` | Tidak | `development` (default) atau `production` |
| `PORT` | Tidak | Port HTTP server, default `8080` (dev lokal biasanya `8070`) |
| `DATABASE_URL` | Ya | Connection string PostgreSQL |
| `SESSION_SECRET` | Ya di production | Signing secret cookie sesi |
| `AI_API_KEY` | Ya untuk fitur AI | API key OpenAI atau gateway kompatibel |
| `AI_MODEL` | Tidak | Model identifier, contoh `openai/gpt-5-nano` |
| `AI_BASE_URL` | Tidak | URL gateway kustom; kosong berarti pakai OpenAI resmi |
| `AI_VISION_MODEL` | Tidak | Override model khusus untuk parsing resume via vision (PDF/gambar) |
| `EMAIL_PROVIDER` | Tidak | `dev` (log ke konsol, default) atau `smtp` |
| `EMAIL_FROM` | Tidak | Alamat pengirim email |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS` | Hanya jika `EMAIL_PROVIDER=smtp` | Kredensial SMTP |

Catatan kompatibilitas: variabel lama `MAIL_HOST`/`MAIL_PORT`/`MAIL_USER`/`MAIL_PASS` (gaya Mailtrap) masih didukung sebagai fallback jika `SMTP_*` kosong.

## 6. Lapisan AI (Provider OpenAI)

JOBHOO menggunakan provider OpenAI secara eksklusif untuk semua fitur AI. Tidak ada pencabangan multi-provider; konfigurasi hanya lewat environment variable untuk memilih model dan gateway.

### 6.1 Interface Provider

Interface `ai.Provider` (di `internal/ai/provider.go`) adalah satu-satunya seam antara JOBHOO dan backend AI. Handler tidak pernah mengimpor SDK vendor secara langsung.

```go
type Provider interface {
    RankCandidates(ctx, job, candidates) ([]CandidateRanking, error)
    ExplainMatch(ctx, job, candidate) (MatchExplanation, error)
    RecommendJobs(ctx, candidate, jobs) ([]JobRecommendation, error)
    ExtractResumeText(ctx, rawText string) (ResumeExtraction, error)
    ExtractResumeFile(ctx, fileData []byte, mediaType string) (ResumeExtraction, error)
    Name() string
}
```

Poin penting:

1. Output AI selalu bersifat advisory. Tidak ada metode yang mengambil keputusan hiring; semua hasil adalah saran untuk ditinjau manusia.
2. Resume (hasil ekstraksi terstruktur) adalah sumber utama untuk semua pencocokan AI. Field skill di profil kandidat bersifat sekunder, hanya sebagai cross-check tambahan.
3. `RankCandidates` tidak pernah menyembunyikan atau menghapus kandidat dari hasil; hanya memberi skor dan urutan.

### 6.2 Konfigurasi

```bash
AI_API_KEY=sk-xxx...
AI_MODEL=openai/gpt-5-nano
AI_VISION_MODEL=openai/gpt-5-nano   # opsional, override model untuk parsing PDF/gambar
AI_BASE_URL=https://api.maiarouter.ai/v1   # opsional, kosongkan untuk OpenAI resmi
```

Jika `AI_BASE_URL` kosong, provider otomatis menggunakan `https://api.openai.com/v1`.

### 6.3 Cara Kerja Gateway

1. Provider membaca `AI_BASE_URL` dari environment saat inisialisasi.
2. Jika tidak diset, default ke endpoint resmi OpenAI.
3. Trailing slash pada base URL dibersihkan otomatis.
4. Model dipilih dari `AI_MODEL`; endpoint yang dipanggil adalah `{baseURL}/chat/completions`.

Gateway kustom yang sudah diuji: Maia Router (`https://api.maiarouter.ai/v1`), dengan model `openai/gpt-5-nano`.

Catatan implementasi untuk model reasoning (keluarga `gpt-5*`, `o1`/`o3`/`o4`): model ini butuh parameter `max_completion_tokens` (bukan `max_tokens`) dan menolak `temperature`/`top_p` non-default. Provider mendeteksi ini otomatis lewat helper `isReasoningModel`.

### 6.4 Karakteristik Performa

| Operasi | Estimasi durasi |
|---|---|
| Ekstraksi/analisis resume (teks) | 2-3 detik |
| Ekstraksi resume via vision (PDF/gambar) | 8-10 detik |
| Ranking kandidat | 3-5 detik per job |
| Rekomendasi lowongan | 2-4 detik per batch |
| Timeout permintaan | 60 detik (server HTTP timeout 90 detik untuk mengakomodasi ini) |

### 6.5 Penanganan Error

Provider mengembalikan error deskriptif untuk kondisi: API key tidak valid, model tidak tersedia untuk key tersebut, error jaringan ke gateway, respons JSON tidak valid, atau timeout context.

## 7. Sistem Desain

Standar visual resmi JOBHOO. Semua komponen, halaman, dan aset baru mengikuti panduan ini agar identitas platform konsisten. Referensi implementasi: `web/static/css/tokens.css` (token) dan `web/static/css/components.css` (komponen, hanya mereferensikan token, tanpa nilai hex/pixel langsung).

### 7.1 Kepribadian Merek

JOBHOO adalah platform rekrutmen yang profesional namun tidak kaku:

1. Percaya diri - tampilan gelap dan bersih, bukan abu-abu atau putih seperti kebanyakan job board.
2. Fokus - tidak ada elemen dekoratif berlebihan; setiap elemen punya fungsi.
3. Hangat - sentuhan oranye yang terkontrol agar tidak terasa dingin atau korporat steril.
4. Modern - rounded corners, spacing besar, tipografi berat.

Kata kunci desain: dark, clean, orange-accented, rounded, purposeful.

### 7.2 Warna

JOBHOO menggunakan dark theme sepenuhnya; tidak ada mode terang.

| Nama | Hex | Penggunaan |
|---|---|---|
| Navy 700 (Brand Navy) | `#192132` | Background utama halaman |
| Navy 900 | `#0f1220` | Ujung bawah gradient atmosphere |
| Surface Card | `#1f2942` | Kartu, panel, form |
| Surface Inset | `#151d30` | Input field, area recessed |
| Orange 500 (Brand Orange) | `#d96600` | Aksen utama, CTA, active state |
| Orange 400 | `#e87d33` | Hover state tombol oranye, icon |
| Ink 100 | `#eef0f6` | Teks utama (body, label, heading) |
| Ink 300 | `#b7bdd1` | Teks sekunder (subtitle, hint) |
| Ink 500 | `#7d84a3` | Teks tersier (placeholder, caption) |
| Border | `#2a3549` | Garis batas komponen |
| Border Strong | `#3a4462` | Garis lebih menonjol (hover, focus) |

Aturan warna:

1. Oranye hanya untuk aksen, bukan background atau teks paragraf.
2. Tidak ada warna lain selain navy dan oranye, kecuali warna status (hijau untuk sukses, merah untuk error).
3. Background halaman selalu gradient: `linear-gradient(180deg, #192132 0%, #0f1220 100%)`.
4. Overlay/backdrop menggunakan `rgba(0, 0, 0, 0.4)` di atas background.

Warna status:

| Status | Warna |
|---|---|
| Sukses / Hired | `#045b25` (hijau gelap) |
| Error / Rejected | `#5b0404` (merah gelap) |
| Warning / Pending | `rgba(217, 102, 0, 0.08)` dengan border `#d96600` |
| Netral / Closed | Surface Inset dengan Ink 500 |

### 7.3 Tipografi

Typeface: Lexend (Google Fonts), satu-satunya typeface di seluruh platform. Fallback: `-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif`.

| Token | Rem | Pixel | Penggunaan |
|---|---|---|---|
| `text-xs` | 0.75rem | 12px | Badge, label kecil, hint |
| `text-sm` | 0.875rem | 14px | Tabel, metadata, nav link |
| `text-base` | 1rem | 16px | Teks body utama, form input |
| `text-lg` | 1.125rem | 18px | Subtitle, deskripsi penting |
| `text-xl` | 1.375rem | 22px | Heading seksi kecil (h3) |
| `text-2xl` | 1.75rem | 28px | Heading level 2 (h2) |
| `text-3xl` | 2.25rem | 36px | Heading besar (h2 hero) |
| `text-4xl` | 3rem | 48px | Heading halaman utama (h1) |

Line height: `leading-tight` (1.2) untuk heading, `leading-normal` (1.55) untuk body text, `leading-relaxed` (1.75) untuk deskripsi panjang.

Bobot font: 700 untuk heading dan label form, 600 untuk tombol dan nav aktif, 500 untuk nav inaktif dan metadata, 400 untuk body text dan placeholder.

### 7.4 Spacing

Sistem berbasis 8px (basis 0.25rem = 4px, kelipatan 2):

| Token | Rem | Pixel |
|---|---|---|
| `space-1` | 0.25rem | 4px |
| `space-2` | 0.5rem | 8px |
| `space-3` | 0.75rem | 12px |
| `space-4` | 1rem | 16px |
| `space-5` | 1.5rem | 24px |
| `space-6` | 2rem | 32px |
| `space-7` | 3rem | 48px |
| `space-8` | 4rem | 64px |
| `space-9` | 6rem | 96px |

Prinsip: elemen dalam satu komponen memakai `space-1` sampai `space-3`. Antar komponen memakai `space-4` sampai `space-6`. Antar section/halaman memakai `space-7` sampai `space-9`.

### 7.5 Border Radius

Rounded corners konsisten; tidak ada elemen kotak (0px) kecuali garis/divider. Semakin besar elemen, semakin besar radius-nya.

| Token | Nilai | Digunakan Pada |
|---|---|---|
| `radius-sm` | 10px | Badge, chip, tombol kecil, tag |
| `radius-md` | 16px | Input field, tombol utama, dropdown |
| `radius-lg` | 22px | Kartu (card), panel, modal |
| `radius-xl` | 28px | Modal besar, hero section |
| `radius-pill` | 999px | Badge pill, filter chip |

### 7.6 Elevasi & Bayangan

| Token | Nilai | Penggunaan |
|---|---|---|
| `shadow-sm` | `0 2px 8px rgba(10,12,22, 0.24)` | Kartu statis |
| `shadow-md` | `0 8px 24px rgba(10,12,22, 0.32)` | Modal, dropdown, floating element |
| `shadow-glow-orange` | `0 0 0 4px rgba(217,102,0, 0.16)` | Focus ring, input aktif |

### 7.7 Komponen Utama

**Tombol** - 3 varian: Primary (background oranye, teks putih, untuk aksi utama), Secondary (transparan, border strong, untuk aksi sekunder), Ghost (transparan, teks ink-300, untuk navigasi ringan). Ukuran: Default (padding `0.75rem 1.375rem`), Small (`0.5rem 1rem`), Large (`0.875rem 1.75rem`). Hover: Primary ke `#e87d33` dengan glow oranye; Secondary berubah jadi background oranye.

**Kartu** - Background `surface-card`, border 1px `border`, radius `radius-lg`, padding standar `space-5`. Job Card memiliki header (logo + nama company + waktu posting + bookmark), judul lowongan, dan body (logo besar + metadata: kategori, lokasi, tipe kerja, tipe kontrak).

**Badge dan Chip** - Radius `radius-pill`, padding `0.3rem 0.7rem`, font `text-xs` weight 600. Varian: Default (surface-inset/ink-300), Orange (aksen aktif), Green (Hired), Red (Rejected), Gray (Closed/Archived).

**Form dan Input** - Background `surface-inset`, border 1px `border`, radius `radius-md`, padding `0.75rem 1rem`. Focus: border oranye + `shadow-glow-orange`. Placeholder: `ink-500`.

**Page Header** - Background gradient `linear-gradient(to right, rgba(0,0,0,0.4), rgba(0,0,0,0.4), transparent)`, min-height 130px, judul `text-4xl` bold putih, subtitle `text-lg` ink-300.

**Navbar** - Sticky top, background `rgba(navy, 0.85)` + backdrop-blur 10px, border bawah 1px. Menu berbeda per peran: Kandidat (Browse Jobs, nama user, Dashboard), Recruiter (Job Management, Public Page, My Profile), Admin (Browse Jobs, Dashboard), Tamu (Browse Jobs, Post a Job).

**Modal** - Overlay `rgba(0,0,0,0.5)`, panel `surface-card` radius-lg, max-width 700px, max-height 90vh dengan scroll internal. Animasi masuk: slide dari atas + fade, durasi 300ms. Mobile: bottom-sheet (radius hanya di atas, lebar 100%).

### 7.8 Layout & Grid

| Class | Max-width | Penggunaan |
|---|---|---|
| `.container` | 1200px | Layout umum |
| `.container-max-sm` | 480px | Modal login, form singkat |
| `.container-max-md` | 520px | Form signup, company setup |
| `.container-max-lg` | 680px | Form post job, profile |
| `.container-max-xl` | 840px | Job detail, company detail kecil |
| `.container-max-2xl` | 900px | Company detail publik |

Grid: auto-fill `minmax(300px, 1fr)` untuk daftar job card, 2 kolom untuk pasangan field, 3 kolom untuk field currency/salary. Mobile (<= 640px): semua grid collapse ke 1 kolom.

Breakpoint: <= 640px mobile (grid 1 kolom, bottom-sheet modal), <= 768px tablet (hamburger menu), <= 1024px sidebar diperkecil.

### 7.9 Animasi & Motion

Easing standar `cubic-bezier(0.4, 0, 0.2, 1)`. Durasi: fast 120ms (hover, focus ring), normal 200ms (tombol, card hover, dropdown), modal enter 300ms, mobile drawer 220ms. Tidak ada animasi dekoratif (bounce/elastic); setiap motion memberi feedback visual yang jelas. `prefers-reduced-motion` mengurangi semua durasi menjadi `0.001ms`.

### 7.10 Ikonografi

SVG inline gaya Lucide/Heroicons (viewBox 24x24, `stroke="currentColor"`, `stroke-width="2"`). Digunakan pada metadata job card untuk seniority, kategori, employment type, work arrangement, lokasi, dan salary.

### 7.11 Scrollbar Kustom

Lebar 8px, track transparan, thumb `orange-500` (hover `orange-400`), border-radius `radius-sm`.

### 7.12 Batasan Desain

1. Jangan gunakan warna latar putih atau abu-abu terang.
2. Jangan tambah typeface selain Lexend.
3. Jangan gunakan oranye sebagai background area besar.
4. Jangan gunakan border-radius 0 pada komponen interaktif.
5. Jangan tambah shadow berwarna selain oranye glow dan navy shadow.
6. Jangan gunakan animasi bounce, elastic, atau dekoratif.
7. Jangan letakkan teks gelap di atas background navy (kontras terlalu rendah).
8. Jangan hard-code nilai hex/pixel langsung di CSS; selalu gunakan token dari `tokens.css`.

## 8. Troubleshooting

### 8.1 Homepage Kosong (Hanya Navbar Tampil)

Penyebab: error rendering template, biasanya nil pointer saat mengevaluasi `.CurrentUser.Role` ketika `.CurrentUser` adalah nil.

Solusi: gunakan nested `if` yang nil-safe di `web/templates/layouts/base.html`:

```go
{{if not .CurrentUser}}
  // Tampilkan menu publik
{{else if eq .CurrentUser.Role "recruiter"}}
  // Menu recruiter
{{else if eq .CurrentUser.Role "candidate"}}
  // Menu candidate
{{end}}
```

Rebuild: `docker compose build --no-cache && docker compose up`.

### 8.2 Seeder Hang atau Tidak Berjalan

Penyebab: `DATABASE_URL` dengan `@localhost` tidak valid di dalam container (nama service adalah `db`), atau field struct tidak nullable padahal database mengembalikan NULL.

Solusi: jalankan seeder dengan `DATABASE_URL` eksplisit menggunakan host `@db`:

```powershell
docker compose run --rm -e DATABASE_URL="postgres://jobhoo:jobhoo_dev_password@db:5432/jobhoo?sslmode=disable" app ./jobhoo-seed
```

Untuk kolom yang bisa NULL di database, gunakan pointer di struct Go (`*string`, bukan `string`).

### 8.3 Foreign Key Constraint Error

Penyebab: migration lupa menambahkan `ON DELETE CASCADE` pada foreign key.

Solusi: buat migration baru untuk memperbaiki constraint, contoh:

```sql
ALTER TABLE jobs
  DROP CONSTRAINT jobs_created_by_fkey,
  ADD CONSTRAINT jobs_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;
```

### 8.4 Port Sudah Dipakai

Ubah `PORT` di `.env`, atau ubah mapping port di `docker-compose.yml`, atau hentikan proses yang memakai port tersebut (`netstat -ano | findstr :8070` lalu `taskkill /PID <PID> /F` di Windows).

### 8.5 Docker Menjalankan Kode Lama

Rebuild tanpa cache: `docker compose build --no-cache` lalu `docker compose down` dan `docker compose up`.

### 8.6 Migrasi Tidak Berjalan Ulang

Migrasi hanya otomatis dijalankan pada volume database yang baru. Untuk memaksa migrasi ulang: `docker compose down -v` (menghapus volume) lalu `docker compose up --build`. Untuk menambahkan migration baru ke database dev yang sudah berjalan, jalankan manual, contoh:

```powershell
Get-Content internal/database/migrations/0016_new_feature.up.sql | docker exec -i jobhoo-db-1 psql -U jobhoo -d jobhoo
```

### 8.7 Docker Desktop Gagal Start (WSL2)

Buka PowerShell sebagai Administrator, jalankan `wsl --install`, restart komputer, lalu coba Docker Desktop lagi.

## 9. Referensi Cepat

| Tugas | Perintah |
|---|---|
| Build & jalankan app + DB | `docker compose up --build` |
| Deploy ulang kode ke dev environment | `docker compose up -d --build app` |
| Seed data demo | `docker compose run --rm -e DATABASE_URL="postgres://jobhoo:jobhoo_dev_password@db:5432/jobhoo?sslmode=disable" app ./jobhoo-seed` |
| Hentikan | `docker compose down` |
| Reset database penuh | `docker compose down -v && docker compose up --build` |
| Lihat log | `docker compose logs -f app` |
| Shell ke database | `docker compose exec db psql -U jobhoo -d jobhoo` |
| Build & vet sebelum deploy | `go build ./... && go vet ./...` |
| Jalankan test | `go test ./...` |

## 10. Dokumen Terkait

1. [DOC-PRODUCT-OVERVIEW.md](DOC-PRODUCT-OVERVIEW.md) - Gambaran produk non-teknis.
2. [DOC-DEVELOPMENT-PHASE.md](DOC-DEVELOPMENT-PHASE.md) - Checklist status pengembangan per fase.
3. [DOC-AUDIT-REPORT.md](DOC-AUDIT-REPORT.md) - Audit teknis, keamanan, performa, dan UI/UX.
