# Buku Panduan Workspace - Teriakan Bocah

Selamat datang di modul panduan penggunaan sistem **Workspace** ("Teriakan Bocah").
Panduan ini dirancang khusus untuk memandu staf dan manajemen dalam menggunakan fitur-fitur yang ada di sistem ini.

**Akses Cepat**
Sistem dapat diakses lewat web browser kapan saja dan di mana saja menggunakan tautan/IP yang telah diberikan oleh tim IT Anda.

---

## 1. Modul Manajemen User (Admin)

Modul ini adalah pusat kendali semua akses ke aplikasi. Hanya user dengan hak akses "Master" yang bisa mengakses halaman ini.

**Langkah menambahkan User Baru:**
1. Klik menu **Admin Panel** di Dashboard utama (atau Navigasi Atas).
2. Di halaman Manajemen Pengguna, klik tombol biru **Tambah User**.
3. Masukkan Username, buatkan Password Awal, dan tentukan **Role**:
   * **User Biasa:** Untuk teknisi atau staf operasional.
   * **Master Admin:** Memiliki hak istimewa seperti Anda.
4. Jangan lupa centang modul apa saja yang boleh diakses (AC / Work) sebelum menekan tombol **Simpan**.

**Reset Password & Nonaktifkan User:**
* Jika ada bawahan yang lupa *password*, Anda cukup menekan tombol jingga **Reset Pass**. Password-nya otomatis akan dirubah ke angka standar **`123456`**.
* Jika staf _resign_, Anda dapat menekan **Nonaktifkan** agar dirinya tidak bisa masuk aplikasi lagi, tanpa harus menghapus (*Delete*) riwayat tugasnya.

---

## 2. Modul Work System (Sistem Kerja / Task Manager)

Ini adalah fitur daftar tugas (to-do list) untuk mencatat aktivitas harian hingga proyek jangka panjang.

* **Cara Menambah Tugas (Quick Add):**
  Ketik nama tugas di kotak input utama, tentukan tenggat waktu jika diperlukan, kemudian klik tombol **Tambah Task**. Tugas baru akan langsung tercatat ke dalam sistem.

* **List Detail Kerja (Langkah Detail):**
  Klik tombol edit atau judul tugas untuk masuk ke halaman detail tugas secara penuh. Di sini, Anda dapat memecah tugas utama menjadi beberapa langkah rincian detail dengan cepat melalui formulir tambah cepat sekali input (cukup masukkan deskripsi kerja dan tekan Tambah). Setelah ditambahkan, Anda dapat mengeklik ikon edit untuk mengisi catatan progres saat ini serta kendala secara inline (langsung di tempat) tanpa memuat ulang halaman.

* **Manajemen Kendala terintegrasi:**
  Fitur kendala kini diintegrasikan secara penuh ke dalam masing-masing butir detail kerja. Jika suatu langkah memiliki kendala, Anda cukup menuliskannya langsung pada butir detail kerja tersebut melalui inline editing. Tugas yang memiliki detail berkendala aktif akan secara otomatis memicu indikator peringatan dan menampilkan ringkasan kendala tersebut di halaman utama dashboard agar mendapat perhatian prioritas.

---

## 3. Modul AC Maintenance Tracker

Untuk pelacakan pemeliharaan AC guna menekan risiko AC bocor atau rusak berat.

* **Tambah Record Baru:**
  Di halaman AC Tracker, isi form yang tersedia secara berurutan. Anda dapat mengeklik pilihan "Lainnya..." jika nama ruangan Anda tak tersedia di opsi cepat _dropdown_. 
* **Tanggal & Jadwal Selanjutnya:**
  Catat waktu _maintenance_ hari ini. Selain itu, **jangan lupa** untuk mengisi form "Service Berikutnya" sebagai alarm. Ketika AC sudah dijadwalkan ulang bulan depan, secara otomatis sistem akan menunjukkan notifikasi bercentang pada record aset tersebut (`NEXT: [Tanggal]`).

* **Filter History:**
  Untuk menelusuri AC ruangan spesifik di bulan lalu, gunakan kotak Filter di atas tabel, pilih range waktu, pilih nama ruang (contoh "Lab"), dan sistem akan merekap seluruh _history_-nya saja.

---

**Butuh Bantuan Lain?**
Bila terjadi masalah *Error 500* atau *Website Error*, harap segera hubungi tim Developer internal. Tetap bekerja dengan cermat, fokus, dan biarkan aplikasi membantu _multitasking_ Anda!

> **Ad Maiora Natus Sum** - _Untuk hal-hal yang lebih besar, aku dilahirkan._
