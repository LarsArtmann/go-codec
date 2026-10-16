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
#     Known blind spot: a load burst shorter than the 1-min loadavg window
#     (sub-~60s) can pollute rows without ever crossing 12 — after acceptance,
#     eyeball benchstat variance against the previous baseline before
#     enshrining (suspiciously wide ranges on normally-tight rows = reject).
#   - Retry: up to 8 attempts within a 10h deadline.
#   - Success: both suites end `ok` -> <out>.ok markers; exit 0.
#
# Usage: scripts/bench-gate.sh OUT_DIR
#   OUT_DIR receives codec-bench-{v1,v2}.txt (+ .ok, .txt.loadpeak).
#   Use a path that survives reboots and /tmp cleaners, and run inside the
#   devShell (needs `go` on PATH). Expect ~30 min per suite.
#
# Other modes:
#   scripts/bench-gate.sh --selftest   run the wait_quiet acceptance self-test
#   scripts/bench-gate.sh --help       show this usage
#
# Test hooks (not for production runs):
#   BENCH_GATE_SNAPSHOTS=DIR    load1() reads snapshots DIR/1,DIR/2,... in
#                               round-robin instead of /proc/loadavg (selftest)
#   BENCH_GATE_SAMPLE_SECS=N    wait_quiet sleep between samples (default 15)

set -euo pipefail

cd "$(dirname "$0")/.."

# Selftest sampler state (see BENCH_GATE_SNAPSHOTS in the header).
_snapshot_count=0

load1() {
	if [ -n "${BENCH_GATE_SNAPSHOTS:-}" ]; then
		local n
		n=$(cat "$BENCH_GATE_SNAPSHOTS/.counter" 2>/dev/null || echo 0)
		n=$((n + 1))
		printf '%s\n' "$n" >"$BENCH_GATE_SNAPSHOTS/.counter"
		awk '{print $1}' "$BENCH_GATE_SNAPSHOTS/$(( ((n - 1) % _snapshot_count) + 1 ))"
	else
		awk '{print $1}' /proc/loadavg
	fi
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
			sleep "$BENCH_GATE_SAMPLE_SECS"
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

# wait_quiet acceptance self-test: four synthetic load histories driven
# through the snapshot sampler (BENCH_GATE_SNAPSHOTS) assert that sustained
# quiet converges quickly while storm/band/high histories never do.
selftest() {
	local dir expected failures=0 name rc
	dir=$(mktemp -d)
	trap 'rm -rf "$dir"' RETURN

	# runs wait_quiet against DIR with a hard timeout; echoes 0 when it
	# converged (exit 0), 1 when it was still waiting when the timeout fired.
	run_wait() {
		if timeout "$2" env "BENCH_GATE_SNAPSHOTS=$1" "$0" --selftest-probe >/dev/null 2>&1; then
			echo 0
		else
			echo 1
		fi
	}

	scenario() {
		name="$1"
		shift
		rc=$(run_wait "$@")
		if [ "$rc" = "$expected" ]; then
			echo "ok: $name"
		else
			echo "FAIL: $name (got converge=$rc, want $expected)" >&2
			failures=$((failures + 1))
		fi
	}

	BENCH_GATE_SAMPLE_SECS=15

	mkdir -p "$dir/quiet"
	printf '2.0 2.0 2.0 5/100 1\n' >"$dir/quiet/1"
	expected=0 scenario "flat quiet load converges" "$dir/quiet" 5

	mkdir -p "$dir/band"
	printf '6.5 6.5 6.5 5/100 1\n' >"$dir/band/1"
	expected=0 scenario "load in the 5..8 band without a storm converges" "$dir/band" 5

	mkdir -p "$dir/high"
	printf '30.0 30.0 30.0 5/100 1\n' >"$dir/high/1"
	expected=1 scenario "sustained load >= 8 never converges" "$dir/high" 2

	mkdir -p "$dir/storm"
	printf '2.0 2.0 2.0 5/100 1\n' >"$dir/storm/1"
	printf '7.0 7.0 7.0 5/100 1\n' >"$dir/storm/2"
	expected=1 scenario "rising storm under 8 never converges" "$dir/storm" 2

	if [ "$failures" -gt 0 ]; then
		echo "selftest: FAIL ($failures scenario(s))" >&2
		return 1
	fi
	echo "selftest: PASS (4 scenarios)"
}

main() {
	local out_dir attempt deadline
	out_dir="${1:?usage: scripts/bench-gate.sh OUT_DIR (somewhere durable, NOT /tmp)}"
	mkdir -p "$out_dir"
	deadline=$(($(date +%s) + 10 * 3600))
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
	case "${1:-}" in
		--selftest)
			selftest
			;;
		--selftest-probe)
			# internal helper of --selftest: one wait_quiet pass with 0s sleeps
			BENCH_GATE_SAMPLE_SECS=0
			_snapshot_count=$(find "${BENCH_GATE_SNAPSHOTS:?}" -maxdepth 1 -type f | wc -l)
			wait_quiet
			;;
		--help | -h)
			echo "usage: scripts/bench-gate.sh OUT_DIR" >&2
			echo "       scripts/bench-gate.sh --selftest" >&2
			echo "       scripts/bench-gate.sh --help" >&2
			;;
		*)
			main "$@"
			;;
	esac
fi
