# ProxyPanel - API代理管理系统

## 项目简介

ProxyPanel是一个可视化的API代理管理平台，基于Go语言开发，提供完整的Web管理界面。用户无需登录SSH，通过网页即可完成所有代理配置操作。

## 功能特性

### 核心功能
- ✅ **代理规则管理**：新增、修改、删除代理规则
- ✅ **域名绑定**：为每个代理配置独立域名
- ✅ **目标API设置**：配置真实的目标API地址
- ✅ **Token认证**：支持API密钥认证
- ✅ **IP白名单**：配置访问白名单限制
- ✅ **Nginx自动生成**：自动生成并部署Nginx配置
- ✅ **Nginx自动重载**：配置变更自动重载Nginx
- ✅ **日志查看**：实时查看访问日志和错误日志
- ✅ **SSL证书管理**：上传证书或生成自签名证书

### 安全特性
- 🔒 **隐藏真实IP**：所有流量统一从代理服务器出口，隐藏业务服务器真实IP
- 🔒 **访问控制**：IP白名单限制，只允许指定IP访问
- 🔒 **Token认证**：API密钥验证，防止未授权访问
- 🔒 **SSL加密**：支持HTTPS，保障数据传输安全

## 技术栈

- **后端**: Go 1.21 + Gin + SQLite
- **前端**: Vue 3 + Element Plus + Vite
- **代理**: Nginx
- **部署**: Docker / Systemd

## 快速部署

### 方式一：Docker部署（推荐）

```bash
# 克隆项目
git clone https://github.com/your-repo/ProxyPanel.git
cd ProxyPanel

# 启动服务
docker-compose up -d

# 访问管理面板
http://localhost:5000
```

### 方式二：Linux一键安装

```bash
# 下载安装脚本
wget https://your-domain/install.sh

# 执行安装（需要root权限）
chmod +x install.sh
sudo ./install.sh
```

### 方式三：手动编译部署

```bash
# 编译后端
cd backend
go mod download
CGO_ENABLED=1 go build -o proxypanel .

# 编译前端
cd frontend
npm install
npm run build

# 启动服务
./backend/proxypanel
```

## 目录结构

```
ProxyPanel/
├── backend/              # 后端Go源码
│   ├── main.go          # 程序入口
│   ├── config/          # 配置管理
│   ├── controllers/     # 业务控制器
│   ├── models/          # 数据模型
│   ├── routes/          # 路由配置
│   ├── utils/           # 工具函数
│   └── templates/       # Nginx模板
│
├── frontend/            # 前端Vue源码
│   ├── src/
│   │   ├── views/      # 页面组件
│   │   ├── api/        # API接口
│   │   └── router/     # 路由配置
│   └── dist/           # 编译产物
│
├── nginx/               # Nginx配置
│   ├── nginx.conf      # 主配置
│   └── sites-enabled/  # 代理配置
│
├── docker-compose.yml   # Docker编排
├── Dockerfile           # 后端镜像
├── install.sh           # 一键安装脚本
└── README.md            # 项目文档
```

## 数据库结构

### proxy_rules表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| name | VARCHAR(255) | 代理名称 |
| domain | VARCHAR(255) | 域名 |
| target_url | VARCHAR(500) | 目标地址 |
| fake_ip | VARCHAR(50) | 伪装IP |
| token | VARCHAR(255) | 认证Token |
| enabled | BOOLEAN | 启用状态 |
| whitelist | TEXT | IP白名单 |
| ssl_enabled | BOOLEAN | SSL启用 |
| ssl_cert | TEXT | 证书路径 |
| ssl_key | TEXT | 私钥路径 |
| created_at | DATETIME | 创建时间 |

## API接口文档

### 代理规则API

```
GET    /api/proxy          # 获取所有代理规则
GET    /api/proxy/:id      # 获取单个代理规则
POST   /api/proxy          # 创建代理规则
PUT    /api/proxy/:id      # 更新代理规则
DELETE /api/proxy/:id      # 删除代理规则
POST   /api/proxy/:id/toggle # 切换状态
```

### Nginx管理API

```
POST   /api/nginx/sync     # 同步配置到Nginx
POST   /api/nginx/test     # 测试Nginx配置
```

### 日志API

```
GET    /api/log/access     # 获取访问日志
GET    /api/log/error      # 获取错误日志
DELETE /api/log/clear      # 清空日志
GET    /api/log/search     # 搜索日志
```

### SSL证书API

```
POST   /api/ssl/:id/upload     # 上传证书
POST   /api/ssl/:id/generate   # 生成自签名证书
DELETE /api/ssl/:id            # 移除证书
```

## 使用示例

### 创建代理规则

```bash
curl -X POST http://localhost:5000/api/proxy \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Gateway",
    "domain": "api.example.com",
    "target_url": "http://192.168.1.100:8080",
    "fake_ip": "127.0.0.1",
    "token": "your-secret-token",
    "enabled": true
  }'
```

### Nginx配置模板

生成的Nginx配置示例：

```nginx
server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://192.168.1.100:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP 127.0.0.1;
        proxy_set_header X-Forwarded-For 127.0.0.1;
        proxy_set_header X-API-KEY "your-secret-token";
    }
}
```

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| SERVER_PORT | 服务端口 | 5000 |
| GIN_MODE | 运行模式 | release |
| TZ | 时区 | Asia/Shanghai |

### Nginx配置

- 主配置：`/etc/nginx/nginx.conf`
- 代理配置：`/etc/nginx/sites-enabled/proxy_*.conf`
- 日志目录：`/var/log/nginx/`

## 安全建议

1. **生产环境务必配置SSL证书**
2. **设置IP白名单限制访问**
3. **定期更新Token认证密钥**
4. **配置防火墙规则**
5. **定期检查访问日志**

## 故障排查

### 查看服务状态

```bash
systemctl status proxypanel
journalctl -u proxypanel -f
```

### 查看Nginx日志

```bash
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log
nginx -t
```

### 检查端口占用

```bash
netstat -tunlp | grep :5000
netstat -tunlp | grep :80
```

## 开发指南

### 编译开发

```bash
# 后端开发
cd backend
go run main.go

# 前端开发
cd frontend
npm run dev
```

### 代码规范

- Go代码遵循官方规范
- Vue代码使用Composition API
- 所有接口必须添加注释
- 错误处理必须完整

## 许可证

MIT License

## 联系方式

- 项目主页：https://github.com/your-repo/ProxyPanel
- 问题反馈：https://github.com/your-repo/ProxyPanel/issues

---

**ProxyPanel - 让API代理管理变得简单！**