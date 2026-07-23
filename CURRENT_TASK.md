# Status Tugas Saat Ini

> Perbarui berkas ini setiap kali memulai sesi baru bersama AI.
> AI wajib membaca berkas ini sebelum memulai pekerjaan apapun.

## Tugas yang Sedang Dikerjakan

(kosong — seluruh tugas sesi telah selesai)

## File yang Relevan

- `templates/sidebar_layout.html`
- `templates/dashboard.html`
- `templates/index.html`
- Template halaman lain yang memakai pola visual serupa.

## Konteks Tambahan

- Antarmuka utama dibuat lebih datar dan tenang: bayangan besar serta animasi pulse dinonaktifkan, radius besar dikurangi, dan istilah menu diseragamkan ke Bahasa Indonesia.
- Beranda diubah dari kartu promosi menjadi daftar navigasi sederhana. Seluruh test, vet, build, dan pemeriksaan whitespace pada file terkait lulus.
- Notifikasi reset password menggunakan `hx-on::after-request` dan hanya tampil jika request berhasil; test, vet, dan build lulus.
- Perbaikan `/improvement`: nama layout template diselaraskan menjadi `sidebar_layout.html`; build dan seluruh test lulus.
- Build & restart service: `go build -o app ./cmd/main.go` lalu `sudo systemctl restart teriakan-bocah.service`
- Service berjalan di port **18001**

## Perbaikan UI yang Ditunda (Lanjutkan Sesi Berikutnya)

(kosong)

## Riwayat Sesi Terakhir

**Tanggal**: 2026-07-22 — Menghilangkan ciri UI generik buatan AI pada layout utama: menyederhanakan beranda menjadi daftar navigasi, meratakan permukaan visual, mengurangi radius, menonaktifkan animasi dekoratif, dan mengganti istilah campuran Inggris dengan label Indonesia yang lugas. Seluruh test, vet, build, dan pemeriksaan whitespace file terkait lulus.

**Tanggal**: 2026-07-22 — Menambahkan preferensi desain permanen ke `AI_RULES.md`: hindari ciri UI generik buatan AI, ikon atau badge tanpa fungsi jelas, serta dekorasi berlebihan seperti gradient, glow, dan kartu yang tidak dibutuhkan. Utamakan UI sederhana, natural, profesional, dan fungsional. Tidak ada sisa tugas.

**Tanggal**: 2026-07-22 — Melakukan pemeriksaan bug umum; `go test ./...`, `go vet ./...`, dan build berhasil. Memperbaiki notifikasi sukses reset password pada halaman manajemen pengguna dengan mengganti atribut event yang tidak valid menjadi `hx-on::after-request`; notifikasi kini hanya muncul ketika request HTMX berhasil. Tidak ada sisa tugas.

**Tanggal**: 2026-07-22 — Memperbaiki bug hapus subtask pada Task Drawer dengan mengganti hx-include yang salah/duplikat id menjadi hx-vals. Memperbaiki redirect form edit/delete subtask pada halaman edit task agar tetap berada di halaman edit task saat disubmit.

**Tanggal**: 2026-07-20 — Memperbaiki dashboard **Work** agar kendala yang sudah resolved tidak lagi dihitung atau ditampilkan sebagai kendala terbuka pada jumlah per task, filter **Kendala**, dan ringkasan global. Menambahkan test regresi yang membedakan kendala terbuka dan resolved. `go test ./...`, test khusus filter, dan build berhasil. Tidak ada sisa tugas; build serta restart service produksi tetap dilakukan pengguna secara mandiri.

**Tanggal**: 2026-07-19 — Memperbaiki halaman **Improvement Log** (`/improvement`) yang gagal dibuka. Penyebabnya adalah pemanggilan nama layout `sidebar_layout` yang tidak sesuai dengan nama template terdaftar `sidebar_layout.html` pada halaman daftar dan edit. Pemanggilan layout telah diselaraskan, error render kini dicatat oleh handler, dan `go test ./...` serta build berhasil. Tidak ada sisa tugas dari perbaikan ini.

**Tanggal**: 2026-07-19 — Membuat modul baru **Improvement Log** (`/improvement`) untuk mencatat perbaikan/improvement yang sudah dilakukan di kampus. Fitur: summary card total item & biaya, form tambah collapsible, filter tab (Semua/Selesai/Proses/Ditunda), tabel dengan badge status + hover action edit/hapus, format Rupiah otomatis. File baru: `models/improvement.go`, `handlers/improvement_handlers.go`, `templates/improvement.html`, `templates/improvement_edit.html`. File dimodifikasi: `db.go` (tabel baru), `main.go` (5 routes baru), `sidebar_layout.html` (nav link + breadcrumb). Build sukses, service restart normal.

**Tanggal**: 2026-07-18 — Menyederhanakan antarmuka UI/UX di seluruh sistem. Menyatukan seluruh modul (termasuk Portal Utama dan Profil) menggunakan layout sidebar. Memindahkan manajemen rincian subtask dan kendala dari tabel utama ke Side Drawer via HTMX untuk menghindari muat ulang halaman. Menyembunyikan formulir masukan utama ke dalam menu lipat (collapsible) menggunakan Alpine.js. Sukses melakukan build dan restart service.

**Tanggal**: 2026-05-27 — Memvalidasi dan menyempurnakan migrasi layout, menu collapsible, dan handler modul Backlog/Projects. Mengimplementasikan pendeteksian rute dinamis pada menu samping (sidebar active state) dan remah roti (dynamic breadcrumbs) menggunakan Alpine.js agar tampilan navigasi dan header responsif terhadap halaman aktif.

**Tanggal**: 2026-05-20 — Migrasi emoji khas AI pada berkas HTML templates ke ikon SVG modern dan interaktif (sidebar, dashboard, profile, edit user) untuk menjaga estetika premium dan kepatuhan terhadap AI_RULES.md.

**Tanggal**: 2026-05-20 — Selesaikan 4 pending UI fixes: #5 Quick Add Form compact (padding dikurangi), #7 work_simple.html migrasi dari base.html → sidebar_layout.html (konsistensi nav), #8 Tombol Create sidebar kini href="/work#quick-add", GlobalObstacles banner besar → badge compact animate-pulse inline di filter bar. Build sukses, service restart normal.

**Tanggal**: 2026-05-20 — Hapus tombol "Tambah Task" redundan dari sidebar (`sidebar_layout.html`) sesuai dengan Opsi C untuk menghilangkan duplikasi visual dan navigasi dengan tombol Quick Add inline di dashboard.

**Tanggal**: 2026-05-20 — Perbaikan UI work_simple.html: hapus header verbose (Fix #1), hapus kolom Kendala duplikat (Fix #2), fix timestamp Update tanpa tanggal → kini tampil "02 Jan 15:04" (Fix #3). Service aktif di port 18001, memory 3.3M. Sesi ditutup normal.

**Tanggal**: 2026-05-20 — Implementasi subtask done checklist + obstacle resolve tracking di Work System. Build sukses, service aktif.
