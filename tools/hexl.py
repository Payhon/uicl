#!/usr/bin/env python3
"""HEXL/1 draft reference codec. ASCII transport, independently encoded byte chunks.

Library unpack() writes provisional bytes before footer verification. Callers must
not consume them as trusted output until unpack() returns. The CLI stages output
in a temporary file and publishes it only after successful verification.
"""
from __future__ import annotations

import argparse
import base64
import binascii
import hashlib
import hmac
import os
import re
import sys
import tempfile
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import BinaryIO
import json

MAX_CHUNK = 1_048_576
MAX_SIZE = (1 << 64) - 1
CODECS = ("hex", "base64", "base64url", "base32")
HEADER = re.compile(rb"!hexl 1 (hex|base64|base64url|base32) ([1-9][0-9]*)\n")
FOOTER = re.compile(rb"!end (0|[1-9][0-9]*) (0|[1-9][0-9]*) sha256:([0-9a-f]{64})\n")
RECORD = re.compile(rb"([0-9a-f]{16}):([^\r\n]+)\n")


class HexlError(ValueError):
    """The input is malformed, noncanonical, incomplete, or over a resource limit."""


@dataclass(frozen=True)
class Result:
    encoding: str
    chunk_bytes: int
    bytes: int
    lines: int
    sha256: str


def encode_chunk(raw: bytes, codec: str) -> bytes:
    if codec == "hex":
        return raw.hex().encode("ascii")
    if codec == "base64":
        return base64.b64encode(raw)
    if codec == "base64url":
        return base64.urlsafe_b64encode(raw).rstrip(b"=")
    if codec == "base32":
        return base64.b32encode(raw).rstrip(b"=")
    raise HexlError(f"Unsupported encoding: {codec}")


def decode_chunk(encoded: bytes, codec: str) -> bytes:
    try:
        if codec == "hex":
            if not re.fullmatch(rb"[0-9a-f]+", encoded) or len(encoded) % 2:
                raise HexlError("Invalid lowercase hexadecimal payload")
            raw = binascii.unhexlify(encoded)
        elif codec == "base64":
            raw = base64.b64decode(encoded, validate=True)
        elif codec == "base64url":
            if not re.fullmatch(rb"[A-Za-z0-9_-]+", encoded):
                raise HexlError("Invalid unpadded base64url alphabet")
            raw = base64.b64decode(encoded + b"=" * (-len(encoded) % 4),
                                   altchars=b"-_", validate=True)
        elif codec == "base32":
            if not re.fullmatch(rb"[A-Z2-7]+", encoded):
                raise HexlError("Invalid unpadded base32 alphabet")
            raw = base64.b32decode(encoded + b"=" * (-len(encoded) % 8), casefold=False)
        else:
            raise HexlError(f"Unsupported encoding: {codec}")
    except (binascii.Error, ValueError) as exc:
        raise HexlError(f"Cannot decode {codec} payload: {exc}") from exc
    if not raw or encode_chunk(raw, codec) != encoded:
        raise HexlError("Noncanonical encoding or nonzero unused padding bits")
    return raw


def _read_block(source: BinaryIO, size: int) -> bytes:
    """Fill a chunk even when a blocking stream returns short reads."""
    parts: list[bytes] = []
    remaining = size
    while remaining:
        part = source.read(remaining)
        if part is None:
            raise HexlError("Nonblocking streams are not supported")
        if not part:
            break
        parts.append(part)
        remaining -= len(part)
    return b"".join(parts)


def _line(source: BinaryIO, limit: int) -> bytes:
    value = source.readline(limit + 1)
    if len(value) > limit:
        raise HexlError("Line exceeds the declared or implementation limit")
    if not value:
        raise HexlError("Truncated input: missing required line or footer")
    if not value.endswith(b"\n") or b"\r" in value:
        raise HexlError("Canonical HEXL requires LF-terminated lines, without CR")
    return value


def pack(source: BinaryIO, target: BinaryIO, codec: str = "hex", chunk_bytes: int = 48) -> Result:
    if codec not in CODECS:
        raise HexlError("Unknown encoding")
    if not isinstance(chunk_bytes, int) or isinstance(chunk_bytes, bool) or not 1 <= chunk_bytes <= MAX_CHUNK:
        raise HexlError(f"chunk_bytes must be an integer between 1 and {MAX_CHUNK}")
    target.write(f"!hexl 1 {codec} {chunk_bytes}\n".encode("ascii"))
    digest = hashlib.sha256()
    total = lines = 0
    while True:
        raw = _read_block(source, chunk_bytes)
        if not raw:
            break
        if total + len(raw) > MAX_SIZE:
            raise HexlError("File exceeds the unsigned 64-bit byte-size limit")
        target.write(f"{total:016x}:".encode("ascii") + encode_chunk(raw, codec) + b"\n")
        digest.update(raw)
        total += len(raw)
        lines += 1
    value = digest.hexdigest()
    target.write(f"!end {total} {lines} sha256:{value}\n".encode("ascii"))
    return Result(codec, chunk_bytes, total, lines, value)


def unpack(source: BinaryIO, target: BinaryIO, max_bytes: int = 1 << 30,
           expected_sha256: str | None = None) -> Result:
    if not isinstance(max_bytes, int) or isinstance(max_bytes, bool) or not 0 <= max_bytes <= MAX_SIZE:
        raise HexlError("max_bytes must be an unsigned 64-bit integer")
    if expected_sha256 is not None and not re.fullmatch(r"[0-9a-f]{64}", expected_sha256):
        raise HexlError("Expected digest must be 64 lowercase hexadecimal characters")
    match = HEADER.fullmatch(_line(source, 128))
    if not match:
        raise HexlError("Invalid HEXL/1 header")
    codec = match[1].decode("ascii")
    chunk_bytes = int(match[2])
    if not 1 <= chunk_bytes <= MAX_CHUNK:
        raise HexlError("Declared chunk size exceeds implementation limit")
    digest = hashlib.sha256()
    total = lines = 0
    previous_short = False
    # Hex is the widest registered codec. Include record framing and footer.
    line_limit = max(192, 2 * chunk_bytes + 18)
    while True:
        line = _line(source, line_limit)
        if line.startswith(b"!end"):
            footer = FOOTER.fullmatch(line)
            if not footer:
                raise HexlError("Malformed footer")
            claimed_size, claimed_lines = int(footer[1]), int(footer[2])
            claimed_digest = footer[3].decode("ascii")
            if claimed_size != total or claimed_lines != lines:
                raise HexlError("Footer byte-size or record count mismatch")
            actual_digest = digest.hexdigest()
            if not hmac.compare_digest(claimed_digest, actual_digest):
                raise HexlError("Decoded payload digest mismatch")
            if expected_sha256 is not None and not hmac.compare_digest(expected_sha256, actual_digest):
                raise HexlError("Decoded payload does not match the trusted external digest")
            if source.read(1) != b"":
                raise HexlError("Trailing bytes after footer")
            return Result(codec, chunk_bytes, total, lines, actual_digest)
        record = RECORD.fullmatch(line)
        if not record:
            raise HexlError("Malformed data record")
        if previous_short:
            raise HexlError("Only the last data record may be shorter than chunk_bytes")
        if int(record[1], 16) != total:
            raise HexlError("Record offset mismatch: gap, reorder, or duplicate")
        raw = decode_chunk(record[2], codec)
        if not 1 <= len(raw) <= chunk_bytes:
            raise HexlError("Decoded record length is outside its declared bounds")
        if total + len(raw) > min(max_bytes, MAX_SIZE):
            raise HexlError("Decoded payload exceeds maximum allowed bytes")
        previous_short = len(raw) < chunk_bytes
        digest.update(raw)
        target.write(raw)
        total += len(raw)
        lines += 1


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    p = sub.add_parser("pack", help="Encode a file without executing or inspecting its format")
    p.add_argument("source", type=Path)
    p.add_argument("target", type=Path)
    p.add_argument("--encoding", choices=CODECS, default="hex")
    p.add_argument("--chunk", type=int, default=48)
    u = sub.add_parser("unpack", help="Decode, verify, and atomically publish an ordinary file")
    u.add_argument("source", type=Path)
    u.add_argument("target", type=Path)
    u.add_argument("--max-bytes", type=int, default=1 << 30)
    u.add_argument("--expect-sha256")
    for child in (p, u):
        child.add_argument("--force", action="store_true", help="Permit replacing the target")
    args = parser.parse_args()
    temp: Path | None = None
    try:
        if args.source.resolve() == args.target.resolve():
            raise HexlError("Source and target must be different paths")
        if args.target.exists() and not args.force:
            raise HexlError("Target exists; use --force to explicitly replace it")
        fd, name = tempfile.mkstemp(prefix=".hexl-", dir=args.target.parent)
        temp = Path(name)
        with args.source.open("rb") as source, os.fdopen(fd, "wb") as target:
            if args.command == "pack":
                result = pack(source, target, args.encoding, args.chunk)
            else:
                result = unpack(source, target, args.max_bytes, args.expect_sha256)
            target.flush()
            os.fsync(target.fileno())
        if args.force:
            os.replace(temp, args.target)
        else:
            # Same-directory hard link gives no-clobber publication on local FS.
            os.link(temp, args.target)
            temp.unlink()
        temp = None
        print(json.dumps(asdict(result), ensure_ascii=False), file=sys.stderr)
        return 0
    except (OSError, HexlError) as exc:
        print(f"hexl: {exc}", file=sys.stderr)
        return 2
    finally:
        if temp is not None:
            temp.unlink(missing_ok=True)


if __name__ == "__main__":
    raise SystemExit(main())
