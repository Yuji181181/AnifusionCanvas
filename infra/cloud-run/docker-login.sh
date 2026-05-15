#!/usr/bin/env bash
set -euo pipefail

: "${GCP_REGION:?GCP_REGION is required}"

gcloud auth print-access-token | docker login -u oauth2accesstoken --password-stdin "https://${GCP_REGION}-docker.pkg.dev"
