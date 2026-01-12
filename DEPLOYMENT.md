# 🚀 Deployment Guide: VPS Deployment

## Prerequisites
- ✅ VPS dengan SSH access
- ✅ GitHub repository
- ✅ Domain (opsional, bisa pakai IP)

---

## Step 1: Persiapan Lokal

### 1.1 Push ke GitHub
```bash
# Di folder project (d:\Teriakan Bocah)
git init
git add .
git commit -m "Initial commit - AC Tracker & Work System"
git branch -M main
git remote add origin https://github.com/USERNAME/REPO_NAME.git
git push -u origin main
```

### 1.2 Buat `.gitignore`
```
app.exe
*.db
.env
```

---

## Step 2: Setup VPS

### 2.1 SSH ke VPS
```bash
ssh user@your-vps-ip
```

### 2.2 Install Go (jika belum ada)
```bash
# Download Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz

# Extract
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# Set PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
```

### 2.3 Install Git (jika belum ada)
```bash
sudo apt update
sudo apt install git -y
```

---

## Step 3: Clone & Build

### 3.1 Clone Repository
```bash
cd ~
git clone https://github.com/USERNAME/REPO_NAME.git
cd REPO_NAME
```

### 3.2 Build Aplikasi
```bash
go build -o app cmd/main.go
```

### 3.3 Test Run
```bash
./app
# Ctrl+C untuk stop
```

---

## Step 4: Setup Systemd Service (Auto-restart)

### 4.1 Buat Service File
```bash
sudo nano /etc/systemd/system/teriakan-bocah.service
```

### 4.2 Isi File Service
```ini
[Unit]
Description=Teriakan Bocah - AC Tracker & Work System
After=network.target

[Service]
Type=simple
User=YOUR_USERNAME
WorkingDirectory=/home/YOUR_USERNAME/REPO_NAME
ExecStart=/home/YOUR_USERNAME/REPO_NAME/app
Restart=always
RestartSec=5
Environment="PORT=8080"

[Install]
WantedBy=multi-user.target
```

**Ganti:**
- `YOUR_USERNAME` dengan username VPS Anda
- `REPO_NAME` dengan nama folder repository

### 4.3 Enable & Start Service
```bash
sudo systemctl daemon-reload
sudo systemctl enable teriakan-bocah
sudo systemctl start teriakan-bocah

# Cek status
sudo systemctl status teriakan-bocah
```

---

## Step 5: Setup Nginx (Reverse Proxy)

### 5.1 Install Nginx
```bash
sudo apt install nginx -y
```

### 5.2 Buat Config Nginx
```bash
sudo nano /etc/nginx/sites-available/teriakan-bocah
```

### 5.3 Isi Config
```nginx
server {
    listen 80;
    server_name your-domain.com;  # atau IP VPS

    location / {
        proxy_pass http://localhost:8080;
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

### 5.4 Enable Site
```bash
sudo ln -s /etc/nginx/sites-available/teriakan-bocah /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

---

## Step 6: Setup Firewall

```bash
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS (untuk nanti)
sudo ufw enable
```

---

## Step 7: Setup SSL (HTTPS) - Opsional tapi Recommended

### 7.1 Install Certbot
```bash
sudo apt install certbot python3-certbot-nginx -y
```

### 7.2 Get SSL Certificate
```bash
sudo certbot --nginx -d your-domain.com
```

---

## Step 8: Maintenance Commands

### Lihat Log
```bash
sudo journalctl -u teriakan-bocah -f
```

### Restart Service
```bash
sudo systemctl restart teriakan-bocah
```

### Update Aplikasi (setelah push ke GitHub)
```bash
cd ~/REPO_NAME
git pull
go build -o app cmd/main.go
sudo systemctl restart teriakan-bocah
```

### Backup Database
```bash
cp ~/REPO_NAME/app.db ~/backup-$(date +%Y%m%d).db
```

---

## Troubleshooting

### Port sudah dipakai
```bash
sudo lsof -i :8080
sudo kill -9 PID_NUMBER
```

### Service gagal start
```bash
sudo journalctl -u teriakan-bocah -n 50
```

### Nginx error
```bash
sudo nginx -t
sudo tail -f /var/log/nginx/error.log
```

---

## 🎯 Akses Aplikasi

Setelah semua selesai, akses di:
- **HTTP**: `http://your-vps-ip` atau `http://your-domain.com`
- **HTTPS**: `https://your-domain.com` (jika sudah setup SSL)

Login dengan akun Master yang sudah dibuat!

---

## 📝 Notes

1. **Database**: SQLite file (`app.db`) akan otomatis dibuat di folder aplikasi
2. **Backup**: Rutin backup file `app.db` 
3. **Update**: Setiap update code, jangan lupa `git pull` dan `go build` lalu restart service
4. **Security**: Ganti password default Master setelah deployment pertama
