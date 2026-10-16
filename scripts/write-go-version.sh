#!/usr/bin/env bash
#
# Write the Go version to all three contract declarations from one argument.
#
# This is the blessed path for toolchain bumps (ADR-0003): go.mod is the
# authoritative source of the consumer-visible floor, and .go-version and
# .golangci.yml run.go must carry the identical full-patch string. Hand edits
# are acceptable; the tripwire (scripts/check-go-version.sh) is the
# verification of record either way and runs automatically at the end.
#
# Usage: scripts/write-go-version.sh 1.27.2

set -euo pipefail

cd "$(dirname "$0")/.."

if [ "$#" -ne 1 ]; then
	echo "usage: scripts/write-go-version.sh VERSION   (full-patch, e.g. 1.27.2)" >&2
	exit 1
fi

version="$1"

if ! grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$' <<<"$version"; then
	echo "FAIL: '$version' is not a full-patch version (want N.N.N) — two-part floors are forbidden (ADR-0003)" >&2
	exit 1
fi

rewrite() {
	local file="$1" pattern="$2" replacement="$3" label="$4"
	if [ "$(grep -cE "$pattern" "$file")" -ne 1 ]; then
		echo "FAIL: expected exactly one '$label' declaration in $file — inspect by hand" >&2
		exit 1
	fi
	sed -i -E "s|$pattern|$replacement|" "$file"
}

rewrite go.mod "^go [0-9]+\.[0-9]+\.[0-9]+$" "go $version" "go.mod go"
printf '%s\n' "$version" >.go-version
rewrite .golangci.yml "^  go: [0-9]+\.[0-9]+\.[0-9]+$" "  go: $version" ".golangci.yml run.go"

bash scripts/check-go-version.sh
echo "write-go-version: all three declarations now carry $version"
