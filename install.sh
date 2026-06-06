#!/bin/bash
set -e

#=======================================
# ProxyPanel 一键安装脚本
# 支持 Debian 12 / Ubuntu 22.04
#=======================================

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROJECT_DIR="/opt/proxypanel"
BACKEND_DIR="${PROJECT_DIR}/backend"
FRONTEND_DIR="${PROJECT_DIR}/frontend"

# 输出函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查root权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root权限运行此脚本"
        exit 1
    fi
}

# 检测系统
detect_system() {
    if [ -f /etc/debian_version ]; then
        OS="debian"
        VERSION=$(cat /etc/debian_version)
    elif [ -f /etc/lsb-release ]; then
        OS="ubuntu"
        VERSION=$(grep "DISTRIB_RELEASE" /etc/lsb-release | cut -d= -f2)
    else
        log_error "不支持的系统，仅支持 Debian 12 / Ubuntu 22.04"
        exit 1
    fi
    log_info "检测到系统: $OS $VERSION"
}

# 检测是否为国内服务器（用于选择镜像源）
is_china() {
    if curl -s --connect-timeout 3 https://www.google.com > /dev/null 2>&1; then
        return 1  # 海外
    else
        return 0  # 国内
    fi
}

# 安装依赖
install_dependencies() {
    log_info "安装系统依赖..."

    # 国内服务器使用阿里云 apt 镜像加速
    if is_china; then
        log_info "检测到国内服务器，配置阿里云 apt 镜像源..."
        if [ "$OS" = "ubuntu" ]; then
            cp /etc/apt/sources.list /etc/apt/sources.list.backup 2>/dev/null || true
            cat > /etc/apt/sources.list << 'APTEOF'
deb http://mirrors.aliyun.com/ubuntu/ jammy main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ jammy-updates main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ jammy-backports main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ jammy-security main restricted universe multiverse
APTEOF
        elif [ "$OS" = "debian" ]; then
            cp /etc/apt/sources.list /etc/apt/sources.list.backup 2>/dev/null || true
            cat > /etc/apt/sources.list << 'APTEOF'
deb http://mirrors.aliyun.com/debian/ bookworm main contrib non-free non-free-firmware
deb http://mirrors.aliyun.com/debian/ bookworm-updates main contrib non-free non-free-firmware
deb http://mirrors.aliyun.com/debian-security/ bookworm-security main contrib non-free non-free-firmware
APTEOF
        fi
    fi

    apt-get update -y

    # 安装基础工具 + 编译必需的 gcc/build-essential + acme.sh 需要的 socat/crontab
    apt-get install -y \
        curl \
        wget \
        git \
        unzip \
        sqlite3 \
        nginx \
        openssl \
        ca-certificates \
        gnupg \
        lsb-release \
        gcc \
        build-essential \
        socat \
        cron

    # 确保 cron 服务运行（acme.sh 自动续签需要）
    systemctl enable cron || true
    systemctl start cron || true

    log_success "系统依赖安装完成"
}

# 安装Go
install_go() {
    log_info "安装Go环境..."

    # 检查是否已安装Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        log_warning "Go已安装: $GO_VERSION"
        return
    fi

    # 下载Go 1.21（自动选择最快源）
    GO_VERSION="1.21.6"
    GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"

    # 根据服务器位置选择下载源
    if is_china; then
        # 国内服务器 — 华为云镜像（下载 Go tarball 最快）
        GO_URL="https://repo.huaweicloud.com/golang/${GO_TARBALL}"
        log_info "下载Go ${GO_VERSION}（国内服务器，使用华为云镜像）..."
    else
        # 海外服务器 — Google 官方源最快
        GO_URL="https://dl.google.com/go/${GO_TARBALL}"
        log_info "下载Go ${GO_VERSION}（海外服务器，使用Google官方源）..."
    fi

    wget -q --show-progress ${GO_URL} -O /tmp/${GO_TARBALL}

    # 解压安装
    tar -C /usr/local -xzf /tmp/${GO_TARBALL}

    # 设置环境变量
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile

    # 创建软链接
    ln -sf /usr/local/go/bin/go /usr/bin/go
    ln -sf /usr/local/go/bin/gofmt /usr/bin/gofmt

    # 清理
    rm -f /tmp/${GO_TARBALL}

    log_success "Go ${GO_VERSION} 安装完成"
}

# 安装Node.js
install_nodejs() {
    log_info "安装Node.js环境..."

    # 检查是否已安装Node.js
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node -v)
        log_warning "Node.js已安装: $NODE_VERSION"
        return
    fi

    # 安装Node.js 18.x
    if is_china; then
        # 国内服务器 — 使用 npmmirror 镜像安装
        log_info "国内服务器，使用 npmmirror 安装 Node.js..."
        curl -fsSL https://cdn.npmmirror.com/binaries/node/v18.20.4/node-v18.20.4-linux-x64.tar.xz -o /tmp/node.tar.xz
        tar -xJf /tmp/node.tar.xz -C /usr/local --strip-components=1
        rm -f /tmp/node.tar.xz
        # 设置 npm 淘宝镜像
        npm config set registry https://registry.npmmirror.com
    else
        curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
        apt-get install -y nodejs
    fi

    # 创建软链接
    ln -sf /usr/bin/node /usr/local/bin/node
    ln -sf /usr/bin/npm /usr/local/bin/npm

    log_success "Node.js安装完成: $(node -v)"
}

# 安装 acme.sh（免费SSL证书工具）
install_acme_sh() {
    log_info "安装 acme.sh..."

    # 检查是否已安装
    if [ -f ~/.acme.sh/acme.sh ]; then
        log_warning "acme.sh 已安装"
        return
    fi

    # 设置默认CA为 Let's Encrypt
    export LE_WORKING_DIR=~/.acme.sh
    export LE_CONFIG_HOME=~/.acme.sh

    # 在线安装 acme.sh（get.acme.sh 被墙时改用 GitHub）
    if curl -s --connect-timeout 5 https://get.acme.sh > /dev/null 2>&1; then
        log_info "通过 get.acme.sh 安装..."
        curl https://get.acme.sh | sh -s email=proxypanel@localhost
    else
        log_warning "get.acme.sh 无法访问，改用 GitHub 源安装..."
        git clone https://github.com/acmesh-official/acme.sh.git /tmp/acme.sh-repo
        cd /tmp/acme.sh-repo
        ./acme.sh --install --email proxypanel@localhost
        cd - > /dev/null
        rm -rf /tmp/acme.sh-repo
    fi

    # 设置默认CA为 ZeroSSL（推荐）
    ~/.acme.sh/acme.sh --set-default-ca --server zerossl

    # 注册 ZeroSSL 账号
    log_info "注册 ZeroSSL 账号..."
    ~/.acme.sh/acme.sh --register-account -m 444722146@qq.com --server zerossl

    log_success "acme.sh 安装完成（默认CA: ZeroSSL）"
}

# 安装Docker（可选）
install_docker() {
    log_info "检查Docker..."

    if command -v docker &> /dev/null; then
        log_warning "Docker已安装"
        return
    fi

    log_info "安装Docker（可选）..."

    # 添加Docker官方GPG密钥
    curl -fsSL https://download.docker.com/linux/${OS}/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

    # 添加Docker仓库
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/${OS} $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

    # 安装Docker
    apt-get update -y
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

    # 启动Docker
    systemctl enable docker
    systemctl start docker

    log_success "Docker安装完成"
}

# 创建项目目录
create_project_structure() {
    log_info "创建项目目录结构..."

    mkdir -p ${PROJECT_DIR}
    mkdir -p ${BACKEND_DIR}/data
    mkdir -p ${BACKEND_DIR}/templates
    mkdir -p ${BACKEND_DIR}/frontend/dist
    mkdir -p ${FRONTEND_DIR}
    mkdir -p ${PROJECT_DIR}/nginx/sites-enabled
    mkdir -p ${PROJECT_DIR}/nginx/logs
    mkdir -p ${PROJECT_DIR}/nginx/acme-challenge
    mkdir -p ${PROJECT_DIR}/ssl
    mkdir -p ${PROJECT_DIR}/scripts

    log_success "项目目录创建完成: ${PROJECT_DIR}"
}

# 配置Nginx
configure_nginx() {
    log_info "配置Nginx..."

    # 备份原有配置
    if [ -f /etc/nginx/nginx.conf ]; then
        cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup
    fi

    # 创建sites-enabled目录
    mkdir -p /etc/nginx/sites-enabled

    # 创建默认SSL证书
    mkdir -p /etc/nginx/ssl
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout /etc/nginx/ssl/default.key \
        -out /etc/nginx/ssl/default.crt \
        -subj "/CN=localhost" 2>/dev/null || true

    # 确保 ACME 验证目录存在且可访问
    mkdir -p /var/www/acme-challenge
    chmod -R 755 /var/www/acme-challenge

    # 在 Nginx 默认配置中添加 ACME 验证 location（如果尚未添加）
    DEFAULT_CONF="/etc/nginx/sites-enabled/default"
    if [ -f "$DEFAULT_CONF" ] && ! grep -q "acme-challenge" "$DEFAULT_CONF"; then
        # 在 server 块末尾的 } 之前插入 ACME location
        sed -i '/^}$/i \
    # ACME HTTP-01 验证目录（用于申请免费SSL证书）\
    location ^~ /.well-known/acme-challenge/ {\
        default_type "text/plain";\
        root /var/www/acme-challenge;\
    }' "$DEFAULT_CONF" 2>/dev/null || true
    fi

    # 重启Nginx
    nginx -t && systemctl restart nginx || true

    log_success "Nginx配置完成"
}

# 编译后端
build_backend() {
    log_info "编译ProxyPanel后端..."

    cd ${BACKEND_DIR}

    # 设置 Go 模块代理（国内用七牛 goproxy.cn，海外直连）
    if is_china; then
        export GOPROXY=https://goproxy.cn,direct
        go env -w GOPROXY=https://goproxy.cn,direct
        log_info "使用 goproxy.cn 加速 Go 模块下载..."
    else
        export GOPROXY=https://proxy.golang.org,direct
    fi

    # 整理依赖（修复 go.sum 缺失问题）
    go mod tidy

    # 下载Go依赖
    go mod download

    # 编译（需要 GCC，CGO_ENABLED=1 用于 SQLite 驱动）
    CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o proxypanel .

    log_success "后端编译完成"
}

# 编译前端
build_frontend() {
    log_info "编译ProxyPanel前端..."

    cd ${FRONTEND_DIR}

    # 安装依赖（国内使用淘宝 npm 镜像）
    if is_china; then
        npm config set registry https://registry.npmmirror.com
        log_info "使用 npmmirror.com 加速 npm 下载..."
    fi
    npm install

    # 编译
    npm run build

    # 复制编译产物到后端静态文件目录
    rm -rf ${BACKEND_DIR}/frontend/dist
    cp -r dist ${BACKEND_DIR}/frontend/dist

    log_success "前端编译完成"
}

# 创建系统服务
create_systemd_service() {
    log_info "创建系统服务..."

    SERVICE_FILE="/etc/systemd/system/proxypanel.service"

    cat > ${SERVICE_FILE} << EOF
[Unit]
Description=ProxyPanel API代理管理系统
After=network.target nginx.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/proxypanel/backend
ExecStart=/opt/proxypanel/backend/proxypanel
Restart=on-failure
RestartSec=5s
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=proxypanel
Environment=GIN_MODE=release
Environment=SERVER_PORT=5000

[Install]
WantedBy=multi-user.target
EOF

    # 重载systemd
    systemctl daemon-reload

    # 启用服务
    systemctl enable proxypanel

    log_success "系统服务创建完成"
}

# 启动服务
start_service() {
    log_info "启动ProxyPanel服务..."

    systemctl start proxypanel || true

    sleep 3

    # 检查服务状态
    if systemctl is-active --quiet proxypanel; then
        log_success "ProxyPanel服务启动成功"
    else
        log_error "ProxyPanel服务启动失败，请检查日志: journalctl -u proxypanel -n 50"
    fi
}

# 显示安装结果
show_result() {
    # 获取服务器IP
    SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || echo "YOUR_SERVER_IP")

    echo ""
    echo "========================================"
    log_success "ProxyPanel安装完成！"
    echo "========================================"
    echo ""
    echo "访问地址:"
    echo "  - Web管理面板: http://${SERVER_IP}:5000"
    echo ""
    echo "服务管理命令:"
    echo "  - 启动服务: systemctl start proxypanel"
    echo "  - 停止服务: systemctl stop proxypanel"
    echo "  - 重启服务: systemctl restart proxypanel"
    echo "  - 查看状态: systemctl status proxypanel"
    echo "  - 查看日志: journalctl -u proxypanel -f"
    echo ""
    echo "Nginx管理命令:"
    echo "  - 测试配置: nginx -t"
    echo "  - 重载配置: nginx -s reload"
    echo "  - 查看日志: tail -f /var/log/nginx/access.log"
    echo ""
    echo "SSL证书:"
    echo "  - 免费证书工具: ~/.acme.sh/acme.sh"
    echo "  - 默认CA: Let's Encrypt（也支持 ZeroSSL）"
    echo "  - 自动续签: 已通过 acme.sh cron 配置"
    echo ""
    echo "项目目录: /opt/proxypanel"
    echo "数据目录: /opt/proxypanel/backend/data"
    echo "SSL目录:  /opt/proxypanel/ssl"
    echo "ACME验证: /var/www/acme-challenge"
    echo ""
}

# 主安装流程
main() {
    echo ""
    echo "========================================"
    echo " ProxyPanel 一键安装脚本"
    echo "========================================"
    echo ""

    # 执行安装步骤
    check_root
    detect_system
    install_dependencies
    install_go
    install_nodejs
    install_acme_sh
    create_project_structure
    configure_nginx

    echo ""
    log_warning "请先将项目源代码放置到 /opt/proxypanel 目录"
    log_warning "然后继续编译和部署步骤"
    echo ""

    read -p "是否继续编译安装? (y/n): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        build_backend
        build_frontend
        create_systemd_service
        start_service
        show_result
    else
        log_warning "安装暂停，请手动编译部署"
    fi
}

# 执行主函数
main
