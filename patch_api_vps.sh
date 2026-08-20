#!/bin/bash
echo "🚀 Bắt đầu gửi file patch lên VPS..."
scp api/internal/delivery/http/shop_handler.go api/internal/delivery/http/employee_handler.go ocean_vps@116.118.3.107:~/oceanexpress/api/internal/delivery/http/

echo "🚀 Đang khởi động lại API trên VPS..."
ssh ocean_vps@116.118.3.107 "cd ~/oceanexpress && docker compose build api && docker compose up -d --no-deps api"

echo "✅ Đã patch xong API!"
