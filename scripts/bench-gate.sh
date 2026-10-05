#!/usr/bin/env bash
#
# Quiet-machine gate for baseline-grade benchmark runs.
#
# This machine is multi-tenant: a load storm during a `go test -bench` run
# uniformly inflates ns/op ~2.6-3x while B/op and allocs/op stay byte-identical
# (contention signature). Such output looks plausible and is garbage as a
# baseline anchor, so this gate refuses to start (and aborts mid-run) unless
# the machine is quiet, validates a canary before committing to a full suite,
# and watches load while the suites run.
#
# Contract (acceptance criteria for a clean run):
#   - Entry: 6 consecutive 15s samples (90s sustained) where
#     (1-min load < 8 AND no storm) OR 1-min load < 5. A storm = load rising
#     by >= 3 within one sample interval — the 1-min loadavg lags, so a bare
#     threshold would fire into load troughs that storms fill minutes later.
#   - Canary: RawCodec_Encode mean over 3 runs <= 25 ns/op
#     (clean ~11-16n, polluted ~40n+). Abort the attempt otherwise.
#   - Watch: 1-min load sampled every 20s during each suite; abort at >= 12.
#     The running peak is recorded to <out>.loadpeak for post-hoc audit.
#   - Retry: up to 8 attempts within a 10h deadline.
#   - Success: both suites end `ok` -> <out>.ok markers; exit 0.
#
# Usage: scripts/bench-gate.sh OUT_DIR
#   OUT_DIR receives codec-bench-{v1,v2}.txt (+ .ok, .txt.loadpeak).
#   Use a path that survives reboots and /tmp cleaners, and run inside the
#   devShell (needs `go` on PATH). Expect ~30 min per suite.

set -euo pipefail

cd "$(dirname "$0")/.."

load1() {
	awk '{print $1}' /proc/loadavg
}

# True when the float in $1 compares < the float in $2.
float_lt() {
	awk -v a="$1" -v b="$2" 'BEGIN { exit (a < b) ? 0 : 1 }'
}

wait_quiet() {
	local consecutive=0 prev="" sample storm
	while [ "$consecutive" -lt 6 ]; do
		sample=$(load1)
		storm=0
		if [ -n "$prev" ] && ! float_lt "$sample" "$(awk -v a="$prev" 'BEGIN { print a + 3 }')"; then
			storm=1
		fi
		if float_lt "$sample" 8 && [ "$storm" = 0 ] || float_lt "$sample" 5; then
			consecutive=$((consecutive + 1))
		else
			consecutive=0
		fi
		prev=$sample
		if [ "$consecutive" -lt 6 ]; then
			sleep 15
		fi
	done
}

canary_ok() {
	local out mean
	out=$(env -u GOEXPERIMENT go test -run '^$' -bench '^BenchmarkRawCodec_Encode$' \
		-benchmem -count=3 -timeout 5m 2>&1) || true
	echo "$out" | grep '^BenchmarkRawCodec_Encode'
	mean=$(echo "$out" | awk '$1 ~ /^BenchmarkRawCodec_Encode/ { s += $3; n++ }
		END { if (n > 0) printf "%.1f", s / n; else print "9999" }')
	echo "canary: RawCodec_Encode mean ${mean}n/op (threshold 25n)"
	float_lt "$mean" 25.001
}

# run_suite MODE OUTFILE — MODE is v1 (default env) or v2 (GOEXPERIMENT=jsonv2).
run_suite() {
	local mode="$1" out="$2" pid sample peak
	: >"$out.loadpeak"
	peak=0
	if [ "$mode" = v1 ]; then
		env -u GOEXPERIMENT go test -run '^$' -bench . -benchmem -count=10 \
			-timeout 40m >"$out" &
	else
		GOEXPERIMENT=jsonv2 go test -run '^$' -bench . -benchmem -count=10 \
			-timeout 40m >"$out" &
	fi
	pid=$!
	while kill -0 "$pid" 2>/dev/null; do
		sample=$(load1)
		if ! float_lt "$sample" "$peak"; then
			peak=$sample
		fi
		printf '%s\n' "$peak" >"$out.loadpeak"
		if ! float_lt "$sample" 12; then
			echo "ABORT: 1-min load $sample >= 12 during $mode suite, killing run" >&2
			kill "$pid" 2>/dev/null || true
			sleep 2
			kill -9 "$pid" 2>/dev/null || true
			pkill -P "$pid" 2>/dev/null || true
			return 1
		fi
		sleep 20
	done
	wait "$pid"
}

suite_ok() {
	local out="$1"
	tail -n 1 "$out" | grep -q '^ok'
}

main() {
	local out_dir attempt deadline
	out_dir="${1:?usage: scripts/bench-gate.sh OUT_DIR (somewhere durable, NOT /tmp)}"
	mkdir -p "$out_dir"
	deadline=$(( $(date +%s) + 10 * 3600 ))
	attempt=0
	while [ "$(date +%s)" -lt "$deadline" ]; do
		attempt=$((attempt + 1))
		if [ "$attempt" -gt 8 ]; then
			echo "FAIL: 8 attempts exhausted without a clean pair of suites" >&2
			exit 1
		fi
		echo "== attempt $attempt/8: waiting for sustained quiet =="
		wait_quiet
		echo "== attempt $attempt/8: canary =="
		if ! canary_ok; then
			echo "canary above threshold — machine not quiet after all; retrying" >&2
			sleep 60
			continue
		fi
		echo "== attempt $attempt/8: v1 suite (default env) =="
		if run_suite v1 "$out_dir/codec-bench-v1.txt" && suite_ok "$out_dir/codec-bench-v1.txt"; then
			: >"$out_dir/codec-bench-v1.ok"
		else
			echo "v1 suite aborted or failed; retrying" >&2
			sleep 60
			continue
		fi
		echo "== attempt $attempt/8: v2 suite (GOEXPERIMENT=jsonv2) =="
		if run_suite v2 "$out_dir/codec-bench-v2.txt" && suite_ok "$out_dir/codec-bench-v2.txt"; then
			: >"$out_dir/codec-bench-v2.ok"
		else
			echo "v2 suite aborted or failed; retrying" >&2
			sleep 60
			continue
		fi
		echo "PASS: both suites clean under the gate:"
		echo "  $out_dir/codec-bench-v1.txt (peak 1-min load $(cat "$out_dir/codec-bench-v1.txt.loadpeak"))"
		echo "  $out_dir/codec-bench-v2.txt (peak 1-min load $(cat "$out_dir/codec-bench-v2.txt.loadpeak"))"
		exit 0
	done
	echo "FAIL: 10h deadline exceeded" >&2
	exit 1
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then
	main "$@"
fi
