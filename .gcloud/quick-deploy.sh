#!/bin/bash

# Quick deployment script for Google Cloud Run
# This is a simplified version - follow DEPLOYMENT_GUIDE.md for detailed steps

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
PROJECT_ID="acoustic-shade-480602-u0"
REGION="asia-east1"
SERVICE_NAME="new-api"
IMAGE_NAME="gcr.io/${PROJECT_ID}/${SERVICE_NAME}"

echo -e "${BLUE}🚀 New-API Quick Deployment to Google Cloud${NC}\n"

# Step 1: Pre-flight checks
echo -e "${YELLOW}Step 1: Checking prerequisites...${NC}"

if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    exit 1
fi

if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}❌ gcloud CLI is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Prerequisites OK${NC}\n"

# Step 2: Build Docker image
echo -e "${YELLOW}Step 2: Building Docker image...${NC}"
echo "   This may take 10-15 minutes..."

docker build \
  -t ${IMAGE_NAME}:latest \
  -t ${IMAGE_NAME}:$(date +%Y%m%d-%H%M%S) \
  --platform linux/amd64 \
  .

echo -e "${GREEN}✓ Docker image built${NC}\n"

# Step 3: Configure Docker authentication
echo -e "${YELLOW}Step 3: Configuring Docker authentication...${NC}"
gcloud auth configure-docker --quiet
echo -e "${GREEN}✓ Docker authentication configured${NC}\n"

# Step 4: Push image to Google Container Registry
echo -e "${YELLOW}Step 4: Pushing image to Google Container Registry...${NC}"
docker push ${IMAGE_NAME}:latest
echo -e "${GREEN}✓ Image pushed${NC}\n"

# Step 5: Enable required APIs
echo -e "${YELLOW}Step 5: Enabling required APIs...${NC}"
gcloud services enable \
  run.googleapis.com \
  containerregistry.googleapis.com \
  cloudsql.googleapis.com \
  redis.googleapis.com \
  --project=${PROJECT_ID} 2>/dev/null || true
echo -e "${GREEN}✓ APIs enabled${NC}\n"

# Step 6: Deploy to Cloud Run
echo -e "${YELLOW}Step 6: Deploying to Cloud Run...${NC}"

# For quick deployment without Cloud SQL setup, use SQLite
# Uncomment below if you have Cloud SQL and Redis set up:

# If you have Cloud SQL and Redis configured, use this:
# gcloud run deploy ${SERVICE_NAME} \
#   --image ${IMAGE_NAME}:latest \
#   --region ${REGION} \
#   --platform managed \
#   --allow-unauthenticated \
#   --memory 2Gi \
#   --cpu 2 \
#   --timeout 3600 \
#   --set-env-vars "SQL_DSN=your_sql_dsn,REDIS_CONN_STRING=your_redis_url,SESSION_SECRET=$(openssl rand -base64 32),CRYPTO_SECRET=$(openssl rand -base64 32),TZ=Asia/Shanghai" \
#   --project=${PROJECT_ID}

# For quick test with SQLite (not recommended for production):
gcloud run deploy ${SERVICE_NAME} \
  --image ${IMAGE_NAME}:latest \
  --region ${REGION} \
  --platform managed \
  --allow-unauthenticated \
  --memory 2Gi \
  --cpu 2 \
  --timeout 3600 \
  --set-env-vars "SESSION_SECRET=$(openssl rand -base64 32),CRYPTO_SECRET=$(openssl rand -base64 32),TZ=Asia/Shanghai" \
  --project=${PROJECT_ID}

echo -e "${GREEN}✓ Deployment completed${NC}\n"

# Step 7: Get service URL
SERVICE_URL=$(gcloud run services describe ${SERVICE_NAME} \
  --region ${REGION} \
  --project=${PROJECT_ID} \
  --format='value(status.url)')

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ Deployment Successful!${NC}"
echo -e "${BLUE}========================================${NC}\n"

echo -e "${YELLOW}📊 Service Information:${NC}"
echo -e "  ${BLUE}Service URL:${NC} ${GREEN}${SERVICE_URL}${NC}"
echo -e "  ${BLUE}Project ID:${NC} ${PROJECT_ID}"
echo -e "  ${BLUE}Service Name:${NC} ${SERVICE_NAME}"
echo -e "  ${BLUE}Region:${NC} ${REGION}\n"

echo -e "${YELLOW}📝 Next Steps:${NC}"
echo -e "  1. Open: ${GREEN}${SERVICE_URL}${NC}"
echo -e "  2. Login with: admin / 123456"
echo -e "  3. Change default password"
echo -e "  4. Configure API providers"
echo -e "  5. Set up payment methods\n"

echo -e "${YELLOW}💾 View logs:${NC}"
echo "  gcloud run services logs read ${SERVICE_NAME} --region=${REGION} --limit=50\n"

echo -e "${YELLOW}❌ To stop service:${NC}"
echo "  gcloud run services delete ${SERVICE_NAME} --region=${REGION} --quiet\n"
