#!/bin/sh
set -e

echo "========================================"
echo "ProxyPanel 启动脚本"
echo "========================================"

# 初始化Nginx
echo "检查Nginx配置..."
if [ ! -d "/etc/nginx/sites-enabled" ]; then
    mkdir -p /etc/nginx/sites-enabled
fi

if [ ! -d "/var/log/nginx" ]; then
    mkdir -p /var/log/nginx
fi

# 创建默认SSL证书（避免Nginx启动失败）
if [ ! -f "/etc/nginx/ssl/default.crt" ]; then
    mkdir -p /etc/nginx/ssl
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout /etc/nginx/ssl/default.key \
        -out /etc/nginx/ssl/default.crt \
        -subj "/CN=localhost" > /dev/null 2>&1 || true
fi

# 测试Nginx配置
echo "测试Nginx配置..."
nginx -t 2>&1 || true

# 启动Nginx
echo "启动Nginx..."
nginx || true

# 等待Nginx启动
sleep 2

# 启动ProxyPanel应用
echo "启动ProxyPanel..."
cd /app
exec /app/proxypanel