#!/bin/bash

# Deployment script for Google Cloud Run
# This script deploys new-api to Google Cloud Run with Cloud SQL MySQL

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="acoustic-shade-480602-u0"
REGION="asia-east1"
SERVICE_NAME="new-api"
INSTANCE_NAME="new-api-mysql"
DB_NAME="new_api"
DB_USER="root"
DB_PASS=$(openssl rand -base64 32)
REDIS_INSTANCE="new-api-redis"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}New-API Google Cloud Deployment Script${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}❌ gcloud CLI is not installed. Please install it first.${NC}"
    exit 1
fi

# Check if we're logged in
echo -e "${YELLOW}📍 Checking GCP credentials...${NC}"
gcloud auth list
echo ""

# Step 1: Enable required APIs
echo -e "${YELLOW}1️⃣  Enabling required APIs...${NC}"
gcloud services enable \
    run.googleapis.com \
    containerregistry.googleapis.com \
    cloudsql.googleapis.com \
    sqladmin.googleapis.com \
    redis.googleapis.com \
    --project=$PROJECT_ID 2>/dev/null || echo -e "${YELLOW}⚠️  Some APIs might already be enabled${NC}"

echo -e "${GREEN}✓ APIs enabled${NC}\n"

# Step 2: Create Cloud SQL MySQL Instance
echo -e "${YELLOW}2️⃣  Creating Cloud SQL MySQL instance...${NC}"
echo "   Instance: $INSTANCE_NAME"
echo "   Region: $REGION"
echo "   DB Name: $DB_NAME"
echo "   Root Password: (saved separately)"

gcloud sql instances create $INSTANCE_NAME \
    --database-version=MYSQL_8_0 \
    --tier=db-f1-micro \
    --region=$REGION \
    --project=$PROJECT_ID \
    --availability-type=ZONAL \
    2>/dev/null || echo -e "${YELLOW}⚠️  Instance might already exist${NC}"

# Wait for instance to be ready
echo -e "${YELLOW}   Waiting for instance to be ready...${NC}"
sleep 30

# Create database
gcloud sql databases create $DB_NAME \
    --instance=$INSTANCE_NAME \
    --project=$PROJECT_ID \
    2>/dev/null || echo -e "${YELLOW}⚠️  Database might already exist${NC}"

# Set root password
gcloud sql users set-password root \
    --instance=$INSTANCE_NAME \
    --password=$DB_PASS \
    --project=$PROJECT_ID \
    2>/dev/null || echo -e "${YELLOW}⚠️  Password might already be set${NC}"

echo -e "${GREEN}✓ Cloud SQL MySQL instance created${NC}\n"

# Step 3: Create Redis instance
echo -e "${YELLOW}3️⃣  Creating Cloud Memorystore (Redis) instance...${NC}"
echo "   Instance: $REDIS_INSTANCE"
echo "   Region: $REGION"

gcloud redis instances create $REDIS_INSTANCE \
    --size=1 \
    --region=$REGION \
    --project=$PROJECT_ID \
    --redis-version=7.0 \
    2>/dev/null || echo -e "${YELLOW}⚠️  Redis instance might already exist${NC}"

echo -e "${GREEN}✓ Redis instance created${NC}\n"

# Step 4: Get Cloud SQL connection string
echo -e "${YELLOW}4️⃣  Retrieving connection information...${NC}"

# Get Cloud SQL public IP (or private IP if available)
SQL_IP=$(gcloud sql instances describe $INSTANCE_NAME \
    --project=$PROJECT_ID \
    --format='value(ipAddresses[0].ipAddress)' 2>/dev/null)

# Get Cloud SQL connection name for Cloud SQL Proxy
SQL_CONNECTION_NAME=$(gcloud sql instances describe $INSTANCE_NAME \
    --project=$PROJECT_ID \
    --format='value(connectionName)' 2>/dev/null)

# Get Redis host and port
REDIS_HOST=$(gcloud redis instances describe $REDIS_INSTANCE \
    --region=$REGION \
    --project=$PROJECT_ID \
    --format='value(host)' 2>/dev/null)

REDIS_PORT=$(gcloud redis instances describe $REDIS_INSTANCE \
    --region=$REGION \
    --project=$PROJECT_ID \
    --format='value(port)' 2>/dev/null)

# Build connection strings
SQL_DSN="${DB_USER}:${DB_PASS}@tcp(${SQL_IP}:3306)/${DB_NAME}?parseTime=true"
REDIS_CONN_STRING="redis://${REDIS_HOST}:${REDIS_PORT}/0"
SESSION_SECRET=$(openssl rand -base64 32)
CRYPTO_SECRET=$(openssl rand -base64 32)

echo -e "${GREEN}✓ Connection information retrieved${NC}\n"

# Step 5: Build and push Docker image
echo -e "${YELLOW}5️⃣  Building and pushing Docker image...${NC}"
echo "   Image: gcr.io/$PROJECT_ID/$SERVICE_NAME"

docker build \
    -t gcr.io/$PROJECT_ID/$SERVICE_NAME:latest \
    -t gcr.io/$PROJECT_ID/$SERVICE_NAME:$(date +%Y%m%d-%H%M%S) \
    . \
    --platform linux/amd64

echo -e "${YELLOW}   Pushing to Container Registry...${NC}"
docker push gcr.io/$PROJECT_ID/$SERVICE_NAME:latest

echo -e "${GREEN}✓ Docker image built and pushed${NC}\n"

# Step 6: Deploy to Cloud Run
echo -e "${YELLOW}6️⃣  Deploying to Cloud Run...${NC}"
echo "   Service: $SERVICE_NAME"
echo "   Region: $REGION"

gcloud run deploy $SERVICE_NAME \
    --image gcr.io/$PROJECT_ID/$SERVICE_NAME:latest \
    --region $REGION \
    --platform managed \
    --allow-unauthenticated \
    --memory 2Gi \
    --cpu 2 \
    --timeout 3600 \
    --set-env-vars \
        SQL_DSN="$SQL_DSN",\
        REDIS_CONN_STRING="$REDIS_CONN_STRING",\
        SESSION_SECRET="$SESSION_SECRET",\
        CRYPTO_SECRET="$CRYPTO_SECRET",\
        TZ="Asia/Shanghai" \
    --project=$PROJECT_ID

echo -e "${GREEN}✓ Deployment completed${NC}\n"

# Step 7: Get the service URL
echo -e "${YELLOW}7️⃣  Retrieving service information...${NC}"

SERVICE_URL=$(gcloud run services describe $SERVICE_NAME \
    --region $REGION \
    --project=$PROJECT_ID \
    --format='value(status.url)')

echo -e "${GREEN}✓ Service is live!${NC}\n"

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ Deployment Successful!${NC}"
echo -e "${BLUE}========================================${NC}\n"

echo -e "${YELLOW}📊 Deployment Summary:${NC}"
echo -e "  ${BLUE}Project ID:${NC} $PROJECT_ID"
echo -e "  ${BLUE}Service:${NC} $SERVICE_NAME"
echo -e "  ${BLUE}Region:${NC} $REGION"
echo -e "  ${BLUE}Service URL:${NC} ${GREEN}$SERVICE_URL${NC}"
echo -e "  ${BLUE}MySQL Instance:${NC} $INSTANCE_NAME"
echo -e "  ${BLUE}MySQL IP:${NC} $SQL_IP"
echo -e "  ${BLUE}Redis Instance:${NC} $REDIS_INSTANCE"
echo -e "  ${BLUE}Redis URL:${NC} redis://$REDIS_HOST:$REDIS_PORT\n"

echo -e "${YELLOW}🔐 Important Credentials (Save These!):${NC}"
echo -e "  ${BLUE}MySQL Root Password:${NC} $DB_PASS"
echo -e "  ${BLUE}SESSION_SECRET:${NC} $SESSION_SECRET"
echo -e "  ${BLUE}CRYPTO_SECRET:${NC} $CRYPTO_SECRET\n"

echo -e "${YELLOW}📝 Next Steps:${NC}"
echo -e "  1. Access the application at: ${GREEN}$SERVICE_URL${NC}"
echo -e "  2. Default login: admin / 123456"
echo -e "  3. Change default password immediately"
echo -e "  4. Configure your AI provider API keys"
echo -e "  5. Set up payment methods (Stripe, Epay, etc.)"
echo -e "  6. Bind custom domain (optional): gcloud run services update-traffic $SERVICE_NAME --to-revisions LATEST=100\n"

echo -e "${YELLOW}💾 Save These Credentials Securely:${NC}"
cat > .gcloud/deployment-credentials.txt << EOF
=====================================
New-API Google Cloud Deployment
=====================================
Deployment Date: $(date)
Project ID: $PROJECT_ID
Service Name: $SERVICE_NAME
Region: $REGION

Service URL: $SERVICE_URL

MySQL Instance: $INSTANCE_NAME
MySQL Host: $SQL_IP
MySQL Database: $DB_NAME
MySQL User: $DB_USER
MySQL Password: $DB_PASS
MySQL DSN: $SQL_DSN

Redis Instance: $REDIS_INSTANCE
Redis Host: $REDIS_HOST
Redis Port: $REDIS_PORT
Redis Connection: $REDIS_CONN_STRING

SESSION_SECRET: $SESSION_SECRET
CRYPTO_SECRET: $CRYPTO_SECRET

Cloud SQL Connection Name: $SQL_CONNECTION_NAME

Default Admin Credentials:
  Username: admin
  Password: 123456 (CHANGE THIS IMMEDIATELY)
=====================================
EOF

echo -e "${GREEN}✅ Credentials saved to: .gcloud/deployment-credentials.txt${NC}"
echo -e "${RED}⚠️  Keep this file secure and never commit it to git!${NC}\n"
