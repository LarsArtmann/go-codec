#!/usr/bin/env bash
#
# erraudit version single-source tripwire.
#
# The erraudit version is pinned in TWO places that must agree:
#   - .github/workflows/ci.yml   (error-audit install step)
#   - flake.nix                  (the .#erraudit app / devShell runner)
# This fails the build when they drift. Dependabot cannot manage
# `go install pkg@version` lines (gomod ecosystem only) and Renovate is not
# installed on this repo, so bumps are deliberate by design — see ROADMAP
# "Monthly erraudit release check; bump the pin deliberately".
#
# Usage: scripts/check-erraudit-version.sh (run from anywhere; cd's to repo root)

set -euo pipefail

cd "$(dirname "$0")/.."

ci_count=$(grep -cE 'erraudit/cmd/erraudit@v[0-9]+\.[0-9]+\.[0-9]+' .github/workflows/ci.yml || true)
flake_count=$(grep -cE 'erraudit/cmd/erraudit@v[0-9]+\.[0-9]+\.[0-9]+' flake.nix || true)

if [ "$ci_count" -ne 1 ] || [ "$flake_count" -ne 1 ]; then
	echo "FAIL: expected exactly 1 pinned erraudit version in ci.yml and flake.nix, found ci=$ci_count flake=$flake_count" >&2
	exit 1
fi

ci_version=$(grep -oE 'erraudit/cmd/erraudit@v[0-9]+\.[0-9]+\.[0-9]+' .github/workflows/ci.yml | sed 's/.*@//')
flake_version=$(grep -oE 'erraudit/cmd/erraudit@v[0-9]+\.[0-9]+\.[0-9]+' flake.nix | sed 's/.*@//')

if [ "$ci_version" != "$flake_version" ]; then
	echo "FAIL: erraudit pin drifted: ci.yml=$ci_version flake.nix=$flake_version — update both together" >&2
	exit 1
fi

echo "erraudit-version: PASS ($ci_version pinned identically in ci.yml and flake.nix)"
