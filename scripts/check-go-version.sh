#!/usr/bin/env bash
#
# Go version single-source tripwire.
#
# The Go version is declared in three places: the go.mod `go` directive
# (authoritative — CI builds with it via go-version-file), .go-version (local
# tooling convention: asdf/mise), and .golangci.yml `run.go` (lint language
# target). None of the three can be deleted without risk (local tooling breaks,
# lint default behavior is implicit), so instead of deleting declarations this
# script makes silent drift impossible: all three must agree exactly.
#
# Contract:
#   - go.mod `go` directive, .go-version content, and .golangci.yml `run.go`
#     must be the identical version string (e.g. 1.26.5).
#
# Usage: scripts/check-go-version.sh (run from anywhere; cd's to repo root)

set -euo pipefail

cd "$(dirname "$0")/.."

go_mod=$(awk '$1 == "go" { print $2; exit }' go.mod)
go_version_file=$(tr -d '[:space:]' <.go-version)
golangci_go=$(awk '/^[[:space:]]*go:/ { print $2; exit }' .golangci.yml)

status=0

check() {
	local name="$1" value="$2"
	if [ "$value" = "$go_mod" ]; then
		echo "ok: $name = $value (matches go.mod)"
	else
		echo "FAIL: $name = '$value' but go.mod says '$go_mod' — update the drifting declaration" >&2
		status=1
	fi
}

check "go.mod (authoritative)" "$go_mod"
check ".go-version" "$go_version_file"
check ".golangci.yml run.go" "$golangci_go"

if [ "$status" = 0 ]; then
	echo "go-version: PASS (all declarations agree on $go_mod)"
fi

exit "$status"
