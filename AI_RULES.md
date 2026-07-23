# AI Interaction Rules - Teriakan Bocah

Berkas ini berisi aturan penting bagi asisten AI (Antigravity atau agen AI lainnya) yang berinteraksi dengan repositori ini. AI wajib membaca berkas ini di awal setiap sesi untuk memastikan efisiensi penggunaan token dan kenyamanan pengguna.

## 1. Aturan Penghematan Token AI
- **Tutup Tab Tidak Relevan**: Jika AI mendeteksi ada lebih dari 3 berkas terbuka yang tidak relevan dengan tugas aktif saat ini, AI wajib mengingatkan pengguna di awal respons agar menutup tab yang tidak digunakan tersebut demi menghemat token konteks.
- **Sunting Parsial (Minimal Chunks)**: Selalu gunakan perubahan parsial (seperti `replace_file_content` atau penulisan kode terarah) daripada menulis ulang seluruh berkas. Jangan pernah menulis ulang kode yang tidak berubah.
- **Ringkas & Padat**: Berikan penjelasan yang singkat, langsung pada intinya, dan hindari penjelasan bertele-tele yang membuang kuota token.
- **Baca AI_RULES.md Lebih Dulu**: Di awal setiap sesi, AI wajib membaca berkas ini sebelum membuka berkas lain manapun.
- **Andalkan Skema di Berkas Ini**: Untuk memahami struktur data, gunakan skema database di Seksi 5 terlebih dahulu. Buka berkas model hanya jika ada ketidaksesuaian atau informasi yang tidak tercakup.

## 2. Preferensi Komunikasi Pengguna
- **Bahasa**: Gunakan Bahasa Indonesia yang sopan, ramah, dan profesional dalam semua interaksi dan dokumentasi.
- **Larangan Emoji**: Dilarang menggunakan emoji apa pun dalam pesan obrolan, komentar kode, berkas HTML, maupun dokumen panduan.

## 2b. Preferensi Desain UI
- **Hindari Ciri Khas UI Generik Buatan AI**: Buat tampilan yang bersih, natural, profesional, dan sesuai konteks aplikasi; hindari dekorasi yang terasa dibuat-buat atau berlebihan.
- **Ikon Harus Memiliki Fungsi Jelas**: Jangan menambahkan ikon, badge, ilustrasi, sparkle, atau ornamen lain jika tidak membantu pengguna memahami fungsi atau melakukan tindakan.
- **Jangan Menambah Elemen Tanpa Kebutuhan**: Hindari gradient, glow, kartu berlebihan, teks promosi, dan komponen dekoratif yang tidak diminta atau tidak mendukung hierarki informasi.
- **Utamakan Kesederhanaan**: Gunakan teks atau kontrol sederhana apabila sudah cukup jelas. Setiap elemen visual baru harus memiliki alasan fungsional.

## 2a. Kewajiban Pengguna (AI Wajib Mengingatkan)
- **Update CURRENT_TASK.md**: Sebelum memulai sesi baru, pengguna wajib mengisi `CURRENT_TASK.md` di root proyek dengan tugas yang akan dikerjakan dan file yang relevan. Jika berkas ini kosong atau tidak diperbarui, AI harus mengingatkan pengguna sebelum mulai bekerja.
- **Sebut File Spesifik**: Saat meminta perubahan, sebutkan nama file yang dituju secara eksplisit (contoh: "perbaiki `work_simple.html` bagian filter") agar AI tidak perlu menelusuri semua file.
- **Tutup Tab yang Tidak Relevan**: Pastikan hanya file yang relevan dengan tugas saat ini yang dibuka di editor.

## 3. Alur Kerja Pengembangan
- **Kompilasi Mandiri**: Ingatkan pengguna untuk melakukan kompilasi ulang (`go build -o app cmd/main.go`) dan restart service (`sudo systemctl restart teriakan-bocah.service`) secara mandiri setelah setiap perubahan selesai diterima (accepted).

## 3a. Manajemen CURRENT_TASK.md (Tanggung Jawab AI)
- **Awal Sesi**: Segera setelah pengguna menjelaskan apa yang ingin dikerjakan, AI wajib update `CURRENT_TASK.md` (tugas, file relevan, konteks).
- **Tengah Sesi**: Jika topik atau tugas bergeser signifikan, AI update `CURRENT_TASK.md` tanpa perlu diminta.
- **Protokol Akhir Sesi**: Jika pengguna mengatakan kata kunci **"end session"**, AI wajib melakukan urutan berikut sebelum mengakhiri:
  1. Update `CURRENT_TASK.md` bagian "Riwayat Sesi Terakhir" dengan ringkasan apa yang dikerjakan.
  2. Update bagian "Tugas yang Sedang Dikerjakan" dengan sisa tugas yang belum selesai (jika ada).
  3. Konfirmasi kepada pengguna bahwa berkas sudah diperbarui dan sesi aman untuk ditutup.

## 4. Konteks Proyek

### Gambaran Umum
- **Nama Aplikasi**: Teriakan Bocah (internal workspace management)
- **Tujuan**: Platform internal multi-modul untuk manajemen task, perawatan AC, dan manajemen pengguna
- **Target Pengguna**: Staf dan manajemen internal perusahaan

### Stack Teknologi
- **Backend**: Go 1.24, net/http standar (tanpa framework eksternal)
- **Database**: SQLite via `modernc.org/sqlite`, file: `ac_maintenance.db`
- **Frontend**: HTML template bawaan Go (`html/template`), tanpa framework JS
- **Auth**: `golang.org/x/crypto` untuk hashing password, session berbasis cookie
- **Service**: Berjalan sebagai systemd service `teriakan-bocah.service`, port default `8080`

### Struktur Direktori
```
cmd/
  main.go              - Entry point & semua routing HTTP
  seed/                - Seed data awal
internal/
  handlers/
    handlers.go        - Handler: auth, AC module, admin user
    work_handlers.go   - Handler: Work System (task manager)
  models/
    record.go          - Model: data AC maintenance
    user.go            - Model: data user
    work.go            - Model: data task/work
  database/            - Inisialisasi & koneksi DB
templates/
  base.html            - Layout dasar (digunakan sebagian halaman)
  sidebar_layout.html  - Layout dengan sidebar (digunakan modul utama)
  login.html           - Halaman login
  dashboard.html       - Halaman portal/dashboard utama
  index.html           - Halaman daftar record AC
  form.html, edit_form.html, record_item.html - Form & partial AC
  admin_users.html, edit_user.html - Halaman admin user
  profile.html         - Halaman profil & ganti password
  work_simple.html, work_dashboard.html, work_edit.html - Halaman Work System
  work_inbox.html, work_projects.html, task_drawer_partial.html - Partial Work
static/               - Aset statis (CSS, JS, gambar)
```

### Modul & Fitur
- **Modul `ac`**: AC Maintenance Tracker - pencatatan & riwayat perawatan AC
- **Modul `work`**: Work System / Task Manager - to-do list dengan sub-detail dan catatan kendala
- **Admin**: Manajemen pengguna (hanya role Master) - tambah, edit, nonaktifkan, reset password
- **Auth Middleware**: `AuthMiddleware` (wajib login), `ModuleMiddleware("ac"/"work")` (cek akses modul), `MasterMiddleware` (hanya Master)
- **Role**: `master` (admin penuh), user biasa (akses terbatas per modul yang diaktifkan)

## 5. Skema Database (SQLite - `ac_maintenance.db`)

> Gunakan referensi ini sebelum membuka file model. Perbarui seksi ini jika ada perubahan skema.

### Tabel `users`
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | INTEGER | Primary key |
| username | TEXT | Nama login |
| password | TEXT | Bcrypt hash |
| role | TEXT | `master` atau `user` |
| created_at | DATETIME | |
| last_login | DATETIME | Nullable |
| is_active | BOOLEAN | Default true |
| allowed_modules | TEXT | Comma-separated: `ac,work` |

### Tabel `permissions`
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | INTEGER | Primary key |
| user_id | INTEGER | FK ke users |
| app_name | TEXT | Nama modul, contoh: `ac` |
| capabilities | TEXT | Comma-separated: `view,create,delete` |

### Tabel `maintenance_records` (Modul AC)
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | INTEGER | Primary key |
| room | TEXT | Nama ruangan |
| unit | TEXT | Nama/identitas unit AC |
| activity | TEXT | Jenis aktivitas maintenance |
| date | DATETIME | Tanggal maintenance |
| status | TEXT | `Scheduled`, `Completed`, `Pending` |
| notes | TEXT | Catatan bebas |
| next_service_date | DATETIME | Nullable, jadwal service berikutnya |

### Tabel `tasks` (Modul Work)
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | INTEGER | Primary key |
| title | TEXT | Judul tugas |
| outcome | TEXT | Target hasil |
| estimate | INTEGER | Estimasi dalam menit |
| status | TEXT | `inbox`, `todo`, `doing`, `done` |
| project_id | INTEGER | Nullable FK ke projects |
| scheduled_date | DATETIME | Nullable |
| assigned_to | TEXT | Kosong jika belum ditugaskan |
| due_date | DATETIME | Nullable, tenggat waktu |
| priority | TEXT | `low`, `medium`, `high` |
| is_archived | BOOLEAN | |
| created_at | DATETIME | |
| updated_at | DATETIME | |
| completed_at | DATETIME | Nullable |

### Tabel `task_details`
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | INTEGER | Primary key |
| task_id | INTEGER | FK ke tasks |
| description | TEXT | Deskripsi langkah kerja |
| progress | TEXT | Catatan progres saat ini |
| obstacle | TEXT | Catatan kendala (jika ada) |
| created_at | DATETIME | |
| updated_at | DATETIME | |

### Tabel `obstacles`
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | INTEGER | Primary key |
| task_id | INTEGER | FK ke tasks |
| description | TEXT | Deskripsi kendala |
| status | TEXT | `open`, `resolved` |
| created_at | DATETIME | |
| resolved_at | DATETIME | Nullable |

### Tabel `projects`
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | INTEGER | Primary key |
| name | TEXT | Nama proyek |
| outcome_goal | TEXT | Target hasil proyek |
| status | TEXT | Status proyek |
| created_at | DATETIME | |
