# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability, please **do not** open a public issue.

Email: **security@lars.software**

Include:
- A description of the vulnerability
- Steps to reproduce or proof of concept
- Affected versions (if known)

You will receive a response within 48 hours. If the vulnerability is confirmed,
a fix will be prioritized and a security advisory published.

## Security Considerations

- **Depth cap**: `normalizeForJSON` limits recursion to 100 levels to prevent
  stack-exhaustion DoS from adversarial CBOR input (v1 JSON mode only).
- **AutoDetect size guard**: trial-decode is skipped for payloads over 1 MiB
  to prevent DoS via oversized input.
- **CBOR modes are NOT interoperable**: `CBORCodec` (canonical) and
  `CBORCompactCodec` (core deterministic) use different key sort orders.
  Data written by one cannot be read by the other.
- **`time.Time`**: must be `.UTC()` before CBOR encoding. CBOR uses
  `TimeUnixDynamic` and decoded times reconstruct in `time.Local`.
