# AGENTS.md - ProxyPanel项目规范

## 项目概览

ProxyPanel是一个基于Go + Vue的API代理管理系统，提供可视化的Web管理界面，用于管理Nginx代理规则。

### 技术栈
- **后端**: Go 1.21 + Gin + SQLite + GORM
- **前端**: Vue 3 + Element Plus + Vite + Vue Router
- **代理服务器**: Nginx
- **部署方式**: Docker + Systemd

### 核心功能
- 代理规则管理（增删改查）
- 域名绑定与目标API配置
- Token认证与IP白名单
- Nginx配置自动生成与重载
- SSL证书管理（上传/自签名）
- 实时日志查看与搜索

## 目录结构

```
/workspace/projects/ProxyPanel/
├── backend/                # Go后端源码
│   ├── main.go            # 程序入口
│   ├── go.mod/go.sum      # Go依赖管理
│   ├── config/config.go   # 配置管理
│   ├── controllers/       # 业务控制器
│   │   ├── proxy.go       # 代理规则API
│   │   ├── log.go         # 日志API
│   │   └── ssl.go         # SSL证书API
│   ├── models/proxy.go    # 数据模型与CRUD
│   ├── routes/routes.go   # Gin路由配置
│   ├── utils/             # 工具函数
│   │   ├── database.go    # SQLite初始化
│   │   ├── response.go    # 统一响应格式
│   │   └── nginx.go       # Nginx配置生成
│   └── templates/         # Nginx配置模板
│       └── nginx.conf.tmpl
│
├── frontend/              # Vue前端源码
│   ├── package.json      # NPM依赖
│   ├── vite.config.js    # Vite配置
│   ├── index.html        # 入口HTML
│   ├── src/
│   │   ├── main.js       # Vue应用入口
│   │   ├── App.vue       # 根组件（含侧边栏）
│   │   ├── api/proxy.js  # Axios API封装
│   │   ├── router/index.js # Vue Router
│   │   └── views/        # 页面组件
│   │       ├── Dashboard.vue  # 仪表盘
│   │       ├── ProxyList.vue  # 代理管理
│   │       ├── Logs.vue       # 日志查看
│   │       └── SSL.vue        # SSL管理
│
├── nginx/                 # Nginx配置文件
│   ├── nginx.conf        # Nginx主配置
│   └── frontend.conf     # 前端静态服务配置
│
├── scripts/start.sh      # Docker启动脚本
├── docker-compose.yml    # Docker编排配置
├── Dockerfile            # 后端镜像构建
├── Dockerfile.frontend   # 前端镜像构建
├── install.sh            # Linux一键安装脚本
├── README.md             # 项目文档
└── AGENTS.md             # 本文件
```

## 构建与部署命令

### 后端开发
```bash
cd /workspace/projects/ProxyPanel/backend
go mod download           # 下载依赖
go run main.go            # 开发运行
CGO_ENABLED=1 go build -o proxypanel .  # 编译生产版本
```

### 前端开发
```bash
cd /workspace/projects/ProxyPanel/frontend
npm install               # 安装依赖
npm run dev               # 开发模式（端口3000）
npm run build             # 编译生产版本
```

### Docker部署
```bash
cd /workspace/projects/ProxyPanel
docker-compose up -d      # 启动所有服务
docker-compose logs -f    # 查看日志
docker-compose down       # 停止服务
```

### Systemd部署（Linux）
```bash
sudo ./install.sh        # 一键安装
systemctl start proxypanel     # 启动服务
systemctl status proxypanel    # 查看状态
journalctl -u proxypanel -f    # 查看日志
```

## 数据库设计

### 表名: proxy_rules
SQLite数据库，使用GORM自动迁移

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，自增 |
| name | VARCHAR(255) | 代理名称，必填 |
| domain | VARCHAR(255) | 域名，必填，唯一 |
| target_url | VARCHAR(500) | 目标地址，必填 |
| fake_ip | VARCHAR(50) | 伪装IP，默认127.0.0.1 |
| token | VARCHAR(255) | API认证Token |
| enabled | BOOLEAN | 启用状态，默认true |
| whitelist | TEXT | IP白名单（逗号分隔） |
| ssl_enabled | BOOLEAN | SSL启用状态 |
| ssl_cert | TEXT | SSL证书路径 |
| ssl_key | TEXT | SSL私钥路径 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

## API接口清单

### 代理规则 (/api/proxy)
- `GET /api/proxy` - 获取所有代理规则
- `GET /api/proxy/:id` - 获取单个代理规则
- `POST /api/proxy` - 创建代理规则（自动生成Nginx配置）
- `PUT /api/proxy/:id` - 更新代理规则（重新生成Nginx配置）
- `DELETE /api/proxy/:id` - 删除代理规则（删除Nginx配置）
- `POST /api/proxy/:id/toggle` - 切换启用状态

### Nginx管理 (/api/nginx)
- `POST /api/nginx/sync` - 同步所有配置到Nginx并重载
- `POST /api/nginx/test` - 测试Nginx配置有效性

### 日志管理 (/api/log)
- `GET /api/log/access` - 获取指定域名访问日志
- `GET /api/log/error` - 获取指定域名错误日志
- `GET /api/log/access/general` - 获取通用访问日志
- `GET /api/log/error/general` - 获取通用错误日志
- `DELETE /api/log/clear` - 清空指定日志
- `GET /api/log/search` - 搜索日志（关键词过滤）

### SSL证书 (/api/ssl)
- `POST /api/ssl/:id/upload` - 上传SSL证书文件（multipart/form-data）
- `POST /api/ssl/:id/generate` - 生成自签名证书（仅用于测试）
- `DELETE /api/ssl/:id` - 移除SSL证书

## 代码风格指南

### Go代码规范
- 使用Gin框架，遵循RESTful API设计
- 统一响应格式：`{code: 200, message: "success", data: {...}}`
- 所有控制器函数需处理错误并返回统一响应
- 使用GORM进行数据库操作，自动迁移表结构
- 配置通过环境变量读取，支持自定义默认值
- Nginx配置使用Go模板引擎生成

### Vue代码规范
- 使用Composition API (script setup)
- Element Plus组件按需引入
- 使用Axios封装API，统一处理响应拦截
- Vue Router配置路由懒加载
- 所有页面组件包含加载状态和错误处理
- 表单验证使用Element Plus内置规则

### 文件命名规范
- Go文件：小写字母，下划线分隔（如：proxy_rule.go）
- Vue文件：驼峰命名（如：ProxyList.vue）
- 配置文件：小写字母，点分隔（如：nginx.conf）

## 常见问题与修复

### Nginx配置生成失败
**问题**: 生成Nginx配置时模板解析失败
**定位**: `backend/utils/nginx.go` -> `GenerateNginxConfig()`
**修复**: 检查模板文件路径和模板语法，确保变量正确传递

### SQLite数据库锁死
**问题**: 高并发下SQLite写入失败
**定位**: `backend/utils/database.go` -> `InitDB()`
**修复**: 使用GORM的连接池配置，限制并发写入

### 前端API调用跨域
**问题**: 前端调用后端API出现CORS错误
**定位**: `backend/routes/routes.go` -> CORS配置
**修复**: Gin配置允许所有源或指定域名

### SSL证书验证失败
**问题**: 上传的证书与私钥不匹配
**定位**: `backend/controllers/ssl.go` -> `validateCertificate()`
**修复**: 使用Go crypto包验证证书和私钥匹配性

## 安全注意事项

1. **数据库文件权限**: SQLite数据库文件需设置适当权限（建议600）
2. **Token存储**: Token存储在数据库中，建议加密存储
3. **SSL证书路径**: 证书文件路径需验证合法性，防止路径遍历
4. **Nginx配置注入**: 模板生成时需清理特殊字符，防止配置注入
5. **IP白名单验证**: 需验证IP格式合法性，防止无效配置

## 性能优化建议

1. **数据库查询**: 使用索引优化查询（domain字段已建立唯一索引）
2. **Nginx配置缓存**: 高频读取的配置可考虑缓存机制
3. **日志读取**: 大日志文件使用分页读取，避免内存溢出
4. **前端懒加载**: 使用Vue Router懒加载，减少首屏加载时间
5. **静态资源**: 前端静态资源使用CDN或Nginx缓存策略

## 测试说明

### 后端测试
```bash
# 启动后端服务
cd backend && go run main.go

# 测试API接口（使用curl）
curl http://localhost:5000/api/proxy
curl -X POST -H "Content-Type: application/json" -d '{"name":"test","domain":"test.com","target_url":"http://127.0.0.1"}' http://localhost:5000/api/proxy
```

### 前端测试
```bash
# 启动前端开发服务器
cd frontend && npm run dev

# 访问 http://localhost:3000 进行UI测试
```

### 集成测试
```bash
# Docker完整测试
docker-compose up -d
# 访问 http://localhost:5000 进行完整功能测试
```

---

**更新日期**: 2024-01
**维护者**: ProxyPanel Team