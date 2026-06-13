#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IP="${1:-}"
DB="${ROOT_DIR}/data/GeoLite2-Country.mmdb"

if [[ -z "${IP}" ]]; then
	echo "usage: bash scripts/lookup-ip-region.sh <IP>" >&2
	exit 1
fi

if [[ ! -f "${DB}" ]]; then
	echo "error: database not found at ${DB}" >&2
	exit 1
fi

echo "=== GeoLite2 lookup ==="
docker run --rm \
	-v "${ROOT_DIR}:/app" \
	-w /app \
	golang:1.21-alpine \
	sh -c "go run ./scripts/geoip-lookup --db /app/data/GeoLite2-Country.mmdb --ip ${IP}"

echo
echo "=== ip-api reference (network) ==="
curl -fsS "http://ip-api.com/json/${IP}?fields=status,countryCode,query" | sed 's/$/\n/'
