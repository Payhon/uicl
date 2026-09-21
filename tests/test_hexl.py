import base64
import hashlib
import io
import random
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "tools"))
from hexl import CODECS, HexlError, decode_chunk, pack, unpack


def encoded(raw=b"AML\n", codec="hex", chunk=4):
    out = io.BytesIO()
    pack(io.BytesIO(raw), out, codec, chunk)
    return out.getvalue()


class HexlTests(unittest.TestCase):
    def rejects(self, wire, **kwargs):
        with self.assertRaises(HexlError):
            unpack(io.BytesIO(wire), io.BytesIO(), **kwargs)

    def test_known_vector(self):
        wire = encoded()
        self.assertEqual(wire, b"!hexl 1 hex 4\n0000000000000000:414d4c0a\n"
                         b"!end 4 1 sha256:e93ef79161e260f8bd71148ee3541839cb8f6b29bd3dcfbaffa62c3db82985db\n")

    def test_roundtrip_matrix(self):
        rng = random.Random(20260919)
        for codec in CODECS:
            for chunk in (1, 2, 3, 4, 5, 16, 48, 127, 4096):
                for size in (0, 1, 2, 3, 4, 47, 48, 49, 256, 4097):
                    with self.subTest(codec=codec, chunk=chunk, size=size):
                        raw = rng.randbytes(size)
                        out = io.BytesIO()
                        report = unpack(io.BytesIO(encoded(raw, codec, chunk)), out)
                        self.assertEqual(out.getvalue(), raw)
                        self.assertEqual(report.sha256, hashlib.sha256(raw).hexdigest())
                        self.assertEqual(report.lines, (size + chunk - 1) // chunk)

    def test_all_byte_values(self):
        for codec in CODECS:
            raw = bytes(range(256))
            out = io.BytesIO()
            unpack(io.BytesIO(encoded(raw, codec, 17)), out)
            self.assertEqual(out.getvalue(), raw)

    def test_empty(self):
        self.assertEqual(len(encoded(b"").splitlines()), 2)
        out = io.BytesIO()
        unpack(io.BytesIO(encoded(b"")), out, max_bytes=0)
        self.assertEqual(out.getvalue(), b"")

    def test_short_reads(self):
        class SmallReads(io.BytesIO):
            def read(self, n=-1):
                return super().read(min(n, 2) if n >= 0 else 2)
        out = io.BytesIO()
        pack(SmallReads(b"abcdefghi"), out, "base64url", 6)
        self.assertEqual(out.getvalue(), encoded(b"abcdefghi", "base64url", 6))

    def test_independent_lines(self):
        raw = bytes(range(200))
        for codec in CODECS:
            lines = encoded(raw, codec, 48).splitlines()[1:-1]
            restored = b"".join(decode_chunk(row.split(b":", 1)[1], codec) for row in lines)
            self.assertEqual(restored, raw)

    def test_truncated_footer(self):
        self.rejects(b"\n".join(encoded().split(b"\n")[:-2]) + b"\n")

    def test_missing_terminal_lf(self):
        self.rejects(encoded()[:-1])

    def test_crlf(self):
        self.rejects(encoded().replace(b"\n", b"\r\n"))

    def test_bom(self):
        self.rejects(b"\xef\xbb\xbf" + encoded())

    def test_trailing_data(self):
        self.rejects(encoded() + b"unexpected\n")

    def test_blank_line(self):
        self.rejects(encoded().replace(b"!end", b"\n!end"))

    def test_bad_header(self):
        for header in (b"!hexl 2 hex 4", b"!hexl 1 base62 4", b"!hexl 1 hex 0",
                       b"!hexl 1 hex 04", b"!hexl 1 hex 1048577"):
            self.rejects(header + b"\n" + encoded().split(b"\n", 1)[1])

    def test_wrong_offset(self):
        self.rejects(encoded().replace(b"0000000000000000:", b"0000000000000001:"))

    def test_duplicate_record(self):
        lines = encoded().splitlines(keepends=True)
        self.rejects(lines[0] + lines[1] + lines[1] + lines[2])

    def test_reordered_records(self):
        lines = encoded(b"abcdefgh", chunk=4).splitlines(keepends=True)
        self.rejects(lines[0] + lines[2] + lines[1] + lines[3])

    def test_gap(self):
        lines = encoded(b"abcdefghijkl", chunk=4).splitlines(keepends=True)
        self.rejects(lines[0] + lines[1] + lines[3] + lines[4])

    def test_short_nonfinal(self):
        wire = encoded(b"abcd", chunk=2).replace(b"!hexl 1 hex 2", b"!hexl 1 hex 4")
        self.rejects(wire)

    def test_oversized_payload(self):
        self.rejects(encoded().replace(b"414d4c0a", b"414d4c0a00"))

    def test_bad_digest(self):
        self.rejects(encoded().replace(b"e93ef791", b"093ef791"))

    def test_wrong_count(self):
        self.rejects(encoded().replace(b"!end 4 1", b"!end 4 2"))

    def test_wrong_size(self):
        self.rejects(encoded().replace(b"!end 4 1", b"!end 5 1"))

    def test_maximum_decoded_bytes(self):
        self.rejects(encoded(), max_bytes=3)

    def test_expected_digest(self):
        out = io.BytesIO()
        unpack(io.BytesIO(encoded()), out, expected_sha256=hashlib.sha256(b"AML\n").hexdigest())
        self.rejects(encoded(), expected_sha256="0" * 64)

    def test_invalid_alphabet_and_noncanonical_bits(self):
        cases = [("hex", b"AA"), ("hex", b"a"), ("hex", b"00 11"),
                 ("base64", b"Zh=="), ("base64", b"Zg"), ("base64", b"Zg===") ,
                 ("base64url", b"Zh"), ("base64url", b"Zg=="),
                 ("base64url", b"Z+"), ("base64url", b"A"),
                 ("base32", b"MZ"), ("base32", b"my"), ("base32", b"MY======")]
        for codec, value in cases:
            with self.subTest(codec=codec, payload=value):
                with self.assertRaises(HexlError):
                    decode_chunk(value, codec)

    def test_line_limit(self):
        self.rejects(b"!hexl 1 hex 1\n" + b"0" * 250 + b"\n")

    def test_invalid_encoder_parameters(self):
        for chunk in (0, -1, 1048577, True, 1.5):
            with self.assertRaises(HexlError):
                pack(io.BytesIO(b"x"), io.BytesIO(), chunk_bytes=chunk)

    def test_cli_roundtrip(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            src, wire, restored = root / "src.bin", root / "src.b64l", root / "restored.bin"
            src.write_bytes(bytes(range(256)))
            cli = [sys.executable, str(ROOT / "tools" / "hexl.py")]
            subprocess.run(cli + ["pack", str(src), str(wire), "--encoding", "base64url"],
                           check=True, capture_output=True)
            subprocess.run(cli + ["unpack", str(wire), str(restored)], check=True, capture_output=True)
            self.assertEqual(src.read_bytes(), restored.read_bytes())

    def test_cli_failure_does_not_publish(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            wire, target = root / "bad.hexl", root / "restored.bin"
            wire.write_bytes(encoded().replace(b"e93ef791", b"093ef791"))
            result = subprocess.run([sys.executable, str(ROOT / "tools" / "hexl.py"),
                                     "unpack", str(wire), str(target)], capture_output=True)
            self.assertEqual(result.returncode, 2)
            self.assertFalse(target.exists())
            self.assertFalse(list(root.glob(".hexl-*")))

    def test_cli_refuses_overwrite(self):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            wire, target = root / "ok.hexl", root / "existing.bin"
            wire.write_bytes(encoded())
            target.write_bytes(b"keep")
            result = subprocess.run([sys.executable, str(ROOT / "tools" / "hexl.py"),
                                     "unpack", str(wire), str(target)], capture_output=True)
            self.assertEqual(result.returncode, 2)
            self.assertEqual(target.read_bytes(), b"keep")


if __name__ == "__main__":
    unittest.main()
