#!/usr/bin/env bash
#
# Error-code registry drift tripwire.
#
# Fails when the set of `codec.*` error-code literals in non-test source
# differs from the set registered in docs/error-codes.md, in either direction.
# Companion to docs/error-codes.md (the registry) and the same drift-lock
# pattern as check-features-planned.sh / check-go-version.sh.
#
# Contract:
#   - Source set: every "codec.<name>" string literal in non-test .go files
#     tracked by git (both JSON build variants included).
#   - Registry set: every codec.<name> token in docs/error-codes.md.
#   - Both sets must be equal (uniqueness is implied: sets are deduplicated).
#
# Usage: scripts/check-error-codes.sh (run from anywhere; cd's to repo root)

set -euo pipefail

cd "$(dirname "$0")/.."

registry="docs/error-codes.md"

if [ ! -f "$registry" ]; then
	echo "FAIL: $registry missing — the error-code registry is the doc side of this tripwire" >&2
	exit 1
fi

src_tmp=$(mktemp)
doc_tmp=$(mktemp)
trap 'rm -f "$src_tmp" "$doc_tmp"' EXIT

# Source set: code literals in non-test Go files.
git ls-files '*.go' | grep -v '_test\.go$' | xargs grep -ohE '"codec\.[a-z0-9_]+"' |
	tr -d '"' | sort -u >"$src_tmp"

# Registry set: backticked code tokens in the registry doc (the registration
# convention — plain-text mentions like "codec.go" filenames do not register).
grep -ohE '`codec\.[a-z0-9_]+`' "$registry" | tr -d '`' | sort -u >"$doc_tmp"

# Well-formedness is guaranteed by the extraction regexes (codec.[a-z0-9_]+
# only), so the sets can be diffed directly.
status=0

# In source but not in the registry.
if comm -23 "$src_tmp" "$doc_tmp" | grep -q .; then
	echo "FAIL: codes emitted by source but missing from $registry:" >&2
	comm -23 "$src_tmp" "$doc_tmp" | sed 's/^/  /' >&2
	status=1
fi

# In the registry but not in source (phantom or renamed codes).
if comm -13 "$src_tmp" "$doc_tmp" | grep -q .; then
	echo "FAIL: codes registered in $registry but absent from source (renamed or removed?):" >&2
	comm -13 "$src_tmp" "$doc_tmp" | sed 's/^/  /' >&2
	status=1
fi

count=$(wc -l <"$src_tmp")

if [ "$status" = 0 ]; then
	echo "error-codes: PASS ($count codes registered, source and registry in sync)"
fi

exit "$status"
