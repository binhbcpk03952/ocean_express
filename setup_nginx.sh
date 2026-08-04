#!/bin/bash

# Define domain names
MAIN_DOMAIN="oceanexpress.bcbdev.id.vn"
API_DOMAIN="api.oceanexpress.bcbdev.id.vn"

echo "Configuring Nginx for Ocean Express..."

# Create Nginx configuration for Main Domain (Frontend - Port 3000)
cat <<EOF > /tmp/oceanexpress.conf
server {
    listen 80;
    server_name $MAIN_DOMAIN admin.$MAIN_DOMAIN shop.$MAIN_DOMAIN tracking.$MAIN_DOMAIN;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOF

# Create Nginx configuration for API Domain (Backend - Port 8080)
cat <<EOF > /tmp/apioceanexpress.conf
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
EOF

# Move to Nginx directory
sudo mv /tmp/oceanexpress.conf /etc/nginx/sites-available/oceanexpress.conf
sudo mv /tmp/apioceanexpress.conf /etc/nginx/sites-available/apioceanexpress.conf

# Enable sites
sudo ln -sf /etc/nginx/sites-available/oceanexpress.conf /etc/nginx/sites-enabled/
sudo ln -sf /etc/nginx/sites-available/apioceanexpress.conf /etc/nginx/sites-enabled/

# Test and reload Nginx
sudo nginx -t && sudo systemctl reload nginx

echo "Nginx configured successfully!"
