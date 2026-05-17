#!/usr/bin/env bash
set -euo pipefail

: "${GCP_PROJECT_ID:?GCP_PROJECT_ID is required}"
: "${GCP_REGION:?GCP_REGION is required}"
: "${GCP_ARTIFACT_REPOSITORY:?GCP_ARTIFACT_REPOSITORY is required}"
: "${GCP_SERVICE_NAME:?GCP_SERVICE_NAME is required}"

ROOT_DIR=$(cd "$(dirname "$0")/../.." && pwd)
IMAGE_URI="${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/${GCP_ARTIFACT_REPOSITORY}/${GCP_SERVICE_NAME}:latest"
ENV_FILE=$(mktemp)
CLOUD_RUN_CPU="${CLOUD_RUN_CPU:-2}"
CLOUD_RUN_MEMORY="${CLOUD_RUN_MEMORY:-2Gi}"
CLOUD_RUN_TIMEOUT="${CLOUD_RUN_TIMEOUT:-900s}"
CLOUD_RUN_CONCURRENCY="${CLOUD_RUN_CONCURRENCY:-4}"

cleanup() {
  rm -f "$ENV_FILE"
}
trap cleanup EXIT

set -a
while IFS= read -r line || [ -n "$line" ]; do
  case "$line" in
    ''|'#'*) continue ;;
  esac
  export "$line"
done < "$ROOT_DIR/.env"
set +a

export FRONTEND_ORIGIN="${FRONTEND_ORIGIN:-https://anifusion-canvas.pages.dev}"
export STUDIO_STORE="${STUDIO_STORE:-database}"
"$ROOT_DIR/infra/cloud-run/render-env.sh" > "$ENV_FILE"

"$ROOT_DIR/infra/cloud-run/docker-login.sh"

cd "$ROOT_DIR/apps/api"
docker build -t "$IMAGE_URI" .
docker push "$IMAGE_URI"

gcloud run deploy "$GCP_SERVICE_NAME" \
  --image "$IMAGE_URI" \
  --project "$GCP_PROJECT_ID" \
  --region "$GCP_REGION" \
  --platform managed \
  --allow-unauthenticated \
  --cpu "$CLOUD_RUN_CPU" \
  --memory "$CLOUD_RUN_MEMORY" \
  --timeout "$CLOUD_RUN_TIMEOUT" \
  --concurrency "$CLOUD_RUN_CONCURRENCY" \
  --no-cpu-throttling \
  --cpu-boost \
  --env-vars-file "$ENV_FILE"
