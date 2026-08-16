# Benchmark Baseline (v1 JSON mode)

Reference performance baseline for regression comparison, recorded 2026-08-15.
Figures in `FEATURES.md` that previously carried indicative `~` values cite this
file. Re-run the same command and diff with benchstat before accepting any
performance-sensitive change.

## Environment

| Factor    | Value                                                         |
| --------- | ------------------------------------------------------------- |
| Date      | 2026-08-15                                                    |
| Toolchain | go1.26.5 linux/amd64                                          |
| CPU       | AMD RYZEN AI MAX+ 395 w/ Radeon 8060S                         |
| JSON mode | v1 (default build; `GOEXPERIMENT` unset)                      |
| Command   | `go test -run '^$' -bench . -benchmem -count=10 -timeout 40m` |

Notes:

- 67 sub-benchmarks x 10 repetitions, summarized with `benchstat` (mean +- range
  across the 10 runs).
- High +- percentages on short benchmarks (e.g. `AutoDetect/json` +-47%) reflect
  scheduler noise on this machine; treat sub-microsecond differences as noise
  unless benchstat reports them significant across full runs.
- The raw output is not committed; the summary below plus the reproduction
  command above fully determine it.

## Summary

```text
goos: linux
goarch: amd64
pkg: github.com/larsartmann/go-codec
cpu: AMD RYZEN AI MAX+ 395 w/ Radeon 8060S          
                                                │ /tmp/codec-baseline-v1.txt │
                                                │           sec/op           │
AutoDetect/json-32                                              98.48n ± 47%
AutoDetect/cbor-32                                              81.21n ±  9%
AutoDetect/unknown-32                                           135.1n ±  1%
AutoDetectDebug-32                                              140.2n ± 13%
JSONCodec_Encode-32                                             306.6n ± 18%
JSONCodec_Decode-32                                             1.602µ ± 51%
CBORCodec_Encode-32                                             376.2n ± 30%
CBORCodec_Decode-32                                             379.9n ± 12%
CodecComparison_Encode/JSON-32                                  251.2n ±  4%
CodecComparison_Encode/CBOR-32                                  222.5n ± 12%
CodecComparison_Decode/JSON-32                                  647.6n ±  7%
CodecComparison_Decode/CBOR-32                                  703.3n ± 35%
RawCodec_Encode-32                                              20.17n ± 12%
RawCodec_Decode-32                                              51.46n ± 32%
CBORCompact_vs_Canon_Size/Canonical/Encode-32                   129.1n ± 13%
CBORCompact_vs_Canon_Size/Compact/Encode-32                     127.8n ±  6%
CBORCompact_vs_Canon_Decode/Canonical-32                        265.3n ± 12%
CBORCompact_vs_Canon_Decode/Compact-32                          261.4n ± 10%
RealisticPayload_Encode/JSON-32                                 859.6n ± 46%
RealisticPayload_Encode/CBOR-32                                 512.8n ± 41%
RealisticPayload_Encode/CBOR_compact_toarray-32                 280.2n ±  5%
RealisticPayload_Decode/JSON-32                                 2.226µ ± 16%
RealisticPayload_Decode/CBOR-32                                 1.051µ ± 33%
RealisticPayload_Decode/CBOR_compact_toarray-32                 670.8n ± 14%
BufferEncoder/JSON-32                                           602.3n ± 33%
BufferEncoder/CBOR-32                                           433.2n ± 26%
BufferEncoder/CBOR_compact-32                                   531.5n ±  8%
EncodePooled/JSON-32                                            577.1n ± 65%
EncodePooled/CBOR-32                                            327.3n ± 15%
EncodePooled/CBOR_compact-32                                    323.0n ±  2%
TagTradeoffs_Encode/small/map-32                                103.4n ± 64%
TagTradeoffs_Encode/small/toarray-32                            151.3n ± 43%
TagTradeoffs_Encode/small/keyasint-32                           159.3n ± 42%
TagTradeoffs_Encode/medium/map-32                               557.2n ± 17%
TagTradeoffs_Encode/medium/toarray-32                           594.7n ± 13%
TagTradeoffs_Encode/medium/keyasint-32                          517.5n ± 26%
TagTradeoffs_Encode/large/map-32                                685.2n ± 18%
TagTradeoffs_Encode/large/toarray-32                            437.7n ± 16%
TagTradeoffs_Encode/large/keyasint-32                           481.5n ± 34%
TagTradeoffs_Decode/small/map-32                                576.3n ± 12%
TagTradeoffs_Decode/small/toarray-32                            265.1n ± 55%
TagTradeoffs_Decode/small/keyasint-32                           206.8n ± 32%
TagTradeoffs_Decode/medium/map-32                               2.165µ ± 56%
TagTradeoffs_Decode/medium/toarray-32                           1.864µ ±  5%
TagTradeoffs_Decode/medium/keyasint-32                          1.402µ ± 37%
TagTradeoffs_Decode/large/map-32                                1.664µ ± 34%
TagTradeoffs_Decode/large/toarray-32                            1.286µ ±  9%
TagTradeoffs_Decode/large/keyasint-32                           937.2n ± 31%
CBORReflectionCache/encode-32                                   331.9n ±  6%
CBORReflectionCache/decode-32                                   831.0n ±  3%
ObserveCodec/encode/raw-32                                      325.1n ±  5%
ObserveCodec/encode/observed-32                                 337.1n ± 12%
ObserveCodec/decode/raw-32                                      811.1n ±  3%
ObserveCodec/decode/observed-32                                 859.5n ± 12%
ObserveCodec/encode_pooled/observed-32                          366.2n ± 33%
WrapEncode-32                                                   275.2n ± 12%
UnwrapDecode-32                                                 772.2n ± 14%
NormalizeForJSON-32                                             1.555µ ± 22%
JSONCodec_MarshalUnmarshal-32                                   1.187µ ± 27%
Size-32                                                         370.0n ± 41%
StreamingJSON_Encode-32                                         82.47µ ±  4%
StreamingJSON_Decode-32                                         419.6µ ± 33%
StreamingCBOR_Encode-32                                         38.03µ ± 22%
StreamingCBOR_Decode-32                                         128.7µ ± 31%
TranscodeToJSON_CBOR_To_JSON-32                                 10.94µ ± 33%
TranscodeToJSON_JSON_Passthrough-32                             2.569n ± 10%
TranscodeToJSON_NestedDeep-32                                   9.998µ ± 33%
geomean                                                         593.0n

                                                │ /tmp/codec-baseline-v1.txt │
                                                │            B/op            │
AutoDetect/json-32                                              64.00 ± 0%
AutoDetect/cbor-32                                              64.00 ± 0%
AutoDetect/unknown-32                                           160.0 ± 0%
AutoDetectDebug-32                                              160.0 ± 0%
JSONCodec_Encode-32                                             192.0 ± 0%
JSONCodec_Decode-32                                             592.0 ± 0%
CBORCodec_Encode-32                                             96.00 ± 0%
CBORCodec_Decode-32                                             416.0 ± 0%
CodecComparison_Encode/JSON-32                                  192.0 ± 0%
CodecComparison_Encode/CBOR-32                                  96.00 ± 0%
CodecComparison_Decode/JSON-32                                  592.0 ± 0%
CodecComparison_Decode/CBOR-32                                  416.0 ± 0%
RawCodec_Encode-32                                              24.00 ± 0%
RawCodec_Decode-32                                              48.00 ± 0%
CBORCompact_vs_Canon_Size/Canonical/Encode-32                   112.0 ± 0%
CBORCompact_vs_Canon_Size/Compact/Encode-32                     112.0 ± 0%
CBORCompact_vs_Canon_Decode/Canonical-32                        77.00 ± 0%
CBORCompact_vs_Canon_Decode/Compact-32                          77.00 ± 0%
RealisticPayload_Encode/JSON-32                                 288.0 ± 0%
RealisticPayload_Encode/CBOR-32                                 224.0 ± 0%
RealisticPayload_Encode/CBOR_compact_toarray-32                 160.0 ± 0%
RealisticPayload_Decode/JSON-32                                 608.0 ± 0%
RealisticPayload_Decode/CBOR-32                                 312.0 ± 0%
RealisticPayload_Decode/CBOR_compact_toarray-32                 312.0 ± 0%
BufferEncoder/JSON-32                                           400.0 ± 0%
BufferEncoder/CBOR-32                                           176.0 ± 0%
BufferEncoder/CBOR_compact-32                                   176.0 ± 0%
EncodePooled/JSON-32                                            401.0 ± 0%
EncodePooled/CBOR-32                                            176.0 ± 0%
EncodePooled/CBOR_compact-32                                    176.0 ± 0%
TagTradeoffs_Encode/small/map-32                                64.00 ± 0%
TagTradeoffs_Encode/small/toarray-32                            48.00 ± 0%
TagTradeoffs_Encode/small/keyasint-32                           48.00 ± 0%
TagTradeoffs_Encode/medium/map-32                               224.0 ± 0%
TagTradeoffs_Encode/medium/toarray-32                           160.0 ± 0%
TagTradeoffs_Encode/medium/keyasint-32                          176.0 ± 0%
TagTradeoffs_Encode/large/map-32                                288.0 ± 0%
TagTradeoffs_Encode/large/toarray-32                            160.0 ± 0%
TagTradeoffs_Encode/large/keyasint-32                           176.0 ± 0%
TagTradeoffs_Decode/small/map-32                                96.00 ± 0%
TagTradeoffs_Decode/small/toarray-32                            96.00 ± 0%
TagTradeoffs_Decode/small/keyasint-32                           96.00 ± 0%
TagTradeoffs_Decode/medium/map-32                               312.0 ± 0%
TagTradeoffs_Decode/medium/toarray-32                           312.0 ± 0%
TagTradeoffs_Decode/medium/keyasint-32                          312.0 ± 0%
TagTradeoffs_Decode/large/map-32                                352.0 ± 0%
TagTradeoffs_Decode/large/toarray-32                            352.0 ± 0%
TagTradeoffs_Decode/large/keyasint-32                           352.0 ± 0%
CBORReflectionCache/encode-32                                   336.0 ± 0%
CBORReflectionCache/decode-32                                   312.0 ± 0%
ObserveCodec/encode/raw-32                                      336.0 ± 0%
ObserveCodec/encode/observed-32                                 336.0 ± 0%
ObserveCodec/decode/raw-32                                      312.0 ± 0%
ObserveCodec/decode/observed-32                                 312.0 ± 0%
ObserveCodec/encode_pooled/observed-32                          176.0 ± 0%
WrapEncode-32                                                   208.0 ± 0%
UnwrapDecode-32                                                 336.0 ± 0%
NormalizeForJSON-32                                           1.182Ki ± 0%
JSONCodec_MarshalUnmarshal-32                                   376.0 ± 0%
Size-32                                                         112.0 ± 0%
StreamingJSON_Encode-32                                       10.97Ki ± 0%
StreamingJSON_Decode-32                                       33.92Ki ± 0%
StreamingCBOR_Encode-32                                       11.03Ki ± 0%
StreamingCBOR_Decode-32                                       32.75Ki ± 0%
TranscodeToJSON_CBOR_To_JSON-32                               4.021Ki ± 0%
TranscodeToJSON_JSON_Passthrough-32                             0.000 ± 0%
TranscodeToJSON_NestedDeep-32                                 4.500Ki ± 0%
geomean                                                                    ¹
¹ summaries must be >0 to compute geomean

                                                │ /tmp/codec-baseline-v1.txt │
                                                │         allocs/op          │
AutoDetect/json-32                                              1.000 ± 0%
AutoDetect/cbor-32                                              1.000 ± 0%
AutoDetect/unknown-32                                           4.000 ± 0%
AutoDetectDebug-32                                              4.000 ± 0%
JSONCodec_Encode-32                                             6.000 ± 0%
JSONCodec_Decode-32                                             12.00 ± 0%
CBORCodec_Encode-32                                             2.000 ± 0%
CBORCodec_Decode-32                                             9.000 ± 0%
CodecComparison_Encode/JSON-32                                  6.000 ± 0%
CodecComparison_Encode/CBOR-32                                  2.000 ± 0%
CodecComparison_Decode/JSON-32                                  12.00 ± 0%
CodecComparison_Decode/CBOR-32                                  9.000 ± 0%
RawCodec_Encode-32                                              1.000 ± 0%
RawCodec_Decode-32                                              2.000 ± 0%
CBORCompact_vs_Canon_Size/Canonical/Encode-32                   2.000 ± 0%
CBORCompact_vs_Canon_Size/Compact/Encode-32                     2.000 ± 0%
CBORCompact_vs_Canon_Decode/Canonical-32                        3.000 ± 0%
CBORCompact_vs_Canon_Decode/Compact-32                          3.000 ± 0%
RealisticPayload_Encode/JSON-32                                 1.000 ± 0%
RealisticPayload_Encode/CBOR-32                                 1.000 ± 0%
RealisticPayload_Encode/CBOR_compact_toarray-32                 1.000 ± 0%
RealisticPayload_Decode/JSON-32                                 16.00 ± 0%
RealisticPayload_Decode/CBOR-32                                 9.000 ± 0%
RealisticPayload_Decode/CBOR_compact_toarray-32                 9.000 ± 0%
BufferEncoder/JSON-32                                           2.000 ± 0%
BufferEncoder/CBOR-32                                           2.000 ± 0%
BufferEncoder/CBOR_compact-32                                   2.000 ± 0%
EncodePooled/JSON-32                                            2.000 ± 0%
EncodePooled/CBOR-32                                            2.000 ± 0%
EncodePooled/CBOR_compact-32                                    2.000 ± 0%
TagTradeoffs_Encode/small/map-32                                1.000 ± 0%
TagTradeoffs_Encode/small/toarray-32                            1.000 ± 0%
TagTradeoffs_Encode/small/keyasint-32                           1.000 ± 0%
TagTradeoffs_Encode/medium/map-32                               1.000 ± 0%
TagTradeoffs_Encode/medium/toarray-32                           1.000 ± 0%
TagTradeoffs_Encode/medium/keyasint-32                          1.000 ± 0%
TagTradeoffs_Encode/large/map-32                                1.000 ± 0%
TagTradeoffs_Encode/large/toarray-32                            1.000 ± 0%
TagTradeoffs_Encode/large/keyasint-32                           1.000 ± 0%
TagTradeoffs_Decode/small/map-32                                4.000 ± 0%
TagTradeoffs_Decode/small/toarray-32                            4.000 ± 0%
TagTradeoffs_Decode/small/keyasint-32                           4.000 ± 0%
TagTradeoffs_Decode/medium/map-32                               9.000 ± 0%
TagTradeoffs_Decode/medium/toarray-32                           9.000 ± 0%
TagTradeoffs_Decode/medium/keyasint-32                          9.000 ± 0%
TagTradeoffs_Decode/large/map-32                                10.00 ± 0%
TagTradeoffs_Decode/large/toarray-32                            10.00 ± 0%
TagTradeoffs_Decode/large/keyasint-32                           10.00 ± 0%
CBORReflectionCache/encode-32                                   2.000 ± 0%
CBORReflectionCache/decode-32                                   9.000 ± 0%
ObserveCodec/encode/raw-32                                      2.000 ± 0%
ObserveCodec/encode/observed-32                                 2.000 ± 0%
ObserveCodec/decode/raw-32                                      9.000 ± 0%
ObserveCodec/decode/observed-32                                 9.000 ± 0%
ObserveCodec/encode_pooled/observed-32                          2.000 ± 0%
WrapEncode-32                                                   3.000 ± 0%
UnwrapDecode-32                                                 8.000 ± 0%
NormalizeForJSON-32                                             19.00 ± 0%
JSONCodec_MarshalUnmarshal-32                                   8.000 ± 0%
Size-32                                                         2.000 ± 0%
StreamingJSON_Encode-32                                         100.0 ± 0%
StreamingJSON_Decode-32                                         914.0 ± 0%
StreamingCBOR_Encode-32                                         101.0 ± 0%
StreamingCBOR_Decode-32                                         905.0 ± 0%
TranscodeToJSON_CBOR_To_JSON-32                                 102.0 ± 0%
TranscodeToJSON_JSON_Passthrough-32                             0.000 ± 0%
TranscodeToJSON_NestedDeep-32                                   78.00 ± 0%
geomean                                                                    ¹
¹ summaries must be >0 to compute geomean
```

## Addendum 2026-08-16 — `UnwrapDecode` first-byte sniff

`UnwrapDecode` now skips the envelope JSON parse when the first byte is ≥ 0x80
(CBOR major types 4-7 can never begin valid JSON). Focused 10-count benchstat
(same machine/toolchain, v1 mode, `go test -run '^$' -bench 'UnwrapDecode'
-benchmem -count=10`) before → after:

```text
                                                 │  before   │              after              │
                                                 │  sec/op   │    sec/op     vs base           │
UnwrapDecode-32                                 296.0n ± 6%   296.9n ± 3%        ~ (p=0.912)
UnwrapDecode_FallbackRawCBOR-32                180.750n ± 1%   1.633n ± 9%  -99.10% (p=0.000)

UnwrapDecode_FallbackRawCBOR-32  B/op    184.0 ± 0%     0.0 ± 0%  -100.00%
UnwrapDecode_FallbackRawCBOR-32  allocs  6.000 ± 0%    0.000 ± 0%  -100.00%
```

`BenchmarkUnwrapDecode_FallbackRawCBOR` is new and should be folded into the
next full baseline refresh. The wrapped-envelope path is unchanged.
