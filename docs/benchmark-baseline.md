# Benchmark Baseline (v1 JSON mode)

Reference performance baseline for regression comparison, recorded 2026-09-11.
Figures in `FEATURES.md` that previously carried indicative `~` values cite this
file. Re-run the same command and diff with benchstat before accepting any
performance-sensitive change.

## Environment

| Factor    | Value                                                         |
| --------- | ------------------------------------------------------------- |
| Date      | 2026-09-11                                                    |
| Toolchain | go1.26.7 linux/amd64                                          |
| CPU       | AMD RYZEN AI MAX+ 395 w/ Radeon 8060S                         |
| JSON mode | v1 (default build; `env -u GOEXPERIMENT`)                     |
| Command   | `go test -run '^$' -bench . -benchmem -count=10 -timeout 40m` |

Notes:

- 67 benchmarks x 10 repetitions, summarized with `benchstat` (mean +- range
  across the 10 runs). Supersedes the 2026-08-15 baseline (go1.26.5, cbor
  v2.9.2): cross-version diffs against that table were dominated by toolchain
  and library noise, so it was retired as a comparison anchor.
- The 2026-09-11 error-contract change was verified separately from this
  baseline via a same-machine A/B (benchstat, 6641688 vs HEAD, 10 runs): B/op
  and allocs/op identical on every hot-path benchmark; WrapEncode,
  UnwrapDecode, EncodePooled, CBORCodec_Decode statistically indistinguishable;
  geomean sec/op -10.2%. Error paths only allocate on failure — happy-path
  cost is nil, now measured rather than assumed.
- High +- percentages on short benchmarks reflect scheduler noise on this
  machine; treat sub-microsecond differences as noise unless benchstat reports
  them significant across full runs.
- The raw output is not committed; the summary below plus the reproduction
  command above fully determine it.

## Summary

goos: linux
goarch: amd64
pkg: github.com/larsartmann/go-codec
cpu: AMD RYZEN AI MAX+ 395 w/ Radeon 8060S\
│ /tmp/codec-rerun-v1.txt │
│ sec/op │
AutoDetect/json-32 99.12n ± 32%
AutoDetect/cbor-32 126.7n ± 27%
AutoDetect/unknown-32 146.4n ± 6%
AutoDetectDebug-32 147.7n ± 34%
JSONCodec_Encode-32 263.1n ± 1%
JSONCodec_Decode-32 600.3n ± 1%
CBORCodec_Encode-32 234.4n ± 5%
CBORCodec_Decode-32 379.1n ± 9%
CodecComparison_Encode/JSON-32 262.7n ± 15%
CodecComparison_Encode/CBOR-32 230.3n ± 32%
CodecComparison_Decode/JSON-32 665.9n ± 22%
CodecComparison_Decode/CBOR-32 425.1n ± 11%
RawCodec_Encode-32 16.35n ± 5%
RawCodec_Decode-32 31.93n ± 19%
CBORCompact_vs_Canon_Size/Canonical/Encode-32 128.8n ± 37%
CBORCompact_vs_Canon_Size/Compact/Encode-32 139.5n ± 3%
CBORCompact_vs_Canon_Decode/Canonical-32 281.5n ± 11%
CBORCompact_vs_Canon_Decode/Compact-32 305.6n ± 10%
RealisticPayload_Encode/JSON-32 461.1n ± 55%
RealisticPayload_Encode/CBOR-32 288.1n ± 3%
RealisticPayload_Encode/CBOR_compact_toarray-32 381.1n ± 21%
RealisticPayload_Decode/JSON-32 2.496µ ± 67%
RealisticPayload_Decode/CBOR-32 875.4n ± 10%
RealisticPayload_Decode/CBOR_compact_toarray-32 776.9n ± 49%
BufferEncoder/JSON-32 523.8n ± 104%
BufferEncoder/CBOR-32 676.6n ± 54%
BufferEncoder/CBOR_compact-32 726.2n ± 5%
EncodePooled/JSON-32 523.4n ± 26%
EncodePooled/CBOR-32 350.5n ± 81%
EncodePooled/CBOR_compact-32 754.5n ± 5%
TagTradeoffs_Encode/small/map-32 240.8n ± 8%
TagTradeoffs_Encode/small/toarray-32 190.7n ± 6%
TagTradeoffs_Encode/small/keyasint-32 208.7n ± 15%
TagTradeoffs_Encode/medium/map-32 637.3n ± 9%
TagTradeoffs_Encode/medium/toarray-32 448.4n ± 25%
TagTradeoffs_Encode/medium/keyasint-32 457.1n ± 36%
TagTradeoffs_Encode/large/map-32 316.0n ± 18%
TagTradeoffs_Encode/large/toarray-32 192.0n ± 10%
TagTradeoffs_Encode/large/keyasint-32 244.8n ± 3%
TagTradeoffs_Decode/small/map-32 234.8n ± 1%
TagTradeoffs_Decode/small/toarray-32 178.4n ± 8%
TagTradeoffs_Decode/small/keyasint-32 243.2n ± 21%
TagTradeoffs_Decode/medium/map-32 979.9n ± 13%
TagTradeoffs_Decode/medium/toarray-32 841.3n ± 14%
TagTradeoffs_Decode/medium/keyasint-32 859.7n ± 11%
TagTradeoffs_Decode/large/map-32 893.7n ± 7%
TagTradeoffs_Decode/large/toarray-32 586.7n ± 6%
TagTradeoffs_Decode/large/keyasint-32 759.3n ± 7%
CBORReflectionCache/encode-32 363.9n ± 12%
CBORReflectionCache/decode-32 958.5n ± 7%
ObserveCodec/encode/raw-32 385.9n ± 11%
ObserveCodec/encode/observed-32 376.4n ± 8%
ObserveCodec/decode/raw-32 940.9n ± 12%
ObserveCodec/decode/observed-32 885.7n ± 9%
ObserveCodec/encode_pooled/observed-32 369.9n ± 10%
WrapEncode-32 238.6n ± 10%
UnwrapDecode-32 864.7n ± 20%
UnwrapDecode_FallbackRawCBOR-32 1.835n ± 13%
NormalizeForJSON-32 1.438µ ± 16%
JSONCodec_MarshalUnmarshal-32 1.043µ ± 9%
Size-32 257.1n ± 44%
StreamingJSON_Encode-32 51.77µ ± 37%
StreamingJSON_Decode-32 306.5µ ± 5%
StreamingCBOR_Encode-32 32.50µ ± 28%
StreamingCBOR_Decode-32 99.84µ ± 8%
TranscodeToJSON_CBOR_To_JSON-32 5.662µ ± 7%
TranscodeToJSON_JSON_Passthrough-32 2.098n ± 5%
TranscodeToJSON_NestedDeep-32 4.967µ ± 19%
geomean 471.2n

    │ /tmp/codec-rerun-v1.txt │
    │          B/op           │

AutoDetect/json-32 64.00 ± 0%
AutoDetect/cbor-32 64.00 ± 0%
AutoDetect/unknown-32 160.0 ± 0%
AutoDetectDebug-32 160.0 ± 0%
JSONCodec_Encode-32 192.0 ± 0%
JSONCodec_Decode-32 592.0 ± 0%
CBORCodec_Encode-32 96.00 ± 0%
CBORCodec_Decode-32 416.0 ± 0%
CodecComparison_Encode/JSON-32 192.0 ± 0%
CodecComparison_Encode/CBOR-32 96.00 ± 0%
CodecComparison_Decode/JSON-32 592.0 ± 0%
CodecComparison_Decode/CBOR-32 416.0 ± 0%
RawCodec_Encode-32 24.00 ± 0%
RawCodec_Decode-32 48.00 ± 0%
CBORCompact_vs_Canon_Size/Canonical/Encode-32 112.0 ± 0%
CBORCompact_vs_Canon_Size/Compact/Encode-32 112.0 ± 0%
CBORCompact_vs_Canon_Decode/Canonical-32 77.00 ± 0%
CBORCompact_vs_Canon_Decode/Compact-32 77.00 ± 0%
RealisticPayload_Encode/JSON-32 288.0 ± 0%
RealisticPayload_Encode/CBOR-32 224.0 ± 0%
RealisticPayload_Encode/CBOR_compact_toarray-32 160.0 ± 0%
RealisticPayload_Decode/JSON-32 608.0 ± 0%
RealisticPayload_Decode/CBOR-32 312.0 ± 0%
RealisticPayload_Decode/CBOR_compact_toarray-32 312.0 ± 0%
BufferEncoder/JSON-32 400.0 ± 0%
BufferEncoder/CBOR-32 176.0 ± 0%
BufferEncoder/CBOR_compact-32 176.0 ± 0%
EncodePooled/JSON-32 401.0 ± 0%
EncodePooled/CBOR-32 176.0 ± 0%
EncodePooled/CBOR_compact-32 176.0 ± 0%
TagTradeoffs_Encode/small/map-32 64.00 ± 0%
TagTradeoffs_Encode/small/toarray-32 48.00 ± 0%
TagTradeoffs_Encode/small/keyasint-32 48.00 ± 0%
TagTradeoffs_Encode/medium/map-32 224.0 ± 0%
TagTradeoffs_Encode/medium/toarray-32 160.0 ± 0%
TagTradeoffs_Encode/medium/keyasint-32 176.0 ± 0%
TagTradeoffs_Encode/large/map-32 288.0 ± 0%
TagTradeoffs_Encode/large/toarray-32 160.0 ± 0%
TagTradeoffs_Encode/large/keyasint-32 176.0 ± 0%
TagTradeoffs_Decode/small/map-32 96.00 ± 0%
TagTradeoffs_Decode/small/toarray-32 96.00 ± 0%
TagTradeoffs_Decode/small/keyasint-32 96.00 ± 0%
TagTradeoffs_Decode/medium/map-32 312.0 ± 0%
TagTradeoffs_Decode/medium/toarray-32 312.0 ± 0%
TagTradeoffs_Decode/medium/keyasint-32 312.0 ± 0%
TagTradeoffs_Decode/large/map-32 352.0 ± 0%
TagTradeoffs_Decode/large/toarray-32 352.0 ± 0%
TagTradeoffs_Decode/large/keyasint-32 352.0 ± 0%
CBORReflectionCache/encode-32 336.0 ± 0%
CBORReflectionCache/decode-32 312.0 ± 0%
ObserveCodec/encode/raw-32 336.0 ± 0%
ObserveCodec/encode/observed-32 336.0 ± 0%
ObserveCodec/decode/raw-32 312.0 ± 0%
ObserveCodec/decode/observed-32 312.0 ± 0%
ObserveCodec/encode_pooled/observed-32 176.0 ± 0%
WrapEncode-32 208.0 ± 0%
UnwrapDecode-32 336.0 ± 0%
UnwrapDecode_FallbackRawCBOR-32 0.000 ± 0%
NormalizeForJSON-32 1.182Ki ± 0%
JSONCodec_MarshalUnmarshal-32 376.0 ± 0%
Size-32 112.0 ± 0%
StreamingJSON_Encode-32 10.97Ki ± 0%
StreamingJSON_Decode-32 33.92Ki ± 0%
StreamingCBOR_Encode-32 11.03Ki ± 0%
StreamingCBOR_Decode-32 32.75Ki ± 0%
TranscodeToJSON_CBOR_To_JSON-32 4.022Ki ± 0%
TranscodeToJSON_JSON_Passthrough-32 0.000 ± 0%
TranscodeToJSON_NestedDeep-32 4.504Ki ± 0%
geomean ¹
¹ summaries must be >0 to compute geomean

    │ /tmp/codec-rerun-v1.txt │
    │        allocs/op        │

AutoDetect/json-32 1.000 ± 0%
AutoDetect/cbor-32 1.000 ± 0%
AutoDetect/unknown-32 4.000 ± 0%
AutoDetectDebug-32 4.000 ± 0%
JSONCodec_Encode-32 6.000 ± 0%
JSONCodec_Decode-32 12.00 ± 0%
CBORCodec_Encode-32 2.000 ± 0%
CBORCodec_Decode-32 9.000 ± 0%
CodecComparison_Encode/JSON-32 6.000 ± 0%
CodecComparison_Encode/CBOR-32 2.000 ± 0%
CodecComparison_Decode/JSON-32 12.00 ± 0%
CodecComparison_Decode/CBOR-32 9.000 ± 0%
RawCodec_Encode-32 1.000 ± 0%
RawCodec_Decode-32 2.000 ± 0%
CBORCompact_vs_Canon_Size/Canonical/Encode-32 2.000 ± 0%
CBORCompact_vs_Canon_Size/Compact/Encode-32 2.000 ± 0%
CBORCompact_vs_Canon_Decode/Canonical-32 3.000 ± 0%
CBORCompact_vs_Canon_Decode/Compact-32 3.000 ± 0%
RealisticPayload_Encode/JSON-32 1.000 ± 0%
RealisticPayload_Encode/CBOR-32 1.000 ± 0%
RealisticPayload_Encode/CBOR_compact_toarray-32 1.000 ± 0%
RealisticPayload_Decode/JSON-32 16.00 ± 0%
RealisticPayload_Decode/CBOR-32 9.000 ± 0%
RealisticPayload_Decode/CBOR_compact_toarray-32 9.000 ± 0%
BufferEncoder/JSON-32 2.000 ± 0%
BufferEncoder/CBOR-32 2.000 ± 0%
BufferEncoder/CBOR_compact-32 2.000 ± 0%
EncodePooled/JSON-32 2.000 ± 0%
EncodePooled/CBOR-32 2.000 ± 0%
EncodePooled/CBOR_compact-32 2.000 ± 0%
TagTradeoffs_Encode/small/map-32 1.000 ± 0%
TagTradeoffs_Encode/small/toarray-32 1.000 ± 0%
TagTradeoffs_Encode/small/keyasint-32 1.000 ± 0%
TagTradeoffs_Encode/medium/map-32 1.000 ± 0%
TagTradeoffs_Encode/medium/toarray-32 1.000 ± 0%
TagTradeoffs_Encode/medium/keyasint-32 1.000 ± 0%
TagTradeoffs_Encode/large/map-32 1.000 ± 0%
TagTradeoffs_Encode/large/toarray-32 1.000 ± 0%
TagTradeoffs_Encode/large/keyasint-32 1.000 ± 0%
TagTradeoffs_Decode/small/map-32 4.000 ± 0%
TagTradeoffs_Decode/small/toarray-32 4.000 ± 0%
TagTradeoffs_Decode/small/keyasint-32 4.000 ± 0%
TagTradeoffs_Decode/medium/map-32 9.000 ± 0%
TagTradeoffs_Decode/medium/toarray-32 9.000 ± 0%
TagTradeoffs_Decode/medium/keyasint-32 9.000 ± 0%
TagTradeoffs_Decode/large/map-32 10.00 ± 0%
TagTradeoffs_Decode/large/toarray-32 10.00 ± 0%
TagTradeoffs_Decode/large/keyasint-32 10.00 ± 0%
CBORReflectionCache/encode-32 2.000 ± 0%
CBORReflectionCache/decode-32 9.000 ± 0%
ObserveCodec/encode/raw-32 2.000 ± 0%
ObserveCodec/encode/observed-32 2.000 ± 0%
ObserveCodec/decode/raw-32 9.000 ± 0%
ObserveCodec/decode/observed-32 9.000 ± 0%
ObserveCodec/encode_pooled/observed-32 2.000 ± 0%
WrapEncode-32 3.000 ± 0%
UnwrapDecode-32 8.000 ± 0%
UnwrapDecode_FallbackRawCBOR-32 0.000 ± 0%
NormalizeForJSON-32 19.00 ± 0%
JSONCodec_MarshalUnmarshal-32 8.000 ± 0%
Size-32 2.000 ± 0%
StreamingJSON_Encode-32 100.0 ± 0%
StreamingJSON_Decode-32 914.0 ± 0%
StreamingCBOR_Encode-32 101.0 ± 0%
StreamingCBOR_Decode-32 905.0 ± 0%
TranscodeToJSON_CBOR_To_JSON-32 102.0 ± 0%
TranscodeToJSON_JSON_Passthrough-32 0.000 ± 0%
TranscodeToJSON_NestedDeep-32 78.00 ± 0%
geomean ¹
¹ summaries must be >0 to compute geomean
