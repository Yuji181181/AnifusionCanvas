#!/usr/bin/env bash
set -euo pipefail

: "${GCP_PROJECT_ID:?GCP_PROJECT_ID is required}"
: "${GCP_REGION:?GCP_REGION is required}"
: "${GCP_ARTIFACT_REPOSITORY:?GCP_ARTIFACT_REPOSITORY is required}"
: "${GCP_SERVICE_NAME:?GCP_SERVICE_NAME is required}"
: "${GCP_SERVICE_ACCOUNT_EMAIL:?GCP_SERVICE_ACCOUNT_EMAIL is required}"

ROOT_DIR=$(cd "$(dirname "$0")/../.." && pwd)
IMAGE_URI="${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/${GCP_ARTIFACT_REPOSITORY}/${GCP_SERVICE_NAME}:latest"
CLOUD_RUN_CPU="${CLOUD_RUN_CPU:-2}"
CLOUD_RUN_MEMORY="${CLOUD_RUN_MEMORY:-2Gi}"
CLOUD_RUN_TIMEOUT="${CLOUD_RUN_TIMEOUT:-900s}"
CLOUD_RUN_CONCURRENCY="${CLOUD_RUN_CONCURRENCY:-4}"

"$ROOT_DIR/infra/cloud-run/docker-login.sh"

cd "$ROOT_DIR/apps/api"
docker build -t "$IMAGE_URI" .
docker push "$IMAGE_URI"

# Clear an older secret binding before setting the optional public URL as a
# literal env var. Cloud Run does not allow changing an env var's value type in
# one deploy command.
gcloud run services update "$GCP_SERVICE_NAME" \
  --project "$GCP_PROJECT_ID" \
  --region "$GCP_REGION" \
  --remove-secrets R2_PUBLIC_BASE_URL \
  >/dev/null || true

gcloud run deploy "$GCP_SERVICE_NAME" \
  --image "$IMAGE_URI" \
  --project "$GCP_PROJECT_ID" \
  --region "$GCP_REGION" \
  --platform managed \
  --allow-unauthenticated \
  --service-account "$GCP_SERVICE_ACCOUNT_EMAIL" \
  --cpu "$CLOUD_RUN_CPU" \
  --memory "$CLOUD_RUN_MEMORY" \
  --timeout "$CLOUD_RUN_TIMEOUT" \
  --concurrency "$CLOUD_RUN_CONCURRENCY" \
  --no-cpu-throttling \
  --cpu-boost \
  --set-env-vars APP_ENV=production,API_PORT=8080,FRONTEND_ORIGIN=https://anifusion-canvas.pages.dev,R2_REGION=auto,R2_PUBLIC_BASE_URL=,STUDIO_STORE=database,REPLICATE_MODE=demo \
  --set-secrets DATABASE_URL=DATABASE_URL:latest,R2_ACCOUNT_ID=R2_ACCOUNT_ID:latest,R2_ACCESS_KEY_ID=R2_ACCESS_KEY_ID:latest,R2_SECRET_ACCESS_KEY=R2_SECRET_ACCESS_KEY:latest,R2_BUCKET=R2_BUCKET:latest,R2_ENDPOINT_URL=R2_ENDPOINT_URL:latest,REPLICATE_API_TOKEN=REPLICATE_API_TOKEN:latest
