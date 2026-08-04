#!/bin/bash
echo "🚀 Bắt đầu quá trình cập nhật Ocean Express..."

# Build lại các image có code thay đổi (admin, api)
docker-compose build

# Khởi động lại các service (zero-downtime tùy khả năng của Nginx, nhưng sẽ nhanh chóng recreate)
docker-compose up -d --remove-orphans

echo "🧹 Dọn dẹp images/containers rác để giải phóng ổ cứng..."
# Dọn dẹp các dangling images (những image sinh ra sau quá trình build đè lên image cũ)
docker image prune -f

echo "✅ Hoàn tất triển khai cập nhật mới nhất và giải phóng tài nguyên!"
