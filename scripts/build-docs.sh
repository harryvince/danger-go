#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

uv run zensical build --clean
mkdir -p "${ROOT_DIR}/site/schema"
cp "${ROOT_DIR}/schema/danger-go.schema.json" "${ROOT_DIR}/site/schema/danger-go.schema.json"
