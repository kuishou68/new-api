# 🚀 New-API Google Cloud Deployment Guide

## 前置条件

- ✅ gcloud CLI 已安装并配置
- ✅ Docker 已安装
- ✅ Git 已安装
- ✅ Google Cloud 项目已创建

## 部署信息

- **GCP 项目 ID**: `acoustic-shade-480602-u0`
- **GCP 账户**: `lijianlin6868@gmail.com`
- **部署服务**: Google Cloud Run
- **部署区域**: asia-east1 (台湾)
- **数据库**: Cloud SQL MySQL
- **缓存**: Cloud Memorystore (Redis)

---

## 📋 部署步骤

### 步骤 1：确保 Docker 已启动

```bash
# macOS: 启动 Docker Desktop
open /Applications/Docker.app

# 验证 Docker 是否运行
docker --version
docker ps
```

---

### 步骤 2：设置环境变量

```bash
export PROJECT_ID="acoustic-shade-480602-u0"
export REGION="asia-east1"
export SERVICE_NAME="new-api"
export IMAGE_NAME="gcr.io/${PROJECT_ID}/${SERVICE_NAME}:latest"
```

---

### 步骤 3：配置 Docker 向 Google Container Registry 推送

```bash
# 配置 Docker 认证
gcloud auth configure-docker

# 验证配置
cat ~/.docker/config.json | grep "gcr.io"
```

---

### 步骤 4：构建 Docker 镜像

```bash
# 进入项目目录
cd /Users/pojian/code/github/new-api

# 构建镜像（可能需要 10-15 分钟）
docker build \
  -t ${IMAGE_NAME} \
  -t gcr.io/${PROJECT_ID}/${SERVICE_NAME}:$(date +%Y%m%d-%H%M%S) \
  --platform linux/amd64 \
  .

# 验证镜像
docker images | grep new-api
```

---

### 步骤 5：推送镜像到 Google Container Registry

```bash
# 推送镜像
docker push ${IMAGE_NAME}

# 验证推送
gcloud container images list --project=${PROJECT_ID}
```

---

### 步骤 6：创建 Cloud SQL MySQL 实例

```bash
# 设置变量
export DB_INSTANCE="new-api-mysql"
export DB_NAME="new_api"
export DB_USER="root"
export DB_PASSWORD=$(openssl rand -base64 32)

# 创建实例
gcloud sql instances create ${DB_INSTANCE} \
  --database-version=MYSQL_8_0 \
  --tier=db-f1-micro \
  --region=${REGION} \
  --project=${PROJECT_ID} \
  --availability-type=ZONAL \
  --no-backup

# 等待实例就绪
gcloud sql operations wait --project=${PROJECT_ID}

# 创建数据库
gcloud sql databases create ${DB_NAME} \
  --instance=${DB_INSTANCE} \
  --project=${PROJECT_ID}

# 设置 root 用户密码
gcloud sql users set-password root \
  --instance=${DB_INSTANCE} \
  --password=${DB_PASSWORD} \
  --project=${PROJECT_ID}

# 获取 IP 地址
export DB_IP=$(gcloud sql instances describe ${DB_INSTANCE} \
  --project=${PROJECT_ID} \
  --format='value(ipAddresses[0].ipAddress)')

echo "Database IP: ${DB_IP}"
echo "Database Password: ${DB_PASSWORD}"
```

---

### 步骤 7：创建 Cloud Memorystore (Redis) 实例

```bash
# 设置变量
export REDIS_INSTANCE="new-api-redis"

# 创建 Redis 实例
gcloud redis instances create ${REDIS_INSTANCE} \
  --size=1 \
  --region=${REGION} \
  --project=${PROJECT_ID} \
  --redis-version=7.0

# 获取 Redis 连接信息
export REDIS_HOST=$(gcloud redis instances describe ${REDIS_INSTANCE} \
  --region=${REGION} \
  --project=${PROJECT_ID} \
  --format='value(host)')

export REDIS_PORT=$(gcloud redis instances describe ${REDIS_INSTANCE} \
  --region=${REGION} \
  --project=${PROJECT_ID} \
  --format='value(port)')

echo "Redis Host: ${REDIS_HOST}"
echo "Redis Port: ${REDIS_PORT}"
```

---

### 步骤 8：准备环境变量

```bash
# 生成密钥
export SESSION_SECRET=$(openssl rand -base64 32)
export CRYPTO_SECRET=$(openssl rand -base64 32)

# 构建数据库连接字符串
export SQL_DSN="${DB_USER}:${DB_PASSWORD}@tcp(${DB_IP}:3306)/${DB_NAME}?parseTime=true"
export REDIS_CONN_STRING="redis://${REDIS_HOST}:${REDIS_PORT}/0"

# 验证变量
cat << EOF
Project ID: ${PROJECT_ID}
Service Name: ${SERVICE_NAME}
Image: ${IMAGE_NAME}
Region: ${REGION}

Database:
  Instance: ${DB_INSTANCE}
  IP: ${DB_IP}
  Name: ${DB_NAME}
  User: ${DB_USER}
  DSN: ${SQL_DSN}

Redis:
  Instance: ${REDIS_INSTANCE}
  Host: ${REDIS_HOST}
  Port: ${REDIS_PORT}
  URL: ${REDIS_CONN_STRING}

Secrets:
  SESSION_SECRET: ${SESSION_SECRET:0:20}...
  CRYPTO_SECRET: ${CRYPTO_SECRET:0:20}...
EOF
```

---

### 步骤 9：部署到 Cloud Run

```bash
# 启用必要的 API
gcloud services enable run.googleapis.com \
  containerregistry.googleapis.com \
  --project=${PROJECT_ID}

# 部署到 Cloud Run
gcloud run deploy ${SERVICE_NAME} \
  --image ${IMAGE_NAME} \
  --region ${REGION} \
  --platform managed \
  --allow-unauthenticated \
  --memory 2Gi \
  --cpu 2 \
  --timeout 3600 \
  --set-env-vars \
    SQL_DSN="${SQL_DSN}",\
    REDIS_CONN_STRING="${REDIS_CONN_STRING}",\
    SESSION_SECRET="${SESSION_SECRET}",\
    CRYPTO_SECRET="${CRYPTO_SECRET}",\
    TZ="Asia/Shanghai" \
  --project=${PROJECT_ID}

# 获取服务 URL
export SERVICE_URL=$(gcloud run services describe ${SERVICE_NAME} \
  --region ${REGION} \
  --project=${PROJECT_ID} \
  --format='value(status.url)')

echo "Service URL: ${SERVICE_URL}"
```

---

### 步骤 10：保存部署信息

```bash
# 创建部署记录文件
cat > .gcloud/deployment-info.txt << EOF
========================================
New-API Google Cloud Deployment
========================================
部署时间: $(date)

GCP 项目: ${PROJECT_ID}
服务名称: ${SERVICE_NAME}
部署区域: ${REGION}
服务 URL: ${SERVICE_URL}

数据库信息:
  实例: ${DB_INSTANCE}
  IP: ${DB_IP}
  数据库: ${DB_NAME}
  用户: ${DB_USER}
  密码: ${DB_PASSWORD}
  DSN: ${SQL_DSN}

Redis 信息:
  实例: ${REDIS_INSTANCE}
  主机: ${REDIS_HOST}
  端口: ${REDIS_PORT}
  连接: ${REDIS_CONN_STRING}

安全信息:
  SESSION_SECRET: ${SESSION_SECRET}
  CRYPTO_SECRET: ${CRYPTO_SECRET}

默认账户:
  用户名: admin
  密码: 123456 (请立即修改!)

访问应用:
  1. 打开浏览器访问: ${SERVICE_URL}
  2. 使用 admin / 123456 登录
  3. 修改默认密码
  4. 配置 AI 提供商 API Keys
  5. 配置支付方式
========================================
EOF

# 显示文件路径
echo "部署信息已保存到: .gcloud/deployment-info.txt"
cat .gcloud/deployment-info.txt
```

---

## ✅ 验证部署

```bash
# 1. 检查 Cloud Run 服务
gcloud run services describe ${SERVICE_NAME} \
  --region=${REGION} \
  --project=${PROJECT_ID}

# 2. 检查服务日志
gcloud run services logs read ${SERVICE_NAME} \
  --region=${REGION} \
  --project=${PROJECT_ID} \
  --limit 50

# 3. 测试服务
curl -s ${SERVICE_URL}/api/status | jq .

# 4. 打开浏览器访问
open ${SERVICE_URL}
```

---

## 🔐 安全检查清单

- [ ] 修改默认管理员密码
- [ ] 配置防火墙规则（仅允许特定 IP）
- [ ] 启用 Cloud SQL Auth
- [ ] 设置数据库备份
- [ ] 配置 SSL 证书
- [ ] 启用审计日志
- [ ] 定期监控 Cloud Run 成本

---

## 💾 备份和恢复

```bash
# 备份数据库
gcloud sql backups create \
  --instance=${DB_INSTANCE} \
  --project=${PROJECT_ID}

# 列出备份
gcloud sql backups list \
  --instance=${DB_INSTANCE} \
  --project=${PROJECT_ID}

# 恢复备份
gcloud sql backups restore BACKUP_ID \
  --instance=${DB_INSTANCE} \
  --project=${PROJECT_ID}
```

---

## 🚨 故障排查

### 常见问题

1. **Docker 构建失败**
   ```bash
   docker build --no-cache -t ${IMAGE_NAME} .
   ```

2. **推送镜像失败**
   ```bash
   gcloud auth configure-docker
   docker push ${IMAGE_NAME}
   ```

3. **Cloud Run 部署失败**
   ```bash
   gcloud run services logs read ${SERVICE_NAME} \
     --region=${REGION} \
     --limit 100
   ```

4. **数据库连接失败**
   - 检查 Cloud SQL 实例是否运行
   - 验证网络连接
   - 检查防火墙规则

---

## 📞 支持资源

- [Google Cloud Run 文档](https://cloud.google.com/run/docs)
- [Cloud SQL 文档](https://cloud.google.com/sql/docs)
- [Cloud Memorystore 文档](https://cloud.google.com/memorystore/docs)
- [New-API GitHub](https://github.com/QuantumNous/new-api)

---

## 成本估算 (月度)

| 服务 | 估算成本 |
|-----|--------|
| Cloud Run | $0.20 - $5 |
| Cloud SQL (db-f1-micro) | $3.65 |
| Cloud Memorystore (1GB) | $0.70 |
| 存储 | $0.50 |
| **总计** | **$5 - $10/月** |

> 注：具体成本取决于流量和使用情况，可使用 [Google Cloud Pricing Calculator](https://cloud.google.com/products/calculator) 计算
