# 🚀 New-API Google Cloud Deployment

Welcome! This directory contains all the resources needed to deploy new-api to Google Cloud.

## 📁 Files

| File | Purpose |
|------|---------|
| `DEPLOYMENT_GUIDE.md` | 📖 **Detailed step-by-step deployment guide** - READ THIS FIRST |
| `quick-deploy.sh` | ⚡ Quick deployment script (simplified) |
| `cloudbuild.yaml` | 🔨 Cloud Build configuration (CI/CD) |
| `deployment-info.txt` | 💾 Deployment credentials (after deployment) |

## 🎯 Quick Start (3 choices)

### Option 1️⃣: Quick Deploy (Recommended for first time)

```bash
# 1. Make script executable
chmod +x quick-deploy.sh

# 2. Start Docker Desktop (important!)
open /Applications/Docker.app

# 3. Run deployment
./quick-deploy.sh

# This will:
#   - Build Docker image (~10-15 min)
#   - Push to Google Container Registry
#   - Deploy to Cloud Run
#   - Show you the service URL
```

### Option 2️⃣: Detailed Manual Deployment (Production recommended)

Follow the comprehensive guide:

```bash
cat DEPLOYMENT_GUIDE.md
```

This covers:
- Setting up Cloud SQL MySQL
- Setting up Cloud Memorystore Redis
- Database configuration
- Security best practices

### Option 3️⃣: Cloud Build CI/CD

Deploy via Google Cloud Build (requires git push):

```bash
gcloud builds submit --config=cloudbuild.yaml
```

---

## 📋 Prerequisites

- ✅ [Docker Desktop](https://www.docker.com/products/docker-desktop) installed
- ✅ [gcloud CLI](https://cloud.google.com/sdk/docs/install) installed
- ✅ Google Cloud project created
- ✅ gcloud configured: `gcloud config set project acoustic-shade-480602-u0`

---

## 🚀 Deployment Process Overview

```
┌─────────────────────────────────────┐
│  1. Build Docker Image              │  (~15 min)
│     docker build -t ...             │
└─────────────┬───────────────────────┘
              ↓
┌─────────────────────────────────────┐
│  2. Push to Container Registry      │  (~2 min)
│     docker push gcr.io/...          │
└─────────────┬───────────────────────┘
              ↓
┌─────────────────────────────────────┐
│  3. Deploy to Cloud Run             │  (~3 min)
│     gcloud run deploy ...           │
└─────────────┬───────────────────────┘
              ↓
┌─────────────────────────────────────┐
│  ✅ Service Live!                   │
│     https://new-api-xxxx.run.app    │
└─────────────────────────────────────┘
```

**Total Time**: ~20-25 minutes ⏱️

---

## 🔍 After Deployment

### Access Your Application

```bash
# Get service URL
gcloud run services describe new-api \
  --region=asia-east1 \
  --format='value(status.url)'

# Open in browser
open https://your-service-url.run.app
```

### Default Credentials

```
Username: admin
Password: 123456
```

⚠️ **IMPORTANT: Change these immediately after login!**

### First Steps

1. ✅ Login to admin dashboard
2. ✅ Change default password
3. ✅ Add your AI API providers (Claude, OpenAI, etc.)
4. ✅ Configure payment methods (Stripe, Epay, etc.)
5. ✅ Invite users

---

## 📊 Monitor Your Deployment

```bash
# View recent logs
gcloud run services logs read new-api --region=asia-east1 --limit=50

# Check service status
gcloud run services describe new-api --region=asia-east1

# View metrics
gcloud run services describe new-api \
  --region=asia-east1 \
  --format='table(status.latestRevision, status.traffic[0].percent)'

# Stream live logs
gcloud run services logs read new-api \
  --region=asia-east1 \
  --follow
```

---

## 💰 Cost Estimation

**Monthly costs (approximate)**:

| Service | Cost |
|---------|------|
| Cloud Run (compute) | $0.20 - $5 |
| Cloud SQL (db-f1-micro) | $3.65 |
| Cloud Memorystore (1GB) | $0.70 |
| Storage | $0.50 |
| **Total** | **$5 - $10/month** |

> Use [Google Cloud Pricing Calculator](https://cloud.google.com/products/calculator) for accurate estimates

---

## 🔐 Security Checklist

- [ ] Change default admin password
- [ ] Enable HTTPS (automatic with Cloud Run)
- [ ] Set up firewall rules
- [ ] Enable Cloud SQL Auth
- [ ] Configure backups
- [ ] Enable audit logs
- [ ] Use environment variables for secrets

---

## 🆘 Troubleshooting

### Docker Build Fails

```bash
# Clear Docker cache and rebuild
docker build --no-cache -t gcr.io/acoustic-shade-480602-u0/new-api .
```

### Cannot Push to Container Registry

```bash
# Reconfigure Docker authentication
gcloud auth configure-docker
```

### Cloud Run Deployment Fails

```bash
# Check logs for detailed error
gcloud run services logs read new-api --region=asia-east1 --limit=100
```

### Service Returns 500 Error

```bash
# Check application logs
gcloud run services logs read new-api --region=asia-east1 --limit=200 --follow
```

---

## 📚 Documentation

- [Complete Deployment Guide](DEPLOYMENT_GUIDE.md)
- [Google Cloud Run Docs](https://cloud.google.com/run/docs)
- [Cloud SQL Docs](https://cloud.google.com/sql/docs)
- [new-api GitHub](https://github.com/QuantumNous/new-api)

---

## 🎯 Common Tasks

### Update Your Application

```bash
# 1. Make changes locally
# 2. Commit and push
# 3. Rebuild and push
docker build -t gcr.io/acoustic-shade-480602-u0/new-api:latest .
docker push gcr.io/acoustic-shade-480602-u0/new-api:latest

# 4. Deploy new version
gcloud run deploy new-api \
  --image gcr.io/acoustic-shade-480602-u0/new-api:latest \
  --region=asia-east1
```

### Scale Your Service

```bash
# Update memory and CPU
gcloud run services update new-api \
  --memory=4Gi \
  --cpu=4 \
  --region=asia-east1
```

### View Usage Statistics

```bash
# Cloud Run metrics
gcloud run services describe new-api \
  --region=asia-east1 \
  --format=json | jq '.status.traffic'
```

---

## 🚨 Rollback to Previous Version

```bash
# List all revisions
gcloud run revisions list --service=new-api --region=asia-east1

# Revert to previous revision
gcloud run services update-traffic new-api \
  --to-revisions REVISION_NAME=100 \
  --region=asia-east1
```

---

## 💡 Pro Tips

1. **Use Cloud Build for CI/CD**: Automatically rebuild and deploy on git push
2. **Set memory limits**: Start with 2GB, scale based on actual usage
3. **Enable caching**: Use Redis for better performance
4. **Monitor costs**: Set up billing alerts in Google Cloud Console
5. **Backup regularly**: Enable Cloud SQL backups

---

## 📞 Support

- 💬 **GitHub Issues**: https://github.com/QuantumNous/new-api/issues
- 📖 **Official Docs**: https://docs.newapi.pro
- 💡 **Discussions**: https://github.com/QuantumNous/new-api/discussions

---

**Good luck with your deployment! 🎉**
