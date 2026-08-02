# JOBHOO Product Overview

Dokumen ini ditujukan untuk pembaca umum yang ingin memahami apa itu JOBHOO, siapa penggunanya, dan bagaimana cara kerjanya, tanpa perlu memahami detail teknis di baliknya. Untuk detail teknis, lihat [DOC-DEVELOPMENT-GUIDE.md](DOC-DEVELOPMENT-GUIDE.md).

## 1. Ringkasan

JOBHOO adalah platform rekrutmen online yang menghubungkan pencari kerja (kandidat) dengan perusahaan yang sedang membuka lowongan. Fokus platform adalah memudahkan proses rekrutmen dari awal hingga keputusan hire, bukan menjadi jaringan sosial profesional.

Prinsip utama JOBHOO:

1. Kandidat menemukan pekerjaan yang sesuai dengan keahlian mereka.
2. Perusahaan mengelola seluruh proses seleksi dalam satu tempat.
3. AI membantu kedua sisi, tetapi tidak pernah mengambil keputusan akhir sendiri.

## 2. Pengguna Platform

JOBHOO melayani tiga jenis pengguna.

### 2.1 Kandidat (Pencari Kerja)

Individu yang sedang mencari pekerjaan baru. Mendaftar secara gratis, melengkapi profil, mengunggah resume, dan melamar ke lowongan yang tersedia.

### 2.2 Recruiter / Perusahaan

Perwakilan dari sebuah perusahaan yang ingin merekrut karyawan baru. Setiap recruiter mendaftarkan perusahaannya, menunggu verifikasi dari admin JOBHOO, lalu dapat membuka lowongan dan mengelola pelamar.

### 2.3 Admin JOBHOO

Tim internal yang memastikan hanya perusahaan sah yang boleh membuka lowongan di platform. Admin memverifikasi setiap pendaftaran perusahaan baru dan menjaga kualitas platform.

## 3. Alur Kerja

### 3.1 Alur Kandidat

1. Daftar akun dan verifikasi (alur email verifikasi masih dalam pengembangan, lihat Bagian 8).
2. Unggah resume dan lengkapi profil (headline, skill, lokasi).
3. Cari lowongan dengan filter kategori, lokasi, dan tipe kerja.
4. Lamar pekerjaan disertai catatan singkat untuk recruiter.
5. Pantau status lamaran di dashboard (Applied - Screening - Interview - Offer - Hired).
6. Simpan lowongan yang menarik untuk dipertimbangkan nanti.
7. Dapatkan rekomendasi lowongan dari AI berdasarkan skill di profil.

### 3.2 Alur Recruiter

1. Daftar akun recruiter dan daftarkan perusahaan (nama, industri, deskripsi, logo).
2. Tunggu verifikasi admin. Recruiter tidak dapat membuka lowongan sebelum disetujui.
3. Setelah disetujui, buka lowongan dengan detail lengkap: judul, deskripsi, lokasi, tipe kerja, skill, dan rentang gaji.
4. Kelola pelamar melalui papan Kanban (ATS board): pindahkan kandidat antar tahap dengan satu aksi.
5. Gunakan ranking AI sebagai saran urutan prioritas kandidat berdasarkan kecocokan skill, bukan keputusan final.
6. Perusahaan yang terverifikasi memiliki halaman publik yang menampilkan profil dan semua lowongan aktif.

### 3.3 Alur Admin

1. Periksa antrian persetujuan perusahaan baru.
2. Tinjau profil perusahaan: industri, deskripsi, informasi recruiter.
3. Setujui atau tolak pendaftaran, dengan alasan penolakan dikirim ke recruiter.
4. Blacklist perusahaan bila diperlukan (blokir permanen untuk pelaku yang tidak sah).
5. Freeze akun user bila ada masalah keamanan; sesi user langsung dicabut.

## 4. Fitur Utama

### 4.1 Untuk Kandidat

| Fitur | Penjelasan |
|---|---|
| Pencarian cerdas | Cari berdasarkan judul, skill, nama perusahaan, atau kategori |
| Filter lokasi | 9 negara (Indonesia, Singapura, Malaysia, Thailand, Vietnam, Filipina, Timor-Leste, Australia, Selandia Baru) beserta provinsinya |
| Filter tipe kerja | Remote, Hybrid, atau Onsite |
| Upload resume | Simpan file resume di platform, dapat diperbarui kapan saja |
| Status lamaran real-time | Status lamaran selalu terlihat di dashboard |
| Bookmark lowongan | Simpan lowongan untuk dipertimbangkan nanti |
| Analisis resume oleh AI | Platform menganalisis resume dan memberikan saran perbaikan |
| Rekomendasi lowongan AI | Sistem merekomendasikan lowongan berdasarkan skill kandidat |

### 4.2 Untuk Recruiter / Perusahaan

| Fitur | Penjelasan |
|---|---|
| Profil perusahaan | Halaman publik dengan logo, deskripsi, industri, dan daftar lowongan aktif |
| Upload logo | Recruiter dapat mengunggah logo langsung dari platform |
| Kelola lowongan | Buat, edit, tutup, atau arsipkan lowongan |
| Jadwal otomatis | Atur kapan lowongan mulai dan berhenti tampil ke publik |
| Pipeline pelamar (ATS) | Kelola pelamar dalam tahapan Applied - Screening - Interview - Offer - Hired |
| Ranking kandidat AI | Saran urutan prioritas kandidat berdasarkan kecocokan skill, sebagai referensi bukan keputusan otomatis |

## 5. Keunggulan JOBHOO

1. **AI sebagai asisten, bukan pengambil keputusan.** AI hanya memberi rekomendasi; recruiter tetap yang memutuskan. Tidak ada kandidat yang tersembunyi atau ditolak otomatis oleh sistem.
2. **Semua dalam satu platform.** Posting lowongan dan sistem manajemen pelamar (ATS) menyatu: satu login, satu alur kerja.
3. **Transparansi untuk kandidat.** Kandidat selalu bisa melihat di tahap mana lamarannya berada, tidak menghilang begitu saja setelah melamar.
4. **Verifikasi perusahaan.** Setiap perusahaan harus diverifikasi admin sebelum membuka lowongan, mengurangi risiko lowongan palsu.
5. **Fokus rekrutmen, bukan jejaring sosial.** Tidak ada feed, koneksi, atau like. Platform ini murni untuk mencari pekerjaan dan merekrut karyawan.

## 6. Cakupan Geografis

JOBHOO saat ini mendukung lokasi kerja di:

- Asia Tenggara: Indonesia (38 provinsi), Singapura, Malaysia, Thailand, Vietnam, Filipina, Timor-Leste.
- Oceania: Australia, Selandia Baru.

Setiap negara memiliki daftar provinsi/negara bagian lengkap sehingga pencarian lokasi lebih presisi.

## 7. Kategori Pekerjaan

Lowongan di JOBHOO dikelompokkan dalam 5 kategori utama:

1. Engineering & Product - Pengembang software, DevOps, Product Manager, QA, dan lainnya.
2. Design & Creative - UI/UX Designer, Illustrator, Creative Director, dan lainnya.
3. Sales & Marketing - Account Executive, Growth Marketer, Content Manager, dan lainnya.
4. Data & Analytics - Data Analyst, Data Scientist, BI Developer, ML Engineer, dan lainnya.
5. Operations & Support - Operations Manager, Customer Success, People Ops, dan lainnya.

Recruiter juga dapat mengetikkan kategori sendiri jika tidak ada yang cocok.

## 8. Keterbatasan Saat Ini

JOBHOO masih dalam tahap pengembangan aktif. Beberapa hal yang belum tersedia sepenuhnya:

1. Notifikasi email - kandidat belum menerima email otomatis saat status lamaran berubah.
2. Reset password dan verifikasi email - infrastruktur token sudah ada, alur pengguna belum lengkap.
3. Validasi file resume - belum ada pemeriksaan tipe file/ukuran maksimum saat upload.
4. Aplikasi mobile - saat ini hanya tersedia versi web, belum ada aplikasi Android/iOS.

Rincian status pengembangan per fitur ada di [DOC-DEVELOPMENT-PHASE.md](DOC-DEVELOPMENT-PHASE.md). Rincian temuan teknis dan keamanan ada di [DOC-AUDIT-REPORT.md](DOC-AUDIT-REPORT.md).

## 9. Status Platform

JOBHOO saat ini berfungsi sebagai prototipe/MVP yang sudah dapat digunakan. Semua fitur utama (pencarian, lamar, ATS, verifikasi perusahaan, ranking AI) berjalan dan telah diuji secara fungsional. Platform siap untuk demo dan pengujian oleh pengguna nyata, dengan sejumlah perbaikan tersisa sebelum peluncuran publik penuh.
