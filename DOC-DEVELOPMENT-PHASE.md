# JOBHOO Development Phase Checklist

Status pengerjaan JOBHOO per fase, dirinci per fitur. Untuk gambaran produk, lihat [DOC-PRODUCT-OVERVIEW.md](DOC-PRODUCT-OVERVIEW.md). Untuk panduan teknis, lihat [DOC-DEVELOPMENT-GUIDE.md](DOC-DEVELOPMENT-GUIDE.md). Untuk temuan audit, lihat [DOC-AUDIT-REPORT.md](DOC-AUDIT-REPORT.md).

Format penanda: `[x]` selesai dan sudah diverifikasi, `[~]` ada tapi belum lengkap/belum teruji penuh, `[ ]` belum dikerjakan.

## 1. Fase 1 - Scaffold dan Fondasi

- [x] Struktur project Go (cmd/, internal/, web/)
- [x] Docker + Docker Compose (app + Postgres)
- [x] Schema database & migration terstruktur
- [x] Design token (warna navy/orange/white/black, spacing, tipografi Lexend)
- [x] Component library CSS dasar
- [x] Sistem utility class CSS (di-upgrade signifikan dari inline style, hasil kerja Anda)
- [x] Homepage dasar
- [x] Job listing dasar
- [x] Seeder data dummy (10 company/recruiter, 100 job lintas 5 kategori)
- [x] Logo & brand asset terpasang di layout

---

## 2. Fase 2 - Autentikasi

- [x] Signup (pilih role candidate/recruiter)
- [x] Login
- [x] Logout
- [x] Password hashing (bcrypt, cost 12)
- [x] Sesi tersimpan di database (bukan JWT), token di-hash sebelum disimpan
- [x] Sesi bisa di-revoke
- [x] Proteksi route berbasis role (RequireAuth, RequireRole)
- [x] Login dipindah ke popup/modal (redesign Anda)
- [x] **Bug: login 404 saat diakses non-modal (mis. redirect otomatis) - DIPERBAIKI**
- [ ] Reset password (lupa password)
- [ ] Verifikasi email saat signup
- [ ] Rate limiting percobaan login
- [ ] Resume validation at signup: when a candidate uploads a resume during registration,
  automatically scan and classify the file to ensure it is a valid resume/profile
  (PDF/DOCX content). If the detector determines the file is not a resume, the
  candidate must not be allowed to apply to jobs until a valid resume is provided
  or a manual review approves the upload.
  - If validation is async/slow, allow the account to be activated immediately
	but disable the `Apply` button and show "Resume verification in progress"
	on the candidate dashboard. Notify the candidate when verification
	succeeds or fails and provide remediation steps.

---

## 3. Fase 3 - Job Detail dan Apply Flow

- [x] Halaman detail job lengkap
- [x] Apply dengan cover note
- [x] Cegah apply dobel ke job yang sama
- [x] Save/bookmark job
- [x] Job detail dipindah ke popup/modal (redesign Anda)
- [x] **Bug: job detail 404 saat diakses langsung/di-share - DIPERBAIKI** (sekarang fallback ke halaman penuh)
- [x] **Bug: form apply merender modal bersarang ke dirinya sendiri - DIPERBAIKI**
- [ ] Upload resume file (PDF/DOCX) saat apply
- [ ] Notifikasi email ke recruiter saat ada apply baru
- [ ] Notifikasi ke kandidat saat status lamaran berubah

---

## 4. Fase 4 - Recruiter dan ATS

- [x] Setup company (sekali di awal, sebelum posting job)
- [x] Form post job (kategori, tipe kerja, lokasi, gaji, skill)
- [x] Dashboard recruiter (daftar job + jumlah pelamar)
- [x] ATS board Kanban (Applied → Screening → Interview → Offer → Hired)
- [x] Bagian Rejected terpisah (collapsible)
- [x] Ubah stage pelamar
- [x] AI ranking kandidat (tombol "Rank candidates with AI", advisory only)
- [ ] **Edit lowongan yang sudah diposting** *(prioritas tinggi, lihat catatan bisnis)*
- [ ] **Tutup/arsipkan lowongan** *(prioritas tinggi, lihat catatan bisnis)*
- [ ] Bulk action (ubah beberapa pelamar sekaligus)
- [ ] Catatan internal recruiter per kandidat
- [ ] Jadwal interview/kalender
- [ ] Multi-recruiter per company (saat ini 1 company = 1 akun pemilik)

---

## 5. Fase 5 - Lapisan AI

- [x] Interface `ai.Provider` modular (ganti vendor tanpa ubah kode aplikasi)
- [x] Mock provider - logic asli (keyword-overlap), tanpa API key, teruji
- [x] Anthropic provider - kode pemanggilan API lengkap dan benar secara logika
- [ ] **Anthropic provider belum pernah dites dengan API key sungguhan** *(prioritas menengah, risiko tersembunyi)*
- [ ] OpenAI provider (masih stub kosong)
- [x] Resume summarization - logic ada
- [x] Job recommendation - terhubung ke dashboard kandidat
- [x] Resume improvement suggestion - terhubung ke halaman profil
- [x] Candidate ranking - terhubung ke ATS board
- [~] `ExplainMatch` (jelaskan alasan match per job) - **logic sudah ada di kode, belum disambungkan ke UI mana pun**
- [ ] Compare candidates side-by-side

---

## 6. Fase 6 - Candidate Profile dan Dashboard

- [x] Edit profil (headline, lokasi, skill, resume teks)
- [x] Dashboard: daftar lamaran + status
- [x] Dashboard: daftar job tersimpan
- [x] Tombol AI saran perbaikan resume
- [x] Tombol AI rekomendasi job
- [ ] Upload resume file (PDF/DOCX) - masih plain text

- [ ] Resume verification on profile: when a candidate uploads or updates a
	resume on their profile, run the same automated validation. While verification
	is pending keep the account active but disable `Apply` (show pending status).
	On rejection, prevent new applications until the candidate replaces the file
	or requests manual review; on success enable `Apply` and surface the verified
	state in the dashboard.
- [ ] Tampilan `ExplainMatch` di halaman detail job (kenapa kamu cocok/tidak untuk job ini)
- [ ] Riwayat/timeline perubahan status lamaran yang lebih detail

---

## 7. Fase 7 - Admin Dashboard

- [x] Angka ringkas platform (total user, kandidat, recruiter, company, job, aplikasi)
- [ ] Aksi moderasi (nonaktifkan user, hapus job spam)
- [ ] Log aktivitas

---

## 8. Fase 8 - Polish, Search, dan Keamanan

- [x] CSRF protection (double-submit cookie) di semua form
- [x] Halaman 404 kustom sesuai desain
- [x] Ownership check (recruiter tidak bisa akses pipeline company lain)
- [x] **Search pintar** - satu kotak mencocokkan judul, company, kategori, skill sekaligus
- [x] Filter Sort, Location, Work Type, Employment Type - benar-benar berfungsi
- [x] Label filter aktif human-readable ("Full Time" bukan "full_time")
- [x] Halaman Explore Companies (directory publik)
- [x] Grid job card responsif (CSS Grid auto-fill/minmax, tanpa breakpoint manual)
- [x] Perbaikan kolateral: grid admin dashboard & ATS rejected-section ikut jalan lagi
- [ ] Halaman detail per company (saat ini baru directory + link ke hasil search)
- [ ] Rate limiting (login, apply, signup)
- [ ] Reset password *(tumpang tindih dengan Fase 2)*
- [ ] Notifikasi email (apply baru, ubah status) *(tumpang tindih dengan Fase 3)*
- [ ] Audit aksesibilitas (screen reader, kontras, keyboard nav) - belum pernah dilakukan
- [ ] Verifikasi visual langsung di browser oleh manusia - semua pengecekan saya sejauh ini lewat kode dan template engine, belum pernah "dilihat mata" oleh saya

---

## 9. Rekap Angka

| Fase | Item selesai | Item belum |
|---|---|---|
| 1. Scaffold | 10/10 | 0 |
| 2. Auth | 9/12 | 3 |
| 3. Job detail + apply | 7/10 | 3 |
| 4. Recruiter + ATS | 7/13 | 6 |
| 5. AI | 8/12 (1 setengah) | 3-4 |
| 6. Candidate profile | 5/8 | 3 |
| 7. Admin | 1/3 | 2 |
| 8. Polish/security | 9/16 | 7 |

**Total kasar: kurang lebih 56/94 item granular (kurang lebih 60%).** Angka ini lebih rendah dari taksiran "88-90%" sebelumnya karena checklist ini jauh lebih rinci (per fitur, bukan per fase). Jangan terkecoh persentase: bagian yang paling menentukan kelayakan pakai (loop inti apply-review-hire) sudah solid; sisanya kebanyakan hal pelengkap (notifikasi, reset password, upload file) yang penting tapi tidak memblokir penggunaan dasar.
