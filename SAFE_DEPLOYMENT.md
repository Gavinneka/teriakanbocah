# ⚠️ Safe Deployment - VPS dengan Web Lain

## 🔍 Langkah 1: CEK DULU - Jangan Langsung Install!

### 1.1 Cek Web Server yang Sudah Ada
```bash
# Cek apakah Nginx sudah terinstall
nginx -v

# Cek status Nginx
sudo systemctl status nginx

# Cek Apache (jika ada)
apache2 -v
sudo systemctl status apache2
```

### 1.2 Cek Port yang Sudah Dipakai
```bash
# Lihat semua port yang aktif
sudo netstat -tulpn | grep LISTEN

# Atau pakai ss
sudo ss -tulpn | grep LISTEN

# Cek port 80 dan 443 khusus
sudo lsof -i :80
sudo lsof -i :443
```

### 1.3 Cek Nginx Sites yang Sudah Ada
```bash
# Lihat semua site yang aktif
ls -la /etc/nginx/sites-enabled/

# Lihat isi config yang ada
cat /etc/nginx/sites-enabled/default
# atau
cat /etc/nginx/sites-enabled/nama-site-lain
```

### 1.4 Cek Systemd Services yang Ada
```bash
# Lihat semua service yang running
sudo systemctl list-units --type=service --state=running | grep -v '@'

# Cek service Go apps yang mungkin ada
sudo systemctl list-units --type=service | grep -i go
```

---

## ✅ Langkah 2: STRATEGI AMAN - Pilih Salah Satu

### **Opsi A: Pakai Port Berbeda (PALING AMAN)**

Jika web lain pakai port 8080, Anda pakai port lain (misal 8081, 8082, dll)

**1. Edit `cmd/main.go` - Ganti Port Default**
```go
port := os.Getenv("PORT")
if port == "" {
    port = "8081"  // GANTI dari 8080 ke 8081
}
```

**2. Build ulang**
```bash
go build -o app cmd/main.go
```

**3. Buat service dengan port berbeda**
```ini
[Service]
Environment="PORT=8081"  # Port berbeda!
```

**4. Nginx config dengan subdomain/path berbeda**
```nginx
# Opsi 1: Pakai subdomain
server {
    listen 80;
    server_name teriakan.yourdomain.com;  # Subdomain berbeda
    
    location / {
        proxy_pass http://localhost:8081;  # Port berbeda
        # ... proxy settings
    }
}

# Opsi 2: Pakai path
server {
    listen 80;
    server_name yourdomain.com;
    
    # Web lain tetap di /
    location / {
        # config web lain
    }
    
    # Teriakan Bocah di /teriakan
    location /teriakan/ {
        proxy_pass http://localhost:8081/;
        # ... proxy settings
    }
}
```

---

### **Opsi B: Tambah Config Nginx Baru (Jika Pakai Subdomain)**

**1. JANGAN edit file yang sudah ada!**

**2. Buat file config BARU**
```bash
sudo nano /etc/nginx/sites-available/teriakan-bocah
```

**3. Isi config (JANGAN sentuh config lain)**
```nginx
server {
    listen 80;
    server_name teriakan.yourdomain.com;  # Subdomain BARU
    
    location / {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

**4. Enable site BARU**
```bash
sudo ln -s /etc/nginx/sites-available/teriakan-bocah /etc/nginx/sites-enabled/

# TEST dulu sebelum restart!
sudo nginx -t

# Kalau OK, baru restart
sudo systemctl reload nginx  # Pakai reload, bukan restart!
```

---

## 🛡️ Langkah 3: SAFETY CHECKLIST

### Sebelum Install Apapun:
- [ ] Backup config Nginx yang ada
  ```bash
  sudo cp -r /etc/nginx/sites-enabled /etc/nginx/sites-enabled.backup
  ```
- [ ] Catat semua port yang sudah dipakai
- [ ] Catat semua service yang sudah running
- [ ] Tanya/cek dengan admin VPS (jika ada)

### Saat Install:
- [ ] **JANGAN** pakai `sudo systemctl restart nginx` (pakai `reload`)
- [ ] **JANGAN** edit file config yang sudah ada
- [ ] **JANGAN** pakai port yang sama dengan web lain
- [ ] **SELALU** test config dulu: `sudo nginx -t`

### Setelah Install:
- [ ] Cek web lain masih jalan: `curl http://localhost:PORT_LAMA`
- [ ] Cek service lain masih running: `sudo systemctl status nama-service-lain`
- [ ] Test akses web lain dari browser

---

## 🚨 Troubleshooting - Jika Ada Masalah

### Jika Nginx Error Setelah Reload
```bash
# Restore backup
sudo rm /etc/nginx/sites-enabled/teriakan-bocah
sudo systemctl reload nginx

# Cek error detail
sudo nginx -t
sudo tail -f /var/log/nginx/error.log
```

### Jika Port Conflict
```bash
# Lihat siapa yang pakai port
sudo lsof -i :8080

# Kill process (HATI-HATI! Pastikan bukan web lain)
# sudo kill -9 PID
```

### Jika Service Tidak Start
```bash
# Lihat log error
sudo journalctl -u teriakan-bocah -n 50

# Cek port sudah dipakai atau belum
sudo netstat -tulpn | grep :8081
```

---

## 📋 Rekomendasi Setup AMAN

```
Web Lain (sudah ada):
- Domain: yourdomain.com
- Port: 8080
- Service: existing-app.service
- Nginx config: /etc/nginx/sites-enabled/default

Teriakan Bocah (baru):
- Domain: teriakan.yourdomain.com (SUBDOMAIN BARU)
- Port: 8081 (PORT BERBEDA)
- Service: teriakan-bocah.service (NAMA BERBEDA)
- Nginx config: /etc/nginx/sites-enabled/teriakan-bocah (FILE BARU)
```

---

## ✅ Command Aman untuk Cek Sebelum Deploy

```bash
# 1. Cek semua yang running
sudo systemctl list-units --type=service --state=running

# 2. Cek semua port
sudo ss -tulpn

# 3. Cek Nginx config
sudo nginx -t
ls -la /etc/nginx/sites-enabled/

# 4. Backup sebelum mulai
sudo cp -r /etc/nginx /etc/nginx.backup.$(date +%Y%m%d)
```

**INGAT: Kalau ragu, JANGAN dulu! Tanya dulu atau backup dulu!** 🛡️
