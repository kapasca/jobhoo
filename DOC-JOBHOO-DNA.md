# JOBHOO Design DNA

Standar visual dan desain resmi JOBHOO. Semua komponen, halaman, dan aset baru harus mengikuti panduan ini agar identitas platform tetap konsisten.

---

## 1. Kepribadian Merek (Brand Personality)

JOBHOO adalah platform rekrutmen yang **profesional namun tidak kaku**. Desainnya mencerminkan:

- **Percaya diri** — tampilan gelap dan bersih, bukan abu-abu atau putih seperti kebanyakan job board
- **Fokus** — tidak ada elemen dekoratif berlebihan; setiap elemen punya fungsi
- **Hangat** — sentuhan oranye yang terkontrol agar tidak dingin atau korporat steril
- **Modern** — rounded corners, spacing besar, tipografi berat

Kata kunci desain: **dark, clean, orange-accented, rounded, purposeful**.

---

## 2. Warna

### Skema Utama

JOBHOO menggunakan **dark theme sepenuhnya**. Tidak ada mode terang (light mode).

| Nama | Hex | Penggunaan |
|---|---|---|
| Navy 700 *(Brand Navy)* | `#192132` | Background utama halaman |
| Navy 900 | `#0f1220` | Ujung bawah gradient atmosphere |
| Surface Card | `#1f2942` | Kartu, panel, form |
| Surface Inset | `#151d30` | Input field, area recessed |
| Orange 500 *(Brand Orange)* | `#d96600` | Aksen utama, CTA, active state |
| Orange 400 | `#e87d33` | Hover state tombol oranye, icon |
| Ink 100 | `#eef0f6` | Teks utama (body, label, heading) |
| Ink 300 | `#b7bdd1` | Teks sekunder (subtitle, hint) |
| Ink 500 | `#7d84a3` | Teks tersier (placeholder, caption) |
| Border | `#2a3549` | Garis batas komponen |
| Border Strong | `#3a4462` | Garis lebih menonjol (hover, focus) |

### Aturan Warna

- **Oranye hanya untuk aksen** — bukan background, bukan teks paragraf
- **Tidak ada warna lain** selain navy dan oranye (kecuali status: hijau untuk sukses, merah untuk error)
- **Background halaman** selalu gradient: `linear-gradient(180deg, #192132 0%, #0f1220 100%)`
- **Overlay/backdrop** menggunakan `rgba(0, 0, 0, 0.4)` di atas background

### Warna Status

| Status | Warna |
|---|---|
| Sukses / Hired | `#045b25` (hijau gelap) |
| Error / Rejected | `#5b0404` (merah gelap) |
| Warning / Pending | `rgba(217, 102, 0, 0.08)` dengan border `#d96600` |
| Netral / Closed | Surface Inset dengan Ink 500 |

---

## 3. Tipografi

### Typeface

**Lexend** — satu-satunya typeface yang digunakan di seluruh platform.

- Diambil dari Google Fonts
- Dipilih karena: keterbacaan tinggi, karakter geometrik yang selaras dengan logo
- Fallback: `-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif`

### Skala Ukuran

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

### Line Height

| Token | Nilai | Penggunaan |
|---|---|---|
| `leading-tight` | 1.2 | Heading, judul besar |
| `leading-normal` | 1.55 | Body text, paragraf |
| `leading-relaxed` | 1.75 | Deskripsi panjang, konten artikel |

### Bobot Font

- **700 (Bold)** — Heading (h1–h4), label form, nama perusahaan
- **600 (SemiBold)** — Tombol, nav link aktif, badge, teks penting
- **500 (Medium)** — Nav link inaktif, metadata
- **400 (Regular)** — Body text, placeholder, deskripsi

---

## 4. Spacing

Sistem spacing berbasis **8px** (0.25rem = 4px base, kelipatan 2):

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

**Prinsip:** Elemen dalam satu komponen pakai `space-1` s/d `space-3`. Antar komponen pakai `space-4` s/d `space-6`. Section/halaman pakai `space-7` s/d `space-9`.

---

## 5. Border Radius (Bentuk Sudut)

JOBHOO menggunakan **rounded corners yang konsisten** — tidak ada elemen kotak (0px) kecuali untuk garis/divider.

| Token | Nilai | Digunakan Pada |
|---|---|---|
| `radius-sm` | 10px | Badge, chip, tombol kecil, tag |
| `radius-md` | 16px | Input field, tombol utama, dropdown |
| `radius-lg` | 22px | Kartu (card), panel, modal |
| `radius-xl` | 28px | Modal besar, hero section |
| `radius-pill` | 999px | Badge pill, filter chip |

**Aturan:** Semakin besar elemen, semakin besar radius-nya. Konsistensi ini mencerminkan bentuk logo JOBHOO yang rounded.

---

## 6. Elevasi & Bayangan

| Token | Nilai | Penggunaan |
|---|---|---|
| `shadow-sm` | `0 2px 8px rgba(10,12,22, 0.24)` | Kartu statis |
| `shadow-md` | `0 8px 24px rgba(10,12,22, 0.32)` | Modal, dropdown, floating element |
| `shadow-glow-orange` | `0 0 0 4px rgba(217,102,0, 0.16)` | Focus ring, input aktif |

---

## 7. Komponen Utama

### Tombol (Button)

Tombol tersedia dalam 3 varian:

| Varian | Tampilan | Penggunaan |
|---|---|---|
| **Primary** | Background oranye (`#d96600`), teks putih | Aksi utama (Publish Job, Apply, Save) |
| **Secondary** | Transparan, border `border-strong`, teks putih | Aksi sekunder (Edit, Cancel) |
| **Ghost** | Transparan, teks `ink-300` | Navigasi ringan, link-like action |

Ukuran tombol:
- **Default** — padding `0.75rem 1.375rem`, font 1rem
- **Small (--sm)** — padding `0.5rem 1rem`, font 0.75rem
- **Large (--lg)** — padding `0.875rem 1.75rem`, font 1rem

Hover state: Primary → `#e87d33` + orange glow. Secondary → background oranye (same as primary).

---

### Kartu (Card)

- Background: `surface-card` (`#1f2942`)
- Border: 1px solid `border`
- Radius: `radius-lg` (22px)
- Padding standar: `space-5` (24px) atas kiri kanan, `space-3` (12px) atas

Job Card memiliki struktur khas:
1. **Header** — logo perusahaan kecil + nama company (kiri) + posted time + bookmark (kanan)
2. **Judul lowongan** — full width, bold
3. **Body** — logo besar + metadata (kategori, lokasi, tipe kerja, tipe kontrak)

---

### Badge & Chip

- Radius: `radius-pill` (999px)
- Padding: `0.3rem 0.7rem`
- Font: `text-xs`, weight 600

Varian badge:
- **Default** — background surface-inset, teks ink-300
- **Orange** — background `rgba(oranye, 0.12)`, teks `orange-400`, border `rgba(oranye, 0.3)`
- **Green** — background `#045b25`, teks putih (status Hired)
- **Red** — background `#5b0404`, teks putih (status Rejected)
- **Gray** — background surface-inset, teks ink-500 (status Closed/Archived)

Skill Chip (input interaktif): oranye transparan dengan tombol × di ujung kanan.

---

### Form & Input

- Background input: `surface-inset` (`#151d30`)
- Border: 1px solid `border`
- Radius: `radius-md` (16px)
- Padding: `0.75rem 1rem`
- Focus: border berubah ke `orange-500` + `shadow-glow-orange`
- Placeholder: warna `ink-500`

File upload button: background oranye (`orange-500`), rounded kiri saja, menyatu dengan field. Hover → `orange-400`.

---

### Page Header (Judul Halaman)

Setiap halaman utama (dashboard, browse, profile, dll.) memiliki header gradient:

- Background: `linear-gradient(to right, rgba(0,0,0,0.4), rgba(0,0,0,0.4), transparent)`
- Min-height: 130px
- Padding: `space-6` atas-bawah
- Judul: `text-4xl`, bold, putih
- Subtitle: `text-lg`, `ink-300`
- Action button (jika ada): rata kanan dalam row yang sama

---

### Navigasi (Navbar)

- Sticky di top
- Background: `rgba(navy, 0.85)` + backdrop-blur 10px (glassmorphism ringan)
- Border bawah: 1px solid `border`
- Logo tinggi: 69px desktop, 60px mobile

Nav berbeda per role:
- **Kandidat:** Browse Jobs | [nama user] | Dashboard
- **Rekruter:** Job Management | Public Page | My Profile
- **Admin:** Browse Jobs | Dashboard
- **Tamu:** Browse Jobs | Post a Job

---

### Modal / Dialog

- Overlay: `rgba(0,0,0,0.5)`
- Panel: `surface-card`, `radius-lg`, shadow-lg
- Max-width: 700px
- Max-height: 90vh dengan scroll internal
- Animasi masuk: slide dari atas (`translateY(-20px)` → 0) dengan `opacity 0 → 1`, durasi 300ms
- Mobile: bottom-sheet (radius hanya di atas, lebar 100%, max-height 92vh)

---

## 8. Layout & Grid

### Container

| Class | Max-width | Penggunaan |
|---|---|---|
| `.container` | 1200px | Layout umum |
| `.container-max-sm` | 480px | Modal login, form singkat |
| `.container-max-md` | 520px | Form signup, company setup |
| `.container-max-lg` | 680px | Form post job, profile |
| `.container-max-xl` | 840px | Job detail, company detail kecil |
| `.container-max-2xl` | 900px | Company detail publik |

### Grid Sistem

- **Auto-fill** (`minmax(300px, 1fr)`) — Daftar job card, auto responsive
- **2 kolom** (`1fr 1fr`) — Form dengan dua field sejajar
- **3 kolom** (`1fr 1fr 1fr`) — Form dengan tiga field (currency, salary min/max)
- Mobile: semua grid collapse ke 1 kolom pada layar ≤ 640px

### Breakpoint

| Breakpoint | Aturan |
|---|---|
| ≤ 640px | Mobile: grid 1 kolom, bottom-sheet modal, FAB filter jobs |
| ≤ 768px | Tablet: hamburger menu, hidden desktop nav |
| ≤ 1024px | Sidebar jobs diperkecil (260px), header featured diperkecil |

---

## 9. Animasi & Motion

- **Easing standar:** `cubic-bezier(0.4, 0, 0.2, 1)` — material design standard ease
- **Fast (120ms):** Hover color, border color change, focus ring
- **Normal (200ms):** Tombol transform, card hover lift, dropdown
- **Modal enter (300ms):** Slide-in dari atas dengan fade
- **Mobile drawer (220ms):** Slide-in dari kanan

**Prinsip:** Tidak ada animasi dekoratif. Setiap motion punya tujuan — memberi feedback visual bahwa sesuatu berubah atau bergerak ke tempat yang dimaksud.

Pengguna yang mengaktifkan `prefers-reduced-motion` akan mendapatkan semua durasi di-set ke `0.001ms`.

---

## 10. Ikonografi

JOBHOO menggunakan **SVG inline** dari set ikon Lucide/Heroicons style (24x24 viewBox, `stroke="currentColor"`, `stroke-width="2"`).

Icon di job card metadata:
- 📄 Seniority level → document/file icon
- 🏷️ Kategori → clock/circle icon
- 👤 Employment type → person icon
- 🏢 Work arrangement → building icon
- 📍 Lokasi → map pin
- 💰 Salary → sun/glow icon

---

## 11. Scrollbar Kustom

JOBHOO meng-override scrollbar bawaan browser:

- **Lebar:** 8px
- **Track:** Transparan
- **Thumb:** `orange-500` (`#d96600`)
- **Thumb hover:** `orange-400`
- **Border-radius:** `radius-sm` (10px)

---

## 12. Hal yang Tidak Boleh Dilakukan

- ❌ Jangan gunakan warna latar putih atau abu-abu terang
- ❌ Jangan tambah typeface selain Lexend
- ❌ Jangan gunakan oranye sebagai background area besar
- ❌ Jangan gunakan border-radius 0 (kotak penuh) pada komponen interaktif
- ❌ Jangan tambah shadow berwarna selain oranye glow dan navy shadow
- ❌ Jangan gunakan animasi bounce, elastic, atau dekoratif
- ❌ Jangan letakkan teks gelap di atas background navy (kontras terlalu rendah)
- ❌ Jangan hard-code hex/pixel langsung di CSS — selalu gunakan token dari `tokens.css`
