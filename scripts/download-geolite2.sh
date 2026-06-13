#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA_DIR="${ROOT_DIR}/data"
DB_FILE="${DATA_DIR}/GeoLite2-Country.mmdb"

LICENSE_KEY="${MAXMIND_LICENSE_KEY:-}"
if [[ -z "${LICENSE_KEY}" && -f "${ROOT_DIR}/.env" ]]; then
	# shellcheck disable=SC1091
	source "${ROOT_DIR}/.env"
	LICENSE_KEY="${MAXMIND_LICENSE_KEY:-}"
fi

if [[ -z "${LICENSE_KEY}" ]]; then
	echo "error: set MAXMIND_LICENSE_KEY in environment or .env" >&2
	exit 1
fi

mkdir -p "${DATA_DIR}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

ARCHIVE="${TMP_DIR}/GeoLite2-Country.tar.gz"
URL="https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=${LICENSE_KEY}&suffix=tar.gz"

echo "Downloading GeoLite2-Country..."
curl -fsSL "${URL}" -o "${ARCHIVE}"

tar -xzf "${ARCHIVE}" -C "${TMP_DIR}"
MMDB="$(find "${TMP_DIR}" -name 'GeoLite2-Country.mmdb' -print -quit)"
if [[ -z "${MMDB}" ]]; then
	echo "error: GeoLite2-Country.mmdb not found in archive" >&2
	exit 1
fi

mv -f "${MMDB}" "${DB_FILE}"
echo "Saved ${DB_FILE}"
