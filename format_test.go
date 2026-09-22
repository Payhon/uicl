package uicl

import (
	"bytes"
	"encoding/base64"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatRepositoryDocuments(t *testing.T) {
	paths := []string{"README.uicl", "UICL-1.0-Structured.uicl"}
	for _, root := range []string{"profiles", "meta", "examples", "guides", "spec"} {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && strings.HasSuffix(path, ".uicl") {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			d := Parse(path, source, ParseOptions{})
			if HasErrors(d.Diagnostics) {
				t.Fatalf("parse %+v", d.Diagnostics)
			}
			formatted, ds := Format(d)
			if HasErrors(ds) {
				t.Fatalf("format %+v", ds)
			}
			again, ds := Format(Parse(path, formatted, ParseOptions{}))
			if HasErrors(ds) || !bytes.Equal(formatted, again) {
				t.Fatalf("not idempotent %+v", ds)
			}
		})
	}
}

func TestFormatCoreAndHosts(t *testing.T) {
	core := "uicl  \"1.0\"\n\n// Comment  stays  \nthing  #x  \"中文🙂\"\n  values: [ 1,2 , 3 ]\n  settings: { mode : system,count:1.00 }\n  text: |\n    native   text\n    \t原文\n\n  other: \"a  b\"\n"
	want := "uicl \"1.0\"\n\n// Comment  stays  \nthing #x \"中文🙂\"\n  values: [1, 2, 3]\n  settings: {mode: system, count: 1.00}\n  text: |\n    native   text\n    \t原文\n\n  other: \"a  b\"\n"
	fragment := strings.TrimPrefix(core, "uicl  \"1.0\"\n")
	fragmentWant := strings.TrimPrefix(want, "uicl \"1.0\"\n")
	encoded := base64.RawURLEncoding.EncodeToString([]byte(fragment))
	encodedWant := base64.RawURLEncoding.EncodeToString([]byte(fragmentWant))
	cases := []struct{ name, source, want string }{
		{"core", core, want},
		{"bom crlf", "\xef\xbb\xbf" + strings.ReplaceAll(core, "\n", "\r\n"), "\xef\xbb\xbf" + strings.ReplaceAll(want, "\n", "\r\n")},
		{"markdown", mdMarker + "# Unchanged  body\n\n<!-- uicl\n" + fragment + "-->\n\n**正文**  \n", mdMarker + "# Unchanged  body\n\n<!-- uicl\n" + fragmentWant + "-->\n\n**正文**  \n"},
		{"html", htmlHeader + "<p style='color:red'> Unchanged  body </p><!-- uicl\n" + fragment + "-->\n</body></html>", htmlHeader + "<p style='color:red'> Unchanged  body </p><!-- uicl\n" + fragmentWant + "-->\n</body></html>"},
		{"encoded markdown", mdMarker + "<!-- uicl:base64url\n" + encoded + "\n-->\n", mdMarker + "<!-- uicl:base64url\n" + encodedWant + "-->\n"},
		{"script", htmlHeader + "<script type=text/plain data-uicl>\n" + core + "</script>", htmlHeader + "<script type=text/plain data-uicl>\n" + want + "</script>"},
		{"encoded script", htmlHeader + "<script type=text/plain data-uicl-encoding=base64url>" + base64.RawURLEncoding.EncodeToString([]byte(core)) + "</script>", htmlHeader + "<script type=text/plain data-uicl-encoding=base64url>" + base64.RawURLEncoding.EncodeToString([]byte(want)) + "</script>"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			d := Parse("test.uicl", []byte(tt.source), ParseOptions{})
			if HasErrors(d.Diagnostics) {
				t.Fatalf("input parse: %+v", d.Diagnostics)
			}
			formatted, ds := Format(d)
			if HasErrors(ds) {
				t.Fatalf("format: %+v", ds)
			}
			if string(formatted) != tt.want {
				t.Fatalf("want:\n%s\ngot:\n%s", tt.want, formatted)
			}
			second, ds := Format(Parse("test.uicl", formatted, ParseOptions{}))
			if HasErrors(ds) || !bytes.Equal(formatted, second) {
				t.Fatalf("not idempotent: %+v", ds)
			}
			if !bytes.Equal(semanticNodes(d.Nodes), semanticNodes(Parse("test.uicl", formatted, ParseOptions{}).Nodes)) {
				t.Fatal("semantic difference")
			}
		})
	}
}

func TestFormatInvalidAndNoop(t *testing.T) {
	d := Parse("broken.uicl", []byte("uicl \"1.0\"\nthing\n  value: [\n"), ParseOptions{Recover: true})
	if out, ds := Format(d); out != nil || !HasErrors(ds) {
		t.Fatal("formatter must refuse broken input")
	}
	payload := []byte("thing #ok\n  value: 2\n")
	enc := base64.RawURLEncoding.EncodeToString(payload)
	source := []byte(mdMarker + "<!-- uicl:base64url\n " + enc[:10] + "\n" + enc[10:] + " \n-->\n")
	out, ds := Format(Parse("encoded.md", source, ParseOptions{}))
	if HasErrors(ds) || !bytes.Equal(out, source) {
		t.Fatalf("unchanged encoded payload was rewritten: %+v", ds)
	}
}

func TestTextEdits(t *testing.T) {
	source := []byte("中文🙂 value\n")
	d := NewDocument("x", source)
	valid := []TextEdit{{Range: d.Span(0, 6), NewText: "标题"}, {Range: d.Span(11, 16), NewText: "field"}}
	out, err := ApplyEdits(source, Digest(source), valid)
	if err != nil || string(out) != "标题🙂 field\n" {
		t.Fatalf("valid edits: %q %v", out, err)
	}
	cases := []struct {
		name   string
		digest string
		edits  []TextEdit
	}{
		{"conflict", "old", valid},
		{"missing digest", "", valid},
		{"overlap", Digest(source), []TextEdit{{Range: d.Span(0, 6)}, {Range: d.Span(3, 10)}}},
		{"same offset", Digest(source), []TextEdit{{Range: d.Span(0, 0)}, {Range: d.Span(0, 0)}}},
		{"rune split", Digest(source), []TextEdit{{Range: Range{Start: Position{Offset: 1}, End: Position{Offset: 3}}}}},
		{"out of bounds", Digest(source), []TextEdit{{Range: Range{Start: Position{Offset: 0}, End: Position{Offset: 100}}}}},
		{"negative", Digest(source), []TextEdit{{Range: Range{Start: Position{Offset: -1}, End: Position{Offset: 0}}}}},
		{"backwards", Digest(source), []TextEdit{{Range: Range{Start: Position{Offset: 3}, End: Position{Offset: 0}}}}},
		{"invalid replacement", Digest(source), []TextEdit{{Range: d.Span(0, 0), NewText: string([]byte{0xff})}}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ApplyEdits(source, tt.digest, tt.edits); err == nil {
				t.Fatal("invalid edit accepted")
			}
		})
	}
}

func FuzzFormat(f *testing.F) {
	f.Add("uicl \"1.0\"\nthing #a\n  values: [1,2]\n")
	f.Add(mdMarker + "<!-- uicl\nthing  #a\n-->\n")
	f.Fuzz(func(t *testing.T, source string) {
		if len(source) > 65536 {
			t.Skip()
		}
		d := Parse("fuzz.uicl", []byte(source), ParseOptions{})
		if HasErrors(d.Diagnostics) {
			return
		}
		out, ds := Format(d)
		if HasErrors(ds) {
			t.Fatalf("valid source failed formatting: %+v", ds)
		}
		again, ds := Format(Parse("fuzz.uicl", out, ParseOptions{}))
		if HasErrors(ds) || !bytes.Equal(out, again) {
			t.Fatalf("format must be idempotent: %+v", ds)
		}
	})
}
