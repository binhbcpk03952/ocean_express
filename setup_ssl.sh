#!/bin/bash

echo "Cài đặt Certbot và cấu hình SSL (HTTPS)..."

# Cập nhật và cài đặt certbot
sudo apt-get update
sudo apt-get install -y certbot python3-certbot-nginx

# Chạy certbot cho các tên miền
# (Certbot sẽ tự động cấu hình lại Nginx để chuyển HTTP sang HTTPS)
sudo certbot --nginx --non-interactive --agree-tos -m admin@bcbdev.id.vn \
    -d oceanexpress.bcbdev.id.vn \
    -d admin.oceanexpress.bcbdev.id.vn \
    -d shop.oceanexpress.bcbdev.id.vn \
    -d tracking.oceanexpress.bcbdev.id.vn \
    -d api.oceanexpress.bcbdev.id.vn

echo "Đã cấu hình SSL thành công! Website của bạn giờ đã có HTTPS."
