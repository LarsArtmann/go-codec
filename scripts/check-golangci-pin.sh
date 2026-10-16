#!/usr/bin/env bash
#
# golangci-lint CI-pin floor guard.
#
# The golangci-lint version installed in ci.yml must be at least
# MIN_GOLANGCI. Rationale: golangci's analyzers are coupled to the stdlib
# they analyze — v2.12.2 panicked (staticcheck buildir) on the go1.27.1
# stdlib, red both lint legs on 2026-10-15; v2.14.0 fixed it. When a
# toolchain bump requires a newer golangci, raise MIN_GOLANGCI in the same
# change, deliberately — dependabot cannot manage `go install pkg@version`
# lines, so nothing else moves this pin.
#
# Usage: scripts/check-golangci-pin.sh (run from anywhere; cd's to repo root)

set -euo pipefail

cd "$(dirname "$0")/.."

MIN_GOLANGCI="2.14.0"

ci_count=$(grep -cE 'golangci/golangci-lint/v2/cmd/golangci-lint@v[0-9]+\.[0-9]+\.[0-9]+' .github/workflows/ci.yml || true)
if [ "$ci_count" -ne 1 ]; then
	echo "FAIL: expected exactly 1 pinned golangci-lint install in ci.yml, found $ci_count" >&2
	exit 1
fi

pin=$(grep -oE 'golangci-lint/v2/cmd/golangci-lint@v[0-9]+\.[0-9]+\.[0-9]+' .github/workflows/ci.yml | sed 's/.*@v//')

ver_ge() {
	awk -v a="$1" -v b="$2" 'BEGIN {
		split(a, x, "."); split(b, y, ".")
		for (i = 1; i <= 3; i++) {
			if (x[i] + 0 > y[i] + 0) exit 0
			if (x[i] + 0 < y[i] + 0) exit 1
		}
		exit 0
	}'
}

if ! ver_ge "$pin" "$MIN_GOLANGCI"; then
	echo "FAIL: golangci-lint pin v$pin is below the floor v$MIN_GOLANGCI" >&2
	echo "  the floor encodes the oldest release known to work with the current Go stdlib;" >&2
	echo "  bump the ci.yml pin (and MIN_GOLANGCI here) together — see the 2026-10-15 buildir panic" >&2
	exit 1
fi

echo "golangci-pin: PASS (v$pin >= floor v$MIN_GOLANGCI)"
