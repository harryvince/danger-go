#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${DIST_DIR:-"${ROOT_DIR}/dist"}"
TAG="${TAG:-}"
COMMIT="${COMMIT:-}"
DATE="${DATE:-}"

if [[ -z "${TAG}" ]]; then
  TAG="$(git -C "${ROOT_DIR}" describe --tags --exact-match 2>/dev/null || git -C "${ROOT_DIR}" describe --tags --abbrev=0 2>/dev/null || echo "danger-go-v0.0.0")"
fi

if [[ -z "${COMMIT}" ]]; then
  COMMIT="$(git -C "${ROOT_DIR}" rev-list -n 1 "${TAG}" 2>/dev/null || git -C "${ROOT_DIR}" rev-parse HEAD)"
fi

if [[ -z "${DATE}" ]]; then
  DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
fi

VERSION="${TAG#danger-go-v}"
VERSION="${VERSION#v}"

LDFLAGS="-s -w"
LDFLAGS+=" -X github.com/harryvince/danger-go/internal/version.Version=${VERSION}"
LDFLAGS+=" -X github.com/harryvince/danger-go/internal/version.Commit=${COMMIT}"
LDFLAGS+=" -X github.com/harryvince/danger-go/internal/version.Date=${DATE}"

rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

build_target() {
  local goos="$1"
  local goarch="$2"
  local ext=""

  if [[ "${goos}" == "windows" ]]; then
    ext=".exe"
  fi

  local name="danger-go_${VERSION}_${goos}_${goarch}"
  local workdir="${DIST_DIR}/${name}"
  mkdir -p "${workdir}"

  CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" go build \
    -trimpath \
    -ldflags "${LDFLAGS}" \
    -o "${workdir}/danger-go${ext}" \
    "${ROOT_DIR}/cmd/danger-go"

  cp "${ROOT_DIR}/README.md" "${workdir}/README.md"
  cp "${ROOT_DIR}/CHANGELOG.md" "${workdir}/CHANGELOG.md"

  tar -C "${DIST_DIR}" -czf "${DIST_DIR}/${name}.tar.gz" "${name}"

  rm -rf "${workdir}"
}

for goos in linux darwin windows; do
  for goarch in amd64 arm64; do
    build_target "${goos}" "${goarch}"
  done
done

(cd "${DIST_DIR}" && sha256sum danger-go_* > checksums.txt)
