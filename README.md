# JOBHOO

**Connecting Recruiters with Outstanding Talent**

Job portal MVP yang production-ready: bersih, cepat, dan sudah bisa dicoba end-to-end (registrasi, login, approval recruiter, membuat lowongan, melamar, ATS kanban, hingga status akhir lamaran).

## Tech Stack

| Layer | Teknologi | Catatan |
|---|---|---|
| Bahasa | Go 1.22 | Backend & frontend jadi satu binary (server-rendered) |
| Routing | `net/http` (stdlib, Go 1.22+ pattern matching) | Tanpa router pihak ketiga |
| Templating | `html/template` (stdlib) | Auto-escape XSS, tanpa build step |
| Interaktivitas | [htmx](https://htmx.org) | ATS Kanban update tanpa reload, tanpa SPA framework |
| Styling | Tailwind CSS (CDN) | Lihat catatan produksi di bawah |
| Database | PostgreSQL 16 | Akses via `database/sql` + `lib/pq` (driver pure-Go) |
| Auth | Session cookie + bcrypt | Role-based access via middleware |
| Deploy | Docker + Docker Compose | Satu image, mudah di-deploy ke VPS mana pun |

### Perubahan dari rencana awal (dan alasannya)

Project ini awalnya direncanakan dengan Next.js + Laravel, lalu disederhanakan menjadi **Go + htmx monolith** atas permintaan eksplisit, dengan dua penyesuaian kecil demi konsistensi filosofi "minimalis, dependency seminimal mungkin":

1. **`html/template` (stdlib) menggantikan `templ`** — sama-sama aman dari XSS, tapi tanpa compiler/tool tambahan saat build.
2. **`lib/pq` + repository manual menggantikan `sqlc`** — sqlc butuh code-gen tool terpisah; pure `database/sql` tetap type-safe di level Go struct tanpa build step tambahan.
3. **Tailwind via CDN untuk development** — untuk production, ganti ke [Tailwind Standalone CLI](https://tailwindcss.com/blog/standalone-cli) agar tidak bergantung ke CDN eksternal (lihat bagian Produksi).

## Menjalankan secara Lokal (Docker — direkomendasikan)

```bash
cp .env.example .env   # opsional, docker-compose sudah punya default sendiri
docker compose up --build
```

Setelah container `app` dan `db` sehat, isi database dengan data dummy (akun demo, contoh lowongan):

```bash
docker compose run --rm seed
```

Buka **http://localhost:8080**

## Menjalankan Mode Development dengan Docker

Kalau Anda ingin perubahan `.go`, template HTML, atau aset static langsung terdeteksi saat save, gunakan compose development:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

Lalu buka **http://localhost:8070**.

Catatan:
- Source code di-mount ke container, jadi perubahan file akan ikut terbaca.
- `air` akan restart server otomatis saat file yang relevan berubah.
- Database tetap memakai volume yang sama, jadi data tidak hilang.

## Menjalankan Tanpa Docker

Prasyarat: Go 1.22+, PostgreSQL 16+.

```bash
createdb jobhoo
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/jobhoo?sslmode=disable"

go run ./cmd/seed     # migrasi otomatis + isi data dummy
go run ./cmd/server    # jalankan aplikasi di :8080
```

## Akun Demo (setelah `seed`)

| Role | Email | Password |
|---|---|---|
| Super Admin | admin@jobhoo.com | admin12345 |
| Recruiter (sudah approved) | recruiter@jobhoo.com | recruiter123 |
| Recruiter (menunggu approval) | pending-recruiter@jobhoo.com | recruiter123 |
| Candidate | candidate@jobhoo.com | candidate123 |

Recruiter yang sudah di-approve akan punya 3 lowongan contoh, dan candidate sudah melamar 2 di antaranya (lengkap dengan AI Match Score dummy).

## Struktur Project

```
cmd/
  server/          entry point aplikasi utama
  seed/             script untuk isi data dummy
internal/
  db/               koneksi & migration runner
  models/           struct domain (User, Job, Application)
  repository/       akses database (query SQL langsung)
  services/
    auth/           hashing password & session
    aimatching/      AI matching (mock, siap ganti provider asli)
  middleware/       auth & role-based access control
  handlers/         HTTP handler per fitur + template renderer
migrations/         SQL migration
web/
  templates/        layout, partials, dan halaman (html/template)
  static/           aset statis (CSS/JS tambahan bila diperlukan)
uploads/            resume & dokumen recruiter yang diupload
```

## Mengganti AI Matching dengan Provider Asli

Semua logika AI matching diisolasi di `internal/services/aimatching/aimatching.go` di balik interface `Provider`:

```go
type Provider interface {
    Analyze(resumePath, jobDescription, jobRequirements string) (Result, error)
}
```

Untuk mengganti mock dengan OpenAI/Gemini/Claude/model lokal, buat implementasi baru yang memenuhi interface ini, lalu ganti `NewProvider()` — tidak ada bagian lain dari codebase yang perlu diubah.

## ATS Kanban

Pipeline lamaran: `Applied → Resume Reviewed → Interview → Offered → Hired` atau `Applied → Rejected`.
Recruiter memindahkan kandidat antar kolom lewat dropdown per kartu (update via htmx, tanpa reload halaman).

Ketika lowongan ditutup, setiap lamaran **wajib** diberi salah satu status akhir: *Proceed to Next Stage*, *Resume Not Reviewed*, atau *Not Matched* — ini ditegakkan lewat UI (dropdown status akhir otomatis muncul saat lowongan berstatus closed).

## Untuk Produksi

1. **Tailwind**: ganti `<script src="https://cdn.tailwindcss.com">` di `web/templates/layout/base.html` dengan CSS hasil compile [Tailwind Standalone CLI](https://tailwindcss.com/blog/standalone-cli), lalu sajikan lewat `web/static/`.
2. **HTTPS**: gunakan `docker-compose.prod.yml` (Nginx reverse proxy) di depan `app`, lalu tambahkan sertifikat TLS (mis. via Certbot) di konfigurasi Nginx.
3. **Secrets**: jangan commit `.env`; set `DATABASE_URL` dan variabel lain lewat environment variable platform deploy kamu (Fly.io, Railway, VPS systemd, dsb).
4. **Upload storage**: saat ini disk lokal (volume Docker). Untuk multi-instance deployment, ganti `saveUpload` di `internal/handlers/app.go` agar menyimpan ke S3-compatible storage.

## Placeholder / Belum Diimplementasi

Sesuai instruksi awal ("jika ada fitur yang belum selesai, buat placeholder yang jelas tanpa merusak alur aplikasi"):

- **Drag-and-drop visual** di ATS Kanban — saat ini perpindahan status lewat dropdown per kartu (fungsional, hasil sama), bukan drag fisik. Bisa ditingkatkan dengan menambahkan SortableJS + htmx `hx-post` di event `dragend` bila diinginkan.
- **Notifikasi email** (approval recruiter, status lamaran) — belum ada; struktur service AI matching bisa dijadikan contoh pola yang sama untuk menambah `internal/services/notification/`.
- **Reset password** — belum ada; tidak disebutkan di requirement awal.
