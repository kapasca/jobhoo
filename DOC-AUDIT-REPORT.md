# JOBHOO Audit Report

Audit teknis, keamanan, performa, dan UI/UX terhadap JOBHOO. Dokumen ini menggabungkan seluruh temuan audit sebelumnya menjadi satu referensi tunggal. Untuk gambaran produk, lihat [DOC-PRODUCT-OVERVIEW.md](DOC-PRODUCT-OVERVIEW.md). Untuk panduan teknis, lihat [DOC-DEVELOPMENT-GUIDE.md](DOC-DEVELOPMENT-GUIDE.md). Untuk status pengerjaan per fitur, lihat [DOC-DEVELOPMENT-PHASE.md](DOC-DEVELOPMENT-PHASE.md).

Catatan tanggal: audit awal dilakukan 1-2 Agustus 2026. Sejumlah temuan (misalnya validasi resume dan status provider AI) perlu diverifikasi ulang terhadap kode saat ini sebelum dieksekusi, karena implementasi terus berubah.

## 1. Ringkasan Eksekutif

JOBHOO memiliki fondasi teknis yang solid: pemisahan tanggung jawab yang jelas, dependency injection konsisten, lapisan AI yang modular, dan sistem desain yang rapi. Fitur inti (pencarian, lamar, ATS, verifikasi perusahaan, ranking AI) berfungsi end-to-end. Kesenjangan utama berada di kelengkapan operasional (notifikasi email, verifikasi email, reset password), validasi input, dan beberapa isu performa yang baru terasa pada skala besar.

### 1.1 Skor Penilaian

| Aspek | Skor | Catatan |
|---|---|---|
| Teknis & Arsitektur | 8.5/10 | Kode bersih, pola DI konsisten, tanpa global state |
| Keamanan | 6.5/10 | Fondasi kuat (bcrypt, CSRF, sesi ter-hash), beberapa gap operasional |
| UI/UX & Aksesibilitas | 5-6/10 | Sistem desain matang, tapi ARIA/semantic HTML dan validasi form belum lengkap |
| Performa | 4-5/10 | N+1 query, tanpa caching, aset belum diminifikasi |
| Kepatuhan Data | 4/10 | Belum ada kebijakan privasi, retensi data, atau penghapusan akun |

### 1.2 Temuan Kritis

| No. | Kategori | Temuan | Prioritas | Estimasi Effort |
|---|---|---|---|---|
| 1 | Keamanan | Validasi MIME/ukuran file resume tidak ada | Kritis | 2-4 jam |
| 2 | Keamanan | Sanitasi HTML pada deskripsi job/company tidak diterapkan (bluemonday sudah jadi dependency tapi belum dipakai di semua input) | Kritis | 1-2 jam |
| 3 | Keamanan | TLS/HTTPS enforcement dan Secure flag pada cookie sesi belum ada | Kritis | 1 jam |
| 4 | Keamanan | Kompleksitas password hanya panjang minimum | Tinggi | 1 jam |
| 5 | Kepatuhan | Tidak ada endpoint penghapusan akun (GDPR right-to-erasure) | Kritis | 6-8 jam |
| 6 | Keamanan | Rate limiter membaca `r.RemoteAddr` langsung, mengabaikan `X-Forwarded-For` di belakang reverse proxy | Tinggi | 1 jam |
| 7 | Performa | N+1 query pada dashboard recruiter (hitung pelamar per job) | Tinggi | 2 jam |
| 8 | Aksesibilitas | Landmark semantic HTML, heading hierarchy, dan ARIA label belum konsisten | Tinggi | 4-6 jam |
| 9 | Bisnis/UX | Recruiter tidak mendapat feedback status approval secara real-time | Tinggi | 2-3 jam |
| 10 | UX | Validasi form hanya di sisi server, pesan error kurang spesifik | Sedang | 2-3 jam |

Estimasi jalur kritis total: sekitar 16-25 jam (2-3 hari kerja developer) untuk menyelesaikan seluruh temuan kritis.

## 2. Analisis Fungsionalitas

### 2.1 Kandidat (Pencari Kerja)

Fitur yang berfungsi: signup dengan verifikasi email (scaffold), upload resume dan manajemen profil, pencarian job dengan filter, detail job dan info perusahaan, apply dengan cegah duplikasi, pelacakan status lamaran real-time, bookmark job, rekomendasi job AI, saran perbaikan resume AI.

Kelemahan: validasi file resume tidak ada (bisa upload `.exe`/`.zip`), token verifikasi email relatif singkat tanpa alur resend, progres kelengkapan profil tidak ditampilkan, tidak ada saved search, transisi status lamaran tidak selalu jelas maknanya bagi kandidat.

### 2.2 Recruiter (Perusahaan)

Fitur yang berfungsi: registrasi company dengan alur approval, profil company lengkap (industri, deskripsi, logo), posting job dengan detail lengkap, penjadwalan `opens_at`/`closes_at`, close/archive/reopen job, papan ATS Kanban 5 tahap, pemindahan kandidat antar tahap, ranking kandidat AI, halaman publik company.

Kelemahan: recruiter menunggu approval admin tanpa indikator ETA atau progres, tidak ada notifikasi real-time saat status approval berubah, ATS board belum punya bulk operation, tidak ada duplikasi/template job, edit job yang sudah live masih terbatas.

Dampak bisnis: setiap jam keterlambatan approval berpotensi membuat recruiter beralih ke platform lain.

### 2.3 Admin

Fitur yang berfungsi: antrian approval company, approve/reject dengan alasan, blacklist company permanen, freeze/unfreeze user, modal detail untuk semua entitas, dashboard ringkasan platform.

Kelemahan: tidak ada bulk operation (misal reject beberapa company sekaligus), audit log minimal (tidak selalu tercatat siapa melakukan approval dan kapan), tidak ada mekanisme undo dalam rentang waktu tertentu, blacklist tidak punya masa berlaku atau mekanisme banding.

### 2.4 Integrasi AI

Status saat ini: JOBHOO menggunakan provider OpenAI secara eksklusif (lihat [DOC-DEVELOPMENT-GUIDE.md](DOC-DEVELOPMENT-GUIDE.md) Bagian 6). Sebagian temuan audit sebelumnya masih merujuk ke arsitektur multi-provider (mock/Anthropic/OpenAI) yang sudah tidak berlaku; hanya isu berikut yang masih relevan terhadap arsitektur saat ini:

1. Tidak ada fallback bila pemanggilan AI gagal atau timeout; permintaan pengguna akan langsung gagal.
2. Hasil ranking/rekomendasi tidak di-cache; permintaan berulang untuk kombinasi job/kandidat yang sama akan memanggil API lagi.
3. Pemrosesan AI berjalan sinkron di dalam request handler (blocking selama beberapa detik), belum async/background job.
4. Tidak ada breakdown transparansi yang ditampilkan ke recruiter tentang faktor apa saja yang membentuk skor ranking.

## 3. Keamanan

### 3.1 Autentikasi & Manajemen Sesi

Kekuatan: bcrypt cost 12, sesi tersimpan di database (bukan JWT) sehingga bisa direvokasi instan, token sesi 32-byte acak yang di-hash SHA256 sebelum disimpan, TTL sesi 30 hari, reset password merevokasi semua sesi, tabel sesi mencatat user agent dan IP.

Kekurangan: tanpa 2FA, cookie sesi belum memakai flag `Secure`, kompleksitas password hanya panjang minimum (tanpa syarat huruf besar/angka/simbol), tanpa lockout otomatis setelah beberapa kali percobaan login gagal.

Contoh perbaikan:

```go
sessionCookie := &http.Cookie{
    Name:     "sessionid",
    Value:    token,
    Path:     "/",
    Secure:   true, // hanya kirim lewat HTTPS
    HttpOnly: true,
    SameSite: http.SameSiteLax,
}
```

```go
func ValidatePassword(password string) error {
    if len(password) < 12 {
        return errors.New("password must be 12+ characters")
    }
    if !hasUpperCase(password) || !hasNumber(password) || !hasSpecialChar(password) {
        return errors.New("password must include uppercase, number, and special character")
    }
    return nil
}
```

### 3.2 Otorisasi & Access Control

Kekuatan: middleware `RequireAuth`/`RequireRole` diterapkan konsisten lewat Chi router group, 3 peran terpisah jelas (candidate/recruiter/admin), ownership check (recruiter hanya bisa kelola job company miliknya), freeze user langsung mencabut sesi.

Kekurangan: tidak ada permission granular di dalam satu peran (semua recruiter setara), tidak ada audit log untuk aksi admin (siapa approve/reject/freeze, kapan), tidak ada role hierarchy.

### 3.3 Validasi Input & Proteksi Data

Kekuatan: seluruh query memakai parameter binding (pgx), template Go otomatis melakukan escape HTML, CSRF double-submit cookie diterapkan pada form.

Kekurangan kritis:

1. **Validasi MIME file resume tidak ada.** File diterima tanpa memeriksa magic bytes atau ekstensi, berisiko arbitrary file upload.
2. **Sanitasi HTML pada deskripsi job/company belum konsisten.** `bluemonday` sudah menjadi dependency tapi belum tentu dipakai di semua alur input, berisiko XSS tersimpan.
3. **Validasi email longgar** dan tanpa pengecekan panjang maksimum.
4. **Tidak ada rate limit generik** pada endpoint apply job dan post job (berpotensi disalahgunakan untuk spam), berbeda dengan endpoint auth/email yang sudah dibatasi.
5. **Rate limiter mengabaikan header proxy** (`X-Forwarded-For`/`X-Real-IP`), sehingga di belakang reverse proxy semua request akan tercatat dari satu IP yang sama.

Contoh perbaikan validasi file:

```go
func ValidateResumeFile(file *multipart.FileHeader) error {
    ext := strings.ToLower(filepath.Ext(file.Filename))
    allowedExts := map[string]bool{".pdf": true, ".docx": true, ".doc": true, ".txt": true}
    if !allowedExts[ext] {
        return errors.New("only PDF, DOCX, DOC, TXT allowed")
    }
    if file.Size > 10*1024*1024 {
        return errors.New("file too large")
    }
    f, err := file.Open()
    if err != nil {
        return err
    }
    defer f.Close()
    buf := make([]byte, 512)
    f.Read(buf)
    mimeType := http.DetectContentType(buf)
    allowedMimes := map[string]bool{
        "application/pdf": true,
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
        "text/plain": true,
    }
    if !allowedMimes[mimeType] {
        return errors.New("invalid file type")
    }
    return nil
}
```

Contoh perbaikan rate limiter:

```go
func ClientIP(r *http.Request) string {
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        parts := strings.Split(xff, ",")
        return strings.TrimSpace(parts[0])
    }
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        return xri
    }
    return r.RemoteAddr
}
```

### 3.4 HTTPS & Transport Security

Status: belum dikonfigurasi. Docker Compose tidak menegakkan HTTPS, tidak ada konfigurasi TLS di aplikasi Go, cookie sesi belum memakai flag `Secure`, header HSTS belum ada.

Rekomendasi: jalankan di belakang reverse proxy (nginx/Caddy/Cloudflare) dengan sertifikat TLS dan header `Strict-Transport-Security`, lalu set flag `Secure` pada cookie saat `APP_ENV=production`.

### 3.5 Blacklist Company & Freeze User

Status: sudah cukup baik. Blacklist bersifat permanen (mencegah recruiter nakal mendaftar ulang), freeze langsung mencabut semua sesi user. Kekurangan: belum ada mekanisme banding, opsi blokir sementara dengan masa berlaku, atau notifikasi email ke pihak yang terdampak.

### 3.6 Kepatuhan & Privasi Data

Kesenjangan kritis terkait GDPR/privasi:

1. Tidak ada endpoint penghapusan akun (right to erasure).
2. Tidak ada dokumen kebijakan privasi.
3. Tidak ada kebijakan retensi data (berapa lama data disimpan, kapan dihapus).
4. Tidak ada data processing agreement dengan vendor AI.

Rekomendasi retensi (usulan awal, perlu tinjauan legal): kandidat tidak aktif dihapus setelah 2 tahun, lamaran yang ditolak dihapus setelah 6 bulan, log email dihapus setelah 90 hari, sesi dihapus 30 hari setelah kedaluwarsa.

## 4. Performa

### 4.1 Query Database

**N+1 query pada dashboard recruiter.** Daftar job diambil satu query, lalu jumlah pelamar per job dihitung dengan query terpisah untuk tiap job. Dengan 50 job, ini menjadi 51+ query per page load.

Perbaikan yang direkomendasikan (single query dengan join):

```sql
SELECT j.id, j.title, j.status, j.created_at,
       COUNT(a.id) AS applicant_count
FROM jobs j
LEFT JOIN applications a ON a.job_id = j.id
WHERE j.company_id = $1
GROUP BY j.id
ORDER BY j.created_at DESC;
```

Temuan performa lain:

1. Indeks belum lengkap untuk beberapa pola filter umum (`jobs(company_id, status)`, `applications(job_id, stage)`, `applications(candidate_id, stage)`, indeks GIN untuk `candidate_profiles.skills`).
2. Halaman admin (antrian approval, manajemen user) memuat seluruh data tanpa pagination; akan melambat signifikan begitu data bertambah banyak.
3. Listing job publik melakukan dua query terpisah (COUNT untuk pagination + SELECT data) yang bisa digabung dengan window function.
4. `GetSavedJobIDs` dipanggil sebagai query tambahan terpisah di beberapa halaman (home, listing, dashboard) yang bisa digabung lewat LEFT JOIN pada query utama.

### 4.2 HTTP & Pengiriman Aset

1. Tidak ada header `Cache-Control` pada aset statis maupun halaman dinamis.
2. Tidak ada kompresi gzip pada response.
3. HTMX dimuat dari CDN publik (unpkg.com), bukan self-hosted; berisiko delay bila CDN lambat/down.
4. Font Google (Lexend) dimuat lewat request eksternal, belum di-self-host.
5. Aset CSS/JS belum diminifikasi untuk production.

### 4.3 Pemrosesan AI

Ranking kandidat dan rekomendasi job diproses secara sinkron di dalam request handler. Recruiter menekan tombol "Rank Candidates" dan UI menunggu beberapa detik hingga selesai. Untuk beban kerja yang lebih besar, ini sebaiknya dipindah ke proses asinkron (job queue + polling atau notifikasi saat selesai).

## 5. UI/UX & Aksesibilitas

### 5.1 Konsistensi Desain

Sistem desain sudah matang: token warna/tipografi/spacing terpusat di `tokens.css`, komponen konsisten di seluruh halaman. Kekurangan minor: kontras warna oranye di atas navy berada di ambang batas WCAG AA (sekitar 3.8:1, cukup untuk teks besar tapi belum ideal untuk teks kecil), dan warna teks tersier (`ink-500`, `#7d84a3`) di atas background navy berada di bawah 4.5:1 sehingga gagal WCAG AA untuk teks normal. Rekomendasi: naikkan lightness `ink-500` atau gunakan warna khusus untuk placeholder.

### 5.2 Validasi Form & Pesan Error

1. Validasi mayoritas hanya di sisi server; tidak ada feedback real-time sebelum submit.
2. Pesan error generik (misalnya "All fields are required" walau hanya satu field yang kosong).
3. Input skill, cover note, dan deskripsi job belum memiliki batas panjang di sisi server (berisiko penyalahgunaan penyimpanan).
4. Beberapa form (termasuk sebagian modal admin) perlu dipastikan selalu menyertakan token CSRF.

### 5.3 Navigasi & Struktur Semantik

1. Beberapa halaman melompat urutan heading (h1 ke h3), termasuk dashboard admin dan papan ATS yang belum punya heading per kolom tahap.
2. Elemen seperti job card memakai `<div>` alih-alih elemen semantik (`<article>`, `<header>`).
3. Landmark HTML (`<main>`, `<footer>`, `<nav aria-label>`) belum diterapkan penuh di layout dasar.
4. Belum ada breadcrumb, dan beberapa modal (job detail, admin) belum mendukung tombol close yang konsisten atau tombol Escape untuk menutup.

### 5.4 ARIA & Dukungan Screen Reader

Beberapa komponen interaktif kekurangan atribut ARIA yang tepat: checkbox filter job tanpa `aria-label`/`aria-checked`, tombol bookmark tanpa `aria-pressed`, tab pada dashboard kandidat belum memakai `role="tablist"`/`role="tab"`/`aria-selected`, alert sukses/error belum memakai `role="status"`/`aria-live`. Ikon SVG dekoratif pada metadata job card belum ditandai `aria-hidden="true"`.

### 5.5 Elemen Lain

1. Alt text logo perusahaan sudah konsisten pada sebagian besar halaman, tapi belum di semua tempat (misalnya preview upload logo).
2. Empty state antar halaman belum konsisten gaya penulisannya.
3. Duplikasi antara versi halaman penuh dan versi modal (login, job detail) berisiko logic rendering yang berbeda; sebaiknya dikonsolidasi dengan satu template dan flag `IsModal`.
4. Skrip inline pada halaman ATS board cukup panjang dan sebaiknya diekstrak ke file JS terpisah agar lebih mudah diuji dan dipelihara.

## 6. Kepatuhan & Tata Kelola Data

Lihat Bagian 3.6 untuk detail kesenjangan GDPR. Ringkasan aksi yang diperlukan sebelum peluncuran publik:

1. Implementasi endpoint penghapusan akun dengan konfirmasi email dan password.
2. Publikasikan kebijakan privasi dan kebijakan retensi data.
3. Tinjau data processing agreement dengan vendor AI (OpenAI/gateway yang dipakai).

## 7. Kesenjangan Bisnis & Pengalaman Pengguna

1. **Friksi approval recruiter.** Recruiter tidak memiliki indikator status/ETA setelah signup; ini berisiko churn di awal funnel.
2. **Notifikasi status lamaran.** Kandidat tidak menerima email saat status lamaran berubah, sehingga platform terasa "sunyi".
3. **Edit job terbatas.** Recruiter tidak leluasa mengubah job yang sudah live tanpa kehilangan riwayat pelamar.
4. **Tidak ada pagination di admin**, yang akan menyulitkan penggunaan begitu jumlah company/job/user bertambah.

## 8. Rencana Perbaikan Bertahap

| Fase | Fokus | Estimasi Waktu |
|---|---|---|
| Fase 0 - Kritis | Validasi file resume, sanitasi HTML, TLS/Secure cookie, rate limit email, perbaikan N+1 query dashboard | 1-2 hari |
| Fase 1 - Kesiapan MVP | Notifikasi email, verifikasi email, reset password, validasi form per-field, feedback status approval recruiter, pagination admin | 2-3 hari |
| Fase 2 - Skala & Operasional | Pemrosesan AI asinkron, caching/HTTP headers, kompresi gzip, logging terstruktur, audit trail admin, kebijakan privasi/retensi data | 3-5 hari |

## 9. Kesimpulan

JOBHOO adalah platform dengan fondasi arsitektur dan keamanan yang solid. Kesenjangan yang ada bersifat operasional dan dapat diselesaikan tanpa perombakan besar: validasi input, notifikasi email, penegakan HTTPS, dan optimasi query adalah pekerjaan berskala hari, bukan minggu. Prioritas yang disarankan: selesaikan Fase 0 sebelum beta terbatas, selesaikan Fase 1 sebelum peluncuran publik, dan jadikan Fase 2 sebagai pekerjaan berkelanjutan pasca-peluncuran.
