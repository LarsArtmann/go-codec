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
#     must be the identical version string (e.g. 1.27.1).
#   - The .golangci.yml token is quote-stripped before compare (YAML quoting
#     style must not defeat the exact-match) and a 2-part go.mod value is
#     rejected outright: it parses as a YAML float in .golangci.yml, which is
#     what left v0.3.1 tagged with this gate red (2026-10-16).
#
# Usage: scripts/check-go-version.sh (run from anywhere; cd's to repo root)

set -euo pipefail

cd "$(dirname "$0")/.."

go_mod=$(awk '$1 == "go" { print $2; exit }' go.mod)
go_version_file=$(tr -d '[:space:]' <.go-version)
golangci_go=$(awk '/^[[:space:]]*go:/ { print $2; exit }' .golangci.yml)
golangci_go=${golangci_go//\"/}
golangci_go=${golangci_go//\'/}

case "$go_mod" in
*.*.*) ;;
*)
	echo "FAIL: go.mod says '$go_mod' (2-part). The repo contract is a full-patch version (e.g. 1.27.1) in all three declarations: a 2-part value parses as a YAML float in .golangci.yml, making the exact-match compare unsatisfiable." >&2
	exit 1
	;;
esac

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
