#!/bin/bash
set -e

echo "1. Đang giải nén mã nguồn..."
cd ~/oceanexpress
tar -xzf oceanexpress.tar.gz -C .
# Remove Windows line endings if any
sed -i 's/\r$//' setup_nginx.sh
sed -i 's/\r$//' setup_ssl.sh
chmod +x setup_nginx.sh setup_ssl.sh

echo "2. Đang cấu hình lại domain cho đúng yêu cầu (chỉ dùng main và api)..."
cat << 'EOF' > setup_nginx.sh
#!/bin/bash
MAIN_DOMAIN="oceanexpress.bcbdev.id.vn"
API_DOMAIN="api.oceanexpress.bcbdev.id.vn"

echo "Configuring Nginx for Ocean Express..."

cat <<INNER_EOF > /tmp/oceanexpress.conf
server {
    listen 80;
    server_name $MAIN_DOMAIN;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }
}
INNER_EOF

cat <<INNER_EOF > /tmp/apioceanexpress.conf
server {
    listen 80;
    server_name $API_DOMAIN;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }
}
INNER_EOF

sudo mv /tmp/oceanexpress.conf /etc/nginx/sites-available/oceanexpress.conf
sudo mv /tmp/apioceanexpress.conf /etc/nginx/sites-available/apioceanexpress.conf

sudo ln -sf /etc/nginx/sites-available/oceanexpress.conf /etc/nginx/sites-enabled/
sudo ln -sf /etc/nginx/sites-available/apioceanexpress.conf /etc/nginx/sites-enabled/

sudo nginx -t && sudo systemctl reload nginx
echo "Nginx configured successfully!"
EOF

cat << 'EOF' > setup_ssl.sh
#!/bin/bash
echo "Cài đặt Certbot và cấu hình SSL (HTTPS)..."
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx
sudo certbot --nginx --non-interactive --agree-tos -m admin@bcbdev.id.vn \
    -d oceanexpress.bcbdev.id.vn \
    -d api.oceanexpress.bcbdev.id.vn
echo "Đã cấu hình SSL thành công!"
EOF

chmod +x setup_nginx.sh setup_ssl.sh

echo "3. Đang cài đặt Docker và Nginx (nếu chưa có)..."
sudo apt-get update
sudo apt-get install -y nginx
# Note: Docker is already installed as checked before

echo "4. Đang load Docker Images..."
docker load -i images.tar

echo "5. Đang khởi động hệ thống Backend và Admin..."
# .env is required. I will make sure we use a dummy one if it doesn't exist, but it should be extracted from tar
if [ ! -f .env ]; then
    echo "Tạo file .env mặc định..."
    echo "DB_USER=root" > .env
    echo "DB_PASS=rootpassword" >> .env
    echo "DB_NAME=ocean_express_db" >> .env
    echo "JWT_SECRET=thisismysecretkeyforproduction" >> .env
fi
docker compose up -d

echo "6. Đang thiết lập Nginx và SSL..."
./setup_nginx.sh
./setup_ssl.sh

echo "Hoàn tất toàn bộ cài đặt!"
