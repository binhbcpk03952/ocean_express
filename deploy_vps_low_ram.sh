#!/bin/bash
echo "🚀 Bắt đầu quá trình cập nhật Ocean Express (Chế độ tiết kiệm RAM)..."

# 1. Dọn dẹp RAM và ổ cứng trước khi build
echo "🧹 Dọn dẹp bộ nhớ đệm và các container rác..."
docker system prune -f

# 2. Build backend trước (tuần tự để không tranh chấp RAM)
echo "⚙️ Đang build Backend (API)..."
docker compose build api

# 3. Build frontend với giới hạn RAM cho Node.js (1GB)
echo "⚙️ Đang build Frontend (Admin) với giới hạn RAM..."
# Chèn NODE_OPTIONS vào quá trình build
export NODE_OPTIONS="--max-old-space-size=1024"
docker compose build admin

# 4. Khởi động lại các service
echo "🔄 Đang khởi động lại hệ thống..."
docker compose up -d --remove-orphans

echo "✅ Hoàn tất triển khai cập nhật mới nhất an toàn trên VPS 2GB RAM!"
