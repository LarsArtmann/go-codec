# Benchmark Baseline

Reference performance baseline for regression comparison, recorded 2026-10-05,
one suite per JSON mode. Re-run the same commands and diff with benchstat
before accepting any performance-sensitive change. **Supersedes the 2026-09-11
baseline** (go1.26.7, v1 mode only): the Go 1.27 toolchain changed the JSON
engine's allocation profile (see notes), so cross-toolchain comparisons for
JSON paths are not like-for-like. Figures elsewhere that cite indicative `~`
values should be re-checked against this table when next touched.

## Environment

| Factor    | Value                                                            |
| --------- | ---------------------------------------------------------------- |
| Date      | 2026-10-05                                                       |
| Toolchain | go1.27.1 linux/amd64                                             |
| CPU       | AMD RYZEN AI MAX+ 395 w/ Radeon 8060S                            |
| JSON mode | v1 (default build) AND v2 (`GOEXPERIMENT=jsonv2`) — one suite each |
| Command   | v1: `env -u GOEXPERIMENT go test -run '^$' -bench . -benchmem -count=10 -timeout 40m`; v2: same with `GOEXPERIMENT=jsonv2` |
| benchstat | `golang.org/x/perf@v0.0.0-20260929162123-406019bb8b68`           |

Notes:

- 70 benchmarks x 10 repetitions per mode (the 2026-09-11 table held 68 rows —
  its note said 67; this one adds the two
  `StreamingJSONV2_DecoderComparison` sub-benchmarks), summarized with
  `benchstat` (mean +- range across the 10 runs).
- **Quiet-machine protocol.** This machine is multi-tenant; both suites ran
  behind a sustained-idle gate (`scripts/bench-gate.sh`) with a
  `RawCodec_Encode` canary and a peak-load abort. Acceptance held with margin:
  canary 11.27n (v1) / 13.10n (v2) against a 25n threshold; peak 1-min load
  during the runs 8.10 / 9.96 against an abort threshold of 12. A first
  attempt the same night was REJECTED: uniform ~2.6-3x ns/op inflation under
  loadavg 27-58 with byte-identical B/op and allocs/op (the contention
  signature). Evidence and diagnosis: `docs/status/2026-10-05_01-49_benchmark-contention-rejection-quiet-gate.html`.
- **Go 1.27 JSON-engine allocation shifts are real, not regressions here.**
  On go1.27.1 the v1 and v2 builds produce byte-identical B/op and allocs/op
  on every row; against the go1.26.7 baseline several JSON-path allocation
  figures moved in both directions (e.g. `JSONCodec_Encode` 192B/6 allocs →
  88B/4, `JSONCodec_Decode` 592B/12 → 376B/5, `StreamingJSON_Encode`
  10.97Ki → 21.92Ki, `NormalizeForJSON` 1.182Ki → 208B). Diff against THIS
  table, not the old one.
- Cross-mode differences (v2 geomean 411.2n vs v1 385.0n with identical
  allocation profiles) are build-layout effects, not library regressions;
  always compare within one mode.
- High +- percentages on short benchmarks reflect scheduler noise on this
  machine; treat sub-microsecond differences as noise unless benchstat reports
  them significant across full runs.
- Raw outputs are not committed (fuzz-corpus precedent). The tables below plus
  the commands above fully determine regeneration; if raw output is needed for
  an audit, attach it as a CI/release artifact rather than committing it to
  the repo.

## Summary — v1 JSON mode
goos: linux
goarch: amd64
pkg: github.com/larsartmann/go-codec
cpu: AMD RYZEN AI MAX+ 395 w/ Radeon 8060S          
                                                                 │ codec-bench2-v1.txt │
                                                                 │       sec/op        │
AutoDetect/json-32                                                        72.46n ±  9%
AutoDetect/cbor-32                                                        98.54n ± 13%
AutoDetect/unknown-32                                                     113.4n ± 57%
AutoDetectDebug-32                                                        138.9n ± 16%
JSONCodec_Encode-32                                                       275.1n ±  4%
JSONCodec_Decode-32                                                       356.3n ± 12%
CBORCodec_Encode-32                                                       192.7n ±  2%
CBORCodec_Decode-32                                                       363.9n ±  8%
CodecComparison_Encode/JSON-32                                            260.5n ±  7%
CodecComparison_Encode/CBOR-32                                            207.3n ±  2%
CodecComparison_Decode/JSON-32                                            310.1n ± 18%
CodecComparison_Decode/CBOR-32                                            334.9n ±  1%
RawCodec_Encode-32                                                        11.27n ±  4%
RawCodec_Decode-32                                                        20.76n ± 16%
CBORCompact_vs_Canon_Size/Canonical/Encode-32                             111.8n ± 11%
CBORCompact_vs_Canon_Size/Compact/Encode-32                               100.2n ±  2%
CBORCompact_vs_Canon_Decode/Canonical-32                                  231.3n ±  1%
CBORCompact_vs_Canon_Decode/Compact-32                                    226.1n ±  1%
RealisticPayload_Encode/JSON-32                                           521.3n ±  2%
RealisticPayload_Encode/CBOR-32                                           273.4n ±  5%
RealisticPayload_Encode/CBOR_compact_toarray-32                           246.7n ± 10%
RealisticPayload_Decode/JSON-32                                           776.9n ±  1%
RealisticPayload_Decode/CBOR-32                                           804.0n ±  2%
RealisticPayload_Decode/CBOR_compact_toarray-32                           649.8n ±  2%
BufferEncoder/JSON-32                                                     512.1n ±  0%
BufferEncoder/CBOR-32                                                     276.8n ±  1%
BufferEncoder/CBOR_compact-32                                             276.7n ±  2%
EncodePooled/JSON-32                                                      532.9n ±  1%
EncodePooled/CBOR-32                                                      288.8n ±  1%
EncodePooled/CBOR_compact-32                                              289.4n ±  1%
TagTradeoffs_Encode/small/map-32                                          83.81n ±  0%
TagTradeoffs_Encode/small/toarray-32                                      69.32n ±  1%
TagTradeoffs_Encode/small/keyasint-32                                     80.64n ±  1%
TagTradeoffs_Encode/medium/map-32                                         286.5n ±  4%
TagTradeoffs_Encode/medium/toarray-32                                     226.2n ±  1%
TagTradeoffs_Encode/medium/keyasint-32                                    256.4n ±  1%
TagTradeoffs_Encode/large/map-32                                          224.2n ±  2%
TagTradeoffs_Encode/large/toarray-32                                      181.9n ±  2%
TagTradeoffs_Encode/large/keyasint-32                                     226.9n ±  7%
TagTradeoffs_Decode/small/map-32                                          203.8n ±  1%
TagTradeoffs_Decode/small/toarray-32                                      145.5n ±  1%
TagTradeoffs_Decode/small/keyasint-32                                     174.3n ±  1%
TagTradeoffs_Decode/medium/map-32                                         769.4n ±  3%
TagTradeoffs_Decode/medium/toarray-32                                     622.1n ±  1%
TagTradeoffs_Decode/medium/keyasint-32                                    680.5n ±  0%
TagTradeoffs_Decode/large/map-32                                          668.2n ±  1%
TagTradeoffs_Decode/large/toarray-32                                      430.1n ±  1%
TagTradeoffs_Decode/large/keyasint-32                                     588.6n ±  1%
CBORReflectionCache/encode-32                                             271.0n ±  1%
CBORReflectionCache/decode-32                                             756.9n ±  2%
ObserveCodec/encode/raw-32                                                278.1n ±  3%
ObserveCodec/encode/observed-32                                           281.1n ±  1%
ObserveCodec/decode/raw-32                                                755.6n ±  0%
ObserveCodec/decode/observed-32                                           769.4n ±  1%
ObserveCodec/encode_pooled/observed-32                                    283.9n ±  1%
WrapEncode-32                                                             297.4n ±  1%
UnwrapDecode-32                                                           244.4n ±  1%
UnwrapDecode_FallbackRawCBOR-32                                           1.472n ±  1%
StreamingJSONV2_DecoderComparison/jsontext.Decoder-32                     77.15µ ±  1%
StreamingJSONV2_DecoderComparison/json.UnmarshalRead_per_line-32          83.55µ ±  1%
NormalizeForJSON-32                                                       697.8n ±  1%
JSONCodec_MarshalUnmarshal-32                                             506.4n ±  1%
Size-32                                                                   276.4n ±  1%
StreamingJSON_Encode-32                                                   48.61µ ±  2%
StreamingJSON_Decode-32                                                   75.24µ ±  3%
StreamingCBOR_Encode-32                                                   24.91µ ±  2%
StreamingCBOR_Decode-32                                                   76.73µ ±  1%
TranscodeToJSON_CBOR_To_JSON-32                                           4.265µ ±  2%
TranscodeToJSON_JSON_Passthrough-32                                       1.810n ±  0%
TranscodeToJSON_NestedDeep-32                                             3.279µ ±  1%
geomean                                                                   385.0n

                                                                 │ codec-bench2-v1.txt │
                                                                 │        B/op         │
AutoDetect/json-32                                                        64.00 ± 0%
AutoDetect/cbor-32                                                        64.00 ± 0%
AutoDetect/unknown-32                                                     160.0 ± 0%
AutoDetectDebug-32                                                        160.0 ± 0%
JSONCodec_Encode-32                                                       88.00 ± 0%
JSONCodec_Decode-32                                                       376.0 ± 0%
CBORCodec_Encode-32                                                       96.00 ± 0%
CBORCodec_Decode-32                                                       416.0 ± 0%
CodecComparison_Encode/JSON-32                                            88.00 ± 0%
CodecComparison_Encode/CBOR-32                                            96.00 ± 0%
CodecComparison_Decode/JSON-32                                            376.0 ± 0%
CodecComparison_Decode/CBOR-32                                            416.0 ± 0%
RawCodec_Encode-32                                                        24.00 ± 0%
RawCodec_Decode-32                                                        48.00 ± 0%
CBORCompact_vs_Canon_Size/Canonical/Encode-32                             112.0 ± 0%
CBORCompact_vs_Canon_Size/Compact/Encode-32                               112.0 ± 0%
CBORCompact_vs_Canon_Decode/Canonical-32                                  77.00 ± 0%
CBORCompact_vs_Canon_Decode/Compact-32                                    77.00 ± 0%
RealisticPayload_Encode/JSON-32                                           400.0 ± 0%
RealisticPayload_Encode/CBOR-32                                           224.0 ± 0%
RealisticPayload_Encode/CBOR_compact_toarray-32                           160.0 ± 0%
RealisticPayload_Decode/JSON-32                                           208.0 ± 0%
RealisticPayload_Decode/CBOR-32                                           312.0 ± 0%
RealisticPayload_Decode/CBOR_compact_toarray-32                           312.0 ± 0%
BufferEncoder/JSON-32                                                     224.0 ± 0%
BufferEncoder/CBOR-32                                                     176.0 ± 0%
BufferEncoder/CBOR_compact-32                                             176.0 ± 0%
EncodePooled/JSON-32                                                      224.0 ± 0%
EncodePooled/CBOR-32                                                      176.0 ± 0%
EncodePooled/CBOR_compact-32                                              176.0 ± 0%
TagTradeoffs_Encode/small/map-32                                          64.00 ± 0%
TagTradeoffs_Encode/small/toarray-32                                      48.00 ± 0%
TagTradeoffs_Encode/small/keyasint-32                                     48.00 ± 0%
TagTradeoffs_Encode/medium/map-32                                         224.0 ± 0%
TagTradeoffs_Encode/medium/toarray-32                                     160.0 ± 0%
TagTradeoffs_Encode/medium/keyasint-32                                    176.0 ± 0%
TagTradeoffs_Encode/large/map-32                                          288.0 ± 0%
TagTradeoffs_Encode/large/toarray-32                                      160.0 ± 0%
TagTradeoffs_Encode/large/keyasint-32                                     176.0 ± 0%
TagTradeoffs_Decode/small/map-32                                          96.00 ± 0%
TagTradeoffs_Decode/small/toarray-32                                      96.00 ± 0%
TagTradeoffs_Decode/small/keyasint-32                                     96.00 ± 0%
TagTradeoffs_Decode/medium/map-32                                         312.0 ± 0%
TagTradeoffs_Decode/medium/toarray-32                                     312.0 ± 0%
TagTradeoffs_Decode/medium/keyasint-32                                    312.0 ± 0%
TagTradeoffs_Decode/large/map-32                                          352.0 ± 0%
TagTradeoffs_Decode/large/toarray-32                                      352.0 ± 0%
TagTradeoffs_Decode/large/keyasint-32                                     352.0 ± 0%
CBORReflectionCache/encode-32                                             336.0 ± 0%
CBORReflectionCache/decode-32                                             312.0 ± 0%
ObserveCodec/encode/raw-32                                                336.0 ± 0%
ObserveCodec/encode/observed-32                                           336.0 ± 0%
ObserveCodec/decode/raw-32                                                312.0 ± 0%
ObserveCodec/decode/observed-32                                           312.0 ± 0%
ObserveCodec/encode_pooled/observed-32                                    176.0 ± 0%
WrapEncode-32                                                             273.0 ± 0%
UnwrapDecode-32                                                           112.0 ± 0%
UnwrapDecode_FallbackRawCBOR-32                                           0.000 ± 0%
StreamingJSONV2_DecoderComparison/jsontext.Decoder-32                   33.84Ki ± 0%
StreamingJSONV2_DecoderComparison/json.UnmarshalRead_per_line-32        25.06Ki ± 0%
NormalizeForJSON-32                                                       208.0 ± 0%
JSONCodec_MarshalUnmarshal-32                                             177.0 ± 0%
Size-32                                                                   160.0 ± 0%
StreamingJSON_Encode-32                                                 21.92Ki ± 0%
StreamingJSON_Decode-32                                                 33.84Ki ± 0%
StreamingCBOR_Encode-32                                                 11.02Ki ± 0%
StreamingCBOR_Decode-32                                                 32.75Ki ± 0%
TranscodeToJSON_CBOR_To_JSON-32                                         2.436Ki ± 0%
TranscodeToJSON_JSON_Passthrough-32                                       0.000 ± 0%
TranscodeToJSON_NestedDeep-32                                           2.577Ki ± 0%
geomean                                                                              ¹
¹ summaries must be >0 to compute geomean

                                                                 │ codec-bench2-v1.txt │
                                                                 │      allocs/op      │
AutoDetect/json-32                                                        1.000 ± 0%
AutoDetect/cbor-32                                                        1.000 ± 0%
AutoDetect/unknown-32                                                     4.000 ± 0%
AutoDetectDebug-32                                                        4.000 ± 0%
JSONCodec_Encode-32                                                       4.000 ± 0%
JSONCodec_Decode-32                                                       5.000 ± 0%
CBORCodec_Encode-32                                                       2.000 ± 0%
CBORCodec_Decode-32                                                       9.000 ± 0%
CodecComparison_Encode/JSON-32                                            4.000 ± 0%
CodecComparison_Encode/CBOR-32                                            2.000 ± 0%
CodecComparison_Decode/JSON-32                                            5.000 ± 0%
CodecComparison_Decode/CBOR-32                                            9.000 ± 0%
RawCodec_Encode-32                                                        1.000 ± 0%
RawCodec_Decode-32                                                        2.000 ± 0%
CBORCompact_vs_Canon_Size/Canonical/Encode-32                             2.000 ± 0%
CBORCompact_vs_Canon_Size/Compact/Encode-32                               2.000 ± 0%
CBORCompact_vs_Canon_Decode/Canonical-32                                  3.000 ± 0%
CBORCompact_vs_Canon_Decode/Compact-32                                    3.000 ± 0%
RealisticPayload_Encode/JSON-32                                           2.000 ± 0%
RealisticPayload_Encode/CBOR-32                                           1.000 ± 0%
RealisticPayload_Encode/CBOR_compact_toarray-32                           1.000 ± 0%
RealisticPayload_Decode/JSON-32                                           3.000 ± 0%
RealisticPayload_Decode/CBOR-32                                           9.000 ± 0%
RealisticPayload_Decode/CBOR_compact_toarray-32                           9.000 ± 0%
BufferEncoder/JSON-32                                                     2.000 ± 0%
BufferEncoder/CBOR-32                                                     2.000 ± 0%
BufferEncoder/CBOR_compact-32                                             2.000 ± 0%
EncodePooled/JSON-32                                                      2.000 ± 0%
EncodePooled/CBOR-32                                                      2.000 ± 0%
EncodePooled/CBOR_compact-32                                              2.000 ± 0%
TagTradeoffs_Encode/small/map-32                                          1.000 ± 0%
TagTradeoffs_Encode/small/toarray-32                                      1.000 ± 0%
TagTradeoffs_Encode/small/keyasint-32                                     1.000 ± 0%
TagTradeoffs_Encode/medium/map-32                                         1.000 ± 0%
TagTradeoffs_Encode/medium/toarray-32                                     1.000 ± 0%
TagTradeoffs_Encode/medium/keyasint-32                                    1.000 ± 0%
TagTradeoffs_Encode/large/map-32                                          1.000 ± 0%
TagTradeoffs_Encode/large/toarray-32                                      1.000 ± 0%
TagTradeoffs_Encode/large/keyasint-32                                     1.000 ± 0%
TagTradeoffs_Decode/small/map-32                                          4.000 ± 0%
TagTradeoffs_Decode/small/toarray-32                                      4.000 ± 0%
TagTradeoffs_Decode/small/keyasint-32                                     4.000 ± 0%
TagTradeoffs_Decode/medium/map-32                                         9.000 ± 0%
TagTradeoffs_Decode/medium/toarray-32                                     9.000 ± 0%
TagTradeoffs_Decode/medium/keyasint-32                                    9.000 ± 0%
TagTradeoffs_Decode/large/map-32                                          10.00 ± 0%
TagTradeoffs_Decode/large/toarray-32                                      10.00 ± 0%
TagTradeoffs_Decode/large/keyasint-32                                     10.00 ± 0%
CBORReflectionCache/encode-32                                             2.000 ± 0%
CBORReflectionCache/decode-32                                             9.000 ± 0%
ObserveCodec/encode/raw-32                                                2.000 ± 0%
ObserveCodec/encode/observed-32                                           2.000 ± 0%
ObserveCodec/decode/raw-32                                                9.000 ± 0%
ObserveCodec/decode/observed-32                                           9.000 ± 0%
ObserveCodec/encode_pooled/observed-32                                    2.000 ± 0%
WrapEncode-32                                                             4.000 ± 0%
UnwrapDecode-32                                                           2.000 ± 0%
UnwrapDecode_FallbackRawCBOR-32                                           0.000 ± 0%
StreamingJSONV2_DecoderComparison/jsontext.Decoder-32                     325.0 ± 0%
StreamingJSONV2_DecoderComparison/json.UnmarshalRead_per_line-32          400.0 ± 0%
NormalizeForJSON-32                                                       9.000 ± 0%
JSONCodec_MarshalUnmarshal-32                                             3.000 ± 0%
Size-32                                                                   3.000 ± 0%
StreamingJSON_Encode-32                                                   200.0 ± 0%
StreamingJSON_Decode-32                                                   325.0 ± 0%
StreamingCBOR_Encode-32                                                   101.0 ± 0%
StreamingCBOR_Decode-32                                                   905.0 ± 0%
TranscodeToJSON_CBOR_To_JSON-32                                           93.00 ± 0%
TranscodeToJSON_JSON_Passthrough-32                                       0.000 ± 0%
TranscodeToJSON_NestedDeep-32                                             71.00 ± 0%
geomean                                                                              ¹
¹ summaries must be >0 to compute geomean

## Summary — v2 JSON mode
goos: linux
goarch: amd64
pkg: github.com/larsartmann/go-codec
cpu: AMD RYZEN AI MAX+ 395 w/ Radeon 8060S          
                                                                 │ codec-bench2-v2.txt │
                                                                 │       sec/op        │
AutoDetect/json-32                                                        64.59n ±  3%
AutoDetect/cbor-32                                                        91.17n ±  1%
AutoDetect/unknown-32                                                     118.6n ±  1%
AutoDetectDebug-32                                                        119.2n ±  2%
JSONCodec_Encode-32                                                       271.8n ±  2%
JSONCodec_Decode-32                                                       320.9n ±  1%
CBORCodec_Encode-32                                                       193.0n ±  3%
CBORCodec_Decode-32                                                       362.4n ±  1%
CodecComparison_Encode/JSON-32                                            271.1n ±  3%
CodecComparison_Encode/CBOR-32                                            202.6n ± 10%
CodecComparison_Decode/JSON-32                                            342.5n ±  4%
CodecComparison_Decode/CBOR-32                                            365.4n ± 11%
RawCodec_Encode-32                                                        13.10n ± 12%
RawCodec_Decode-32                                                        19.61n ±  7%
CBORCompact_vs_Canon_Size/Canonical/Encode-32                             116.2n ± 13%
CBORCompact_vs_Canon_Size/Compact/Encode-32                               117.8n ± 16%
CBORCompact_vs_Canon_Decode/Canonical-32                                  246.7n ± 12%
CBORCompact_vs_Canon_Decode/Compact-32                                    252.1n ±  9%
RealisticPayload_Encode/JSON-32                                           617.6n ±  5%
RealisticPayload_Encode/CBOR-32                                           267.6n ± 23%
RealisticPayload_Encode/CBOR_compact_toarray-32                           268.1n ± 10%
RealisticPayload_Decode/JSON-32                                           751.9n ±  2%
RealisticPayload_Decode/CBOR-32                                           901.2n ± 12%
RealisticPayload_Decode/CBOR_compact_toarray-32                           764.7n ±  8%
BufferEncoder/JSON-32                                                     507.6n ±  5%
BufferEncoder/CBOR-32                                                     277.7n ±  3%
BufferEncoder/CBOR_compact-32                                             285.2n ±  7%
EncodePooled/JSON-32                                                      665.6n ±  6%
EncodePooled/CBOR-32                                                      338.2n ±  5%
EncodePooled/CBOR_compact-32                                              316.4n ± 14%
TagTradeoffs_Encode/small/map-32                                          81.45n ±  2%
TagTradeoffs_Encode/small/toarray-32                                      67.68n ±  2%
TagTradeoffs_Encode/small/keyasint-32                                     79.65n ±  5%
TagTradeoffs_Encode/medium/map-32                                         260.4n ±  2%
TagTradeoffs_Encode/medium/toarray-32                                     284.9n ±  7%
TagTradeoffs_Encode/medium/keyasint-32                                    314.9n ±  9%
TagTradeoffs_Encode/large/map-32                                          260.5n ± 14%
TagTradeoffs_Encode/large/toarray-32                                      178.4n ±  3%
TagTradeoffs_Encode/large/keyasint-32                                     237.8n ±  7%
TagTradeoffs_Decode/small/map-32                                          211.9n ±  4%
TagTradeoffs_Decode/small/toarray-32                                      157.0n ±  4%
TagTradeoffs_Decode/small/keyasint-32                                     178.3n ± 16%
TagTradeoffs_Decode/medium/map-32                                         994.5n ± 13%
TagTradeoffs_Decode/medium/toarray-32                                     752.7n ±  3%
TagTradeoffs_Decode/medium/keyasint-32                                    825.5n ±  5%
TagTradeoffs_Decode/large/map-32                                          718.2n ± 12%
TagTradeoffs_Decode/large/toarray-32                                      457.1n ± 10%
TagTradeoffs_Decode/large/keyasint-32                                     615.0n ±  3%
CBORReflectionCache/encode-32                                             286.2n ±  5%
CBORReflectionCache/decode-32                                             802.5n ±  5%
ObserveCodec/encode/raw-32                                                347.6n ±  6%
ObserveCodec/encode/observed-32                                           363.5n ±  6%
ObserveCodec/decode/raw-32                                                905.2n ±  4%
ObserveCodec/decode/observed-32                                           815.0n ±  4%
ObserveCodec/encode_pooled/observed-32                                    315.5n ±  7%
WrapEncode-32                                                             329.2n ±  4%
UnwrapDecode-32                                                           258.8n ±  9%
UnwrapDecode_FallbackRawCBOR-32                                           1.619n ±  4%
StreamingJSONV2_DecoderComparison/jsontext.Decoder-32                     93.20µ ±  4%
StreamingJSONV2_DecoderComparison/json.UnmarshalRead_per_line-32          97.57µ ±  2%
NormalizeForJSON-32                                                       816.7n ±  8%
JSONCodec_MarshalUnmarshal-32                                             547.0n ±  7%
Size-32                                                                   282.4n ±  1%
StreamingJSON_Encode-32                                                   49.34µ ±  1%
StreamingJSON_Decode-32                                                   76.83µ ±  3%
StreamingCBOR_Encode-32                                                   25.58µ ±  2%
StreamingCBOR_Decode-32                                                   78.30µ ±  1%
TranscodeToJSON_CBOR_To_JSON-32                                           4.526µ ± 16%
TranscodeToJSON_JSON_Passthrough-32                                       1.841n ±  1%
TranscodeToJSON_NestedDeep-32                                             3.375µ ± 11%
geomean                                                                   411.2n

                                                                 │ codec-bench2-v2.txt │
                                                                 │        B/op         │
AutoDetect/json-32                                                        64.00 ± 0%
AutoDetect/cbor-32                                                        64.00 ± 0%
AutoDetect/unknown-32                                                     160.0 ± 0%
AutoDetectDebug-32                                                        160.0 ± 0%
JSONCodec_Encode-32                                                       88.00 ± 0%
JSONCodec_Decode-32                                                       376.0 ± 0%
CBORCodec_Encode-32                                                       96.00 ± 0%
CBORCodec_Decode-32                                                       416.0 ± 0%
CodecComparison_Encode/JSON-32                                            88.00 ± 0%
CodecComparison_Encode/CBOR-32                                            96.00 ± 0%
CodecComparison_Decode/JSON-32                                            376.0 ± 0%
CodecComparison_Decode/CBOR-32                                            416.0 ± 0%
RawCodec_Encode-32                                                        24.00 ± 0%
RawCodec_Decode-32                                                        48.00 ± 0%
CBORCompact_vs_Canon_Size/Canonical/Encode-32                             112.0 ± 0%
CBORCompact_vs_Canon_Size/Compact/Encode-32                               112.0 ± 0%
CBORCompact_vs_Canon_Decode/Canonical-32                                  77.00 ± 0%
CBORCompact_vs_Canon_Decode/Compact-32                                    77.00 ± 0%
RealisticPayload_Encode/JSON-32                                           400.0 ± 0%
RealisticPayload_Encode/CBOR-32                                           224.0 ± 0%
RealisticPayload_Encode/CBOR_compact_toarray-32                           160.0 ± 0%
RealisticPayload_Decode/JSON-32                                           208.0 ± 0%
RealisticPayload_Decode/CBOR-32                                           312.0 ± 0%
RealisticPayload_Decode/CBOR_compact_toarray-32                           312.0 ± 0%
BufferEncoder/JSON-32                                                     224.0 ± 0%
BufferEncoder/CBOR-32                                                     176.0 ± 0%
BufferEncoder/CBOR_compact-32                                             176.0 ± 0%
EncodePooled/JSON-32                                                      224.0 ± 0%
EncodePooled/CBOR-32                                                      176.0 ± 0%
EncodePooled/CBOR_compact-32                                              176.0 ± 0%
TagTradeoffs_Encode/small/map-32                                          64.00 ± 0%
TagTradeoffs_Encode/small/toarray-32                                      48.00 ± 0%
TagTradeoffs_Encode/small/keyasint-32                                     48.00 ± 0%
TagTradeoffs_Encode/medium/map-32                                         224.0 ± 0%
TagTradeoffs_Encode/medium/toarray-32                                     160.0 ± 0%
TagTradeoffs_Encode/medium/keyasint-32                                    176.0 ± 0%
TagTradeoffs_Encode/large/map-32                                          288.0 ± 0%
TagTradeoffs_Encode/large/toarray-32                                      160.0 ± 0%
TagTradeoffs_Encode/large/keyasint-32                                     176.0 ± 0%
TagTradeoffs_Decode/small/map-32                                          96.00 ± 0%
TagTradeoffs_Decode/small/toarray-32                                      96.00 ± 0%
TagTradeoffs_Decode/small/keyasint-32                                     96.00 ± 0%
TagTradeoffs_Decode/medium/map-32                                         312.0 ± 0%
TagTradeoffs_Decode/medium/toarray-32                                     312.0 ± 0%
TagTradeoffs_Decode/medium/keyasint-32                                    312.0 ± 0%
TagTradeoffs_Decode/large/map-32                                          352.0 ± 0%
TagTradeoffs_Decode/large/toarray-32                                      352.0 ± 0%
TagTradeoffs_Decode/large/keyasint-32                                     352.0 ± 0%
CBORReflectionCache/encode-32                                             336.0 ± 0%
CBORReflectionCache/decode-32                                             312.0 ± 0%
ObserveCodec/encode/raw-32                                                336.0 ± 0%
ObserveCodec/encode/observed-32                                           336.0 ± 0%
ObserveCodec/decode/raw-32                                                312.0 ± 0%
ObserveCodec/decode/observed-32                                           312.0 ± 0%
ObserveCodec/encode_pooled/observed-32                                    176.0 ± 0%
WrapEncode-32                                                             273.0 ± 0%
UnwrapDecode-32                                                           112.0 ± 0%
UnwrapDecode_FallbackRawCBOR-32                                           0.000 ± 0%
StreamingJSONV2_DecoderComparison/jsontext.Decoder-32                   33.84Ki ± 0%
StreamingJSONV2_DecoderComparison/json.UnmarshalRead_per_line-32        25.06Ki ± 0%
NormalizeForJSON-32                                                       208.0 ± 0%
JSONCodec_MarshalUnmarshal-32                                             177.0 ± 0%
Size-32                                                                   160.0 ± 0%
StreamingJSON_Encode-32                                                 21.92Ki ± 0%
StreamingJSON_Decode-32                                                 33.84Ki ± 0%
StreamingCBOR_Encode-32                                                 11.02Ki ± 0%
StreamingCBOR_Decode-32                                                 32.75Ki ± 0%
TranscodeToJSON_CBOR_To_JSON-32                                         2.436Ki ± 0%
TranscodeToJSON_JSON_Passthrough-32                                       0.000 ± 0%
TranscodeToJSON_NestedDeep-32                                           2.577Ki ± 0%
geomean                                                                              ¹
¹ summaries must be >0 to compute geomean

                                                                 │ codec-bench2-v2.txt │
                                                                 │      allocs/op      │
AutoDetect/json-32                                                        1.000 ± 0%
AutoDetect/cbor-32                                                        1.000 ± 0%
AutoDetect/unknown-32                                                     4.000 ± 0%
AutoDetectDebug-32                                                        4.000 ± 0%
JSONCodec_Encode-32                                                       4.000 ± 0%
JSONCodec_Decode-32                                                       5.000 ± 0%
CBORCodec_Encode-32                                                       2.000 ± 0%
CBORCodec_Decode-32                                                       9.000 ± 0%
CodecComparison_Encode/JSON-32                                            4.000 ± 0%
CodecComparison_Encode/CBOR-32                                            2.000 ± 0%
CodecComparison_Decode/JSON-32                                            5.000 ± 0%
CodecComparison_Decode/CBOR-32                                            9.000 ± 0%
RawCodec_Encode-32                                                        1.000 ± 0%
RawCodec_Decode-32                                                        2.000 ± 0%
CBORCompact_vs_Canon_Size/Canonical/Encode-32                             2.000 ± 0%
CBORCompact_vs_Canon_Size/Compact/Encode-32                               2.000 ± 0%
CBORCompact_vs_Canon_Decode/Canonical-32                                  3.000 ± 0%
CBORCompact_vs_Canon_Decode/Compact-32                                    3.000 ± 0%
RealisticPayload_Encode/JSON-32                                           2.000 ± 0%
RealisticPayload_Encode/CBOR-32                                           1.000 ± 0%
RealisticPayload_Encode/CBOR_compact_toarray-32                           1.000 ± 0%
RealisticPayload_Decode/JSON-32                                           3.000 ± 0%
RealisticPayload_Decode/CBOR-32                                           9.000 ± 0%
RealisticPayload_Decode/CBOR_compact_toarray-32                           9.000 ± 0%
BufferEncoder/JSON-32                                                     2.000 ± 0%
BufferEncoder/CBOR-32                                                     2.000 ± 0%
BufferEncoder/CBOR_compact-32                                             2.000 ± 0%
EncodePooled/JSON-32                                                      2.000 ± 0%
EncodePooled/CBOR-32                                                      2.000 ± 0%
EncodePooled/CBOR_compact-32                                              2.000 ± 0%
TagTradeoffs_Encode/small/map-32                                          1.000 ± 0%
TagTradeoffs_Encode/small/toarray-32                                      1.000 ± 0%
TagTradeoffs_Encode/small/keyasint-32                                     1.000 ± 0%
TagTradeoffs_Encode/medium/map-32                                         1.000 ± 0%
TagTradeoffs_Encode/medium/toarray-32                                     1.000 ± 0%
TagTradeoffs_Encode/medium/keyasint-32                                    1.000 ± 0%
TagTradeoffs_Encode/large/map-32                                          1.000 ± 0%
TagTradeoffs_Encode/large/toarray-32                                      1.000 ± 0%
TagTradeoffs_Encode/large/keyasint-32                                     1.000 ± 0%
TagTradeoffs_Decode/small/map-32                                          4.000 ± 0%
TagTradeoffs_Decode/small/toarray-32                                      4.000 ± 0%
TagTradeoffs_Decode/small/keyasint-32                                     4.000 ± 0%
TagTradeoffs_Decode/medium/map-32                                         9.000 ± 0%
TagTradeoffs_Decode/medium/toarray-32                                     9.000 ± 0%
TagTradeoffs_Decode/medium/keyasint-32                                    9.000 ± 0%
TagTradeoffs_Decode/large/map-32                                          10.00 ± 0%
TagTradeoffs_Decode/large/toarray-32                                      10.00 ± 0%
TagTradeoffs_Decode/large/keyasint-32                                     10.00 ± 0%
CBORReflectionCache/encode-32                                             2.000 ± 0%
CBORReflectionCache/decode-32                                             9.000 ± 0%
ObserveCodec/encode/raw-32                                                2.000 ± 0%
ObserveCodec/encode/observed-32                                           2.000 ± 0%
ObserveCodec/decode/raw-32                                                9.000 ± 0%
ObserveCodec/decode/observed-32                                           9.000 ± 0%
ObserveCodec/encode_pooled/observed-32                                    2.000 ± 0%
WrapEncode-32                                                             4.000 ± 0%
UnwrapDecode-32                                                           2.000 ± 0%
UnwrapDecode_FallbackRawCBOR-32                                           0.000 ± 0%
StreamingJSONV2_DecoderComparison/jsontext.Decoder-32                     325.0 ± 0%
StreamingJSONV2_DecoderComparison/json.UnmarshalRead_per_line-32          400.0 ± 0%
NormalizeForJSON-32                                                       9.000 ± 0%
JSONCodec_MarshalUnmarshal-32                                             3.000 ± 0%
Size-32                                                                   3.000 ± 0%
StreamingJSON_Encode-32                                                   200.0 ± 0%
StreamingJSON_Decode-32                                                   325.0 ± 0%
StreamingCBOR_Encode-32                                                   101.0 ± 0%
StreamingCBOR_Decode-32                                                   905.0 ± 0%
TranscodeToJSON_CBOR_To_JSON-32                                           93.00 ± 0%
TranscodeToJSON_JSON_Passthrough-32                                       0.000 ± 0%
TranscodeToJSON_NestedDeep-32                                             71.00 ± 0%
geomean                                                                              ¹
¹ summaries must be >0 to compute geomean
