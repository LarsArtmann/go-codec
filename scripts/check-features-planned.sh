#!/usr/bin/env bash
#
# FEATURES.md drift tripwire.
#
# Fails when a symbol that FEATURES.md marks as ⚪ PLANNED actually resolves in
# the package API ("no code exists yet" claims that have silently shipped).
# This is the CI-side lock for the drift class caught on 2026-08-14, where
# FEATURES.md still listed DeterministicCodec as PLANNED two sessions after the
# code had shipped (see docs/status/2026-08-14_20-07_self-review-resume-verification-drift-repair.md §d-1).
#
# Contract:
#   - Only the "## Status legend" table is exempt; every other table row whose
#     first cell is a single backticked exported Go identifier AND that
#     contains a ⚪ PLANNED status cell is checked.
#   - A checked symbol must NOT resolve via `go doc`.
#
# Usage: scripts/check-features-planned.sh (run from anywhere; cd's to repo root)

set -euo pipefail

cd "$(dirname "$0")/.."

pkg="github.com/larsartmann/go-codec"
status=0
in_legend=0
checked=0

while IFS= read -r line; do
	case "$line" in
	'## Status legend'*)
		in_legend=1
		continue
		;;
	esac

	if [ "$in_legend" = 1 ]; then
		case "$line" in
		'|'* | '') continue ;;
		*) in_legend=0 ;;
		esac
	fi

	case "$line" in
	*'⚪'*)
		sym=$(printf '%s' "$line" | awk -F'|' '{cell=$2; gsub(/^[ \t`]+|[ \t`]+$/, "", cell); print cell}')

		if printf '%s' "$sym" | grep -Eq '^[A-Z][A-Za-z0-9_]*$'; then
			checked=$((checked + 1))
			if go doc "$pkg" "$sym" >/dev/null 2>&1; then
				echo "FAIL: FEATURES.md marks '$sym' as PLANNED, but it resolves via 'go doc' — shipped code is drifting the docs" >&2
				status=1
			else
				echo "ok: PLANNED symbol '$sym' does not resolve (docs honest)"
			fi
		else
			echo "note: skipping PLANNED row whose first cell is not a bare exported identifier: $line" >&2
		fi
		;;
	esac
done <FEATURES.md

if [ "$checked" = 0 ]; then
	echo "features-planned: PASS (0 PLANNED symbol rows found)"
else
	echo "features-planned: checked $checked PLANNED symbol row(s)"
fi

exit "$status"
