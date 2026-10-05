#!/usr/bin/env python3
#
# Raw-benchmark-output comparator (baseline vs candidate).
#
# Parses two raw `go test -bench ... -benchmem -count=N` output files, computes
# the per-benchmark mean of each op-normalized metric (ns/op, B/op, allocs/op)
# across the N repetitions, and sanity-bounds the candidate/baseline ratio of
# every mean: outside [--lower, --upper] (default 0.33..3.0) the comparator
# exits 1. Those bounds are deliberately wide — they catch gross drift (wrong
# toolchain, thermal throttling, a broken run), NOT fine-grained regressions;
# use benchstat for statistical A/B comparison.
#
# Contract:
#   - Input: raw `go test -bench` stdout files (headers like goos:/goarch: and
#     PASS/ok/fail lines are ignored; sub-benchmarks with `/` are supported).
#   - Names are normalized by stripping the trailing `-<GOMAXPROCS>` suffix, so
#     runs from machines with different core counts still match.
#   - Only benchmarks present in BOTH files are compared; names unique to one
#     file are reported as warnings on stderr (benchmark-set drift).
#   - Mean of zero (e.g. 0 B/op) in both files counts as ratio 1.0; a mean of
#     zero on one side only counts as out of bounds.
#
# Usage: scripts/bench-compare.py BASELINE.txt CANDIDATE.txt [--lower 0.33 --upper 3.0]
# Exit:  0 = all ratios within bounds, 1 = at least one ratio out of bounds,
#        2 = usage or parse error (unreadable file, no benchmarks, no overlap).

import argparse
import re
import sys
from pathlib import Path

BENCH_LINE = re.compile(r"^(Benchmark\S*)\s+\d+\s+(.*)$")
NAME_PROCS_SUFFIX = re.compile(r"-\d+$")
METRICS = ("ns/op", "B/op", "allocs/op")


def parse_file(path: Path) -> dict[str, dict[str, list[float]]]:
    """Map benchmark name -> metric label -> list of values (one per count)."""
    results: dict[str, dict[str, list[float]]] = {}
    for line in path.read_text(encoding="utf-8", errors="replace").splitlines():
        match = BENCH_LINE.match(line.strip())
        if not match:
            continue
        name = NAME_PROCS_SUFFIX.sub("", match.group(1)[len("Benchmark") :])
        metrics = results.setdefault(name, {})
        tokens = match.group(2).split()
        for value, label in zip(tokens[0::2], tokens[1::2]):
            if label not in METRICS:
                continue
            try:
                metrics.setdefault(label, []).append(float(value))
            except ValueError:
                print(
                    f"WARN: unparseable value {value!r} for {name} {label}",
                    file=sys.stderr,
                )
    return results


def mean(values: list[float]) -> float:
    return sum(values) / len(values)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Sanity-gate raw go benchmark output: baseline vs candidate mean ratios (wide bounds; benchstat does statistics)."
    )
    parser.add_argument("baseline", type=Path)
    parser.add_argument("candidate", type=Path)
    parser.add_argument(
        "--lower", type=float, default=0.33, help="lower sanity bound (default 0.33)"
    )
    parser.add_argument(
        "--upper", type=float, default=3.0, help="upper sanity bound (default 3.0)"
    )
    args = parser.parse_args()

    for path in (args.baseline, args.candidate):
        if not path.is_file():
            print(f"FAIL: {path} is not a readable file", file=sys.stderr)
            return 2

    baseline = parse_file(args.baseline)
    candidate = parse_file(args.candidate)
    if not baseline or not candidate:
        print("FAIL: no benchmark lines parsed from one of the files", file=sys.stderr)
        return 2

    for label, names in (("baseline", baseline), ("candidate", candidate)):
        other = candidate if label == "baseline" else baseline
        for name in sorted(set(names) - set(other)):
            print(
                f"WARN: {name} only present in {label} file (skipped)", file=sys.stderr
            )

    common = sorted(set(baseline) & set(candidate))
    if not common:
        print("FAIL: no benchmarks common to both files", file=sys.stderr)
        return 2

    header = f"{'benchmark':<48} {'metric':<10} {'baseline':>12} {'candidate':>12} {'ratio':>8}  verdict"
    print(header)
    print("-" * len(header))

    failures = 0
    for name in common:
        for metric in METRICS:
            base_values = baseline[name].get(metric)
            cand_values = candidate[name].get(metric)
            if not base_values or not cand_values:
                continue
            base_mean, cand_mean = mean(base_values), mean(cand_values)
            if base_mean == 0 and cand_mean == 0:
                ratio = 1.0
            elif base_mean == 0:
                ratio = float("inf")
            else:
                ratio = cand_mean / base_mean
            verdict = "OK" if args.lower <= ratio <= args.upper else "OUT OF BOUNDS"
            if verdict != "OK":
                failures += 1
            ratio_text = "inf" if ratio == float("inf") else f"{ratio:.3f}"
            print(
                f"{name:<48} {metric:<10} {base_mean:>12.2f} {cand_mean:>12.2f} {ratio_text:>8}  {verdict}"
            )

    compared = sum(
        1
        for name in common
        for metric in METRICS
        if baseline[name].get(metric) and candidate[name].get(metric)
    )
    print(
        f"\n{compared} metric means compared across {len(common)} benchmarks; {failures} out of bounds"
    )
    print(
        f"bounds: [{args.lower}, {args.upper}] — wide sanity gate only; use benchstat for A/B statistics"
    )
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
