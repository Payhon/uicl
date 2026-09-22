package uicl

import (
	"bytes"
	"encoding/base64"
	"os"
	"strings"
	"testing"
)

const mdMarker = "<!-- uicl:markdown 1.0 -->\n\n"
const htmlHeader = "<!doctype html><html><head><!-- uicl:html 1.0 --></head><body>"

func hostCode(d *Document, code string) bool {
	for _, diag := range d.Diagnostics {
		if diag.Code == code {
			return true
		}
	}
	return false
}
func mustHost(t *testing.T, source string) *Document {
	t.Helper()
	d := Parse("test.uicl", []byte(source), ParseOptions{Recover: true})
	if HasErrors(d.Diagnostics) {
		t.Fatalf("parse host: %+v", d.Diagnostics)
	}
	return d
}

func TestHostedRepositoryDocuments(t *testing.T) {
	for _, name := range []string{"examples/hosted/article.uicl", "examples/hosted/page.uicl", "examples/hosted/data-block.uicl", "spec/UICL-1.0.uicl", "README.uicl"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			d := mustHost(t, string(source))
			if len(d.Islands) == 0 {
				t.Fatal("no annotation found")
			}
			if !bytes.Equal(source, d.Source) {
				t.Fatal("source changed")
			}
			var joined strings.Builder
			for _, token := range d.Tokens {
				joined.WriteString(token.Text)
			}
			if joined.String() != string(source) {
				t.Fatal("host token coverage is not byte-exact")
			}
		})
	}
}

func TestMarkdownHostBoundaries(t *testing.T) {
	annotation := "<!-- uicl\ndoc.block #a\n  target: next\n-->\n\n"
	cases := []struct {
		name, source, code string
		count              int
	}{
		{"next", mdMarker + annotation + "# Title\n", "", 1},
		{"fence inert", mdMarker + "```html\n" + annotation + "```\n", "", 0},
		{"quote inert", mdMarker + "> <!-- uicl\n> thing #x\n> -->\n", "", 0},
		{"list inert", mdMarker + "- <!-- uicl\n  thing #x\n  -->\n", "", 0},
		{"indented inert", mdMarker + "    <!-- uicl\n    thing #x\n    -->\n", "", 0},
		{"column", mdMarker + " " + annotation + "# Title\n", "HOST_ANNOTATION_COLUMN", 0},
		{"missing next", mdMarker + annotation, "HOST_TARGET_MISSING", 1},
		{"bad target", mdMarker + strings.Replace(annotation, "next", "previous", 1) + "text", "MARKDOWN_TARGET_REQUIRES_NEXT", 1},
		{"duplicate", mdMarker + mdMarker + "text\n", "HOST_DUPLICATE_MARKER", 0},
		{"version", strings.Replace(mdMarker, "1.0", "2.0", 1), "HOST_MODE_OR_VERSION_CONFLICT", 0},
		{"conflict", mdMarker + "<!-- uicl:html 1.0 -->\n", "HOST_MODE_OR_VERSION_CONFLICT", 0},
		{"truncated", mdMarker + strings.TrimSuffix(annotation, "-->\n\n"), "HOST_COMMENT_BOUNDARY", 0},
		{"two comments one block", mdMarker + "<!-- uicl\nthing #a\n--><!-- extra -->\n", "HOST_COMMENT_BOUNDARY", 0},
		{"trailing text", mdMarker + "<!-- uicl\nthing #a\n-->text\n", "HOST_COMMENT_BOUNDARY", 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			d := Parse("test.md", []byte(tt.source), ParseOptions{Recover: true})
			if tt.code == "" && HasErrors(d.Diagnostics) {
				t.Fatalf("unexpected errors: %+v", d.Diagnostics)
			}
			if tt.code != "" && !hostCode(d, tt.code) {
				t.Fatalf("want %s got %+v", tt.code, d.Diagnostics)
			}
			if len(d.Islands) != tt.count {
				t.Fatalf("want %d islands got %d", tt.count, len(d.Islands))
			}
		})
	}
	d := mustHost(t, mdMarker+annotation+"<!-- uicl\nthing #b\n-->\n\n# Target\n")
	if d.Islands[0].TargetRange == nil || string(d.Source[d.Islands[0].TargetRange.Start.Offset:d.Islands[0].TargetRange.End.Offset]) != "# Target" {
		t.Fatalf("next skipped annotation incorrectly: %+v", d.Islands[0].TargetRange)
	}
}

func TestHTMLHostBoundaries(t *testing.T) {
	annotation := "<!-- uicl\ndoc.block #contract\n  target: {html_id: \"summary\"}\n-->"
	body := "<p id=summary>Visible</p>" + annotation
	data := "<script type=\"text/plain\" data-uicl>\nuicl \"1.0\"\nthing #data\n</script>"
	cases := []struct {
		name, source, code string
		count              int
	}{
		{"annotation", htmlHeader + body + "</body></html>", "", 1},
		{"data", htmlHeader + data, "", 1},
		{"bad script type", htmlHeader + strings.Replace(data, "text/plain", "text/javascript", 1), "HOST_SCRIPT_TYPE", 0},
		{"unclosed script", htmlHeader + strings.TrimSuffix(data, "</script>"), "HOST_UNCLOSED_DATA_BLOCK", 0},
		{"missing id", htmlHeader + annotation, "HOST_TARGET_MISSING_OR_DUPLICATE", 1},
		{"duplicate id", htmlHeader + body + "<p id=summary>Duplicate</p>", "HOST_TARGET_MISSING_OR_DUPLICATE", 1},
		{"duplicate target", htmlHeader + body + strings.Replace(annotation, "#contract", "#other", 1), "HOST_TARGET_CONFLICT", 2},
		{"bad target", htmlHeader + strings.Replace(annotation, "{html_id: \"summary\"}", "next", 1), "HTML_TARGET_REQUIRES_ID", 1},
		{"template inert", htmlHeader + "<template>" + annotation + data + "</template>", "", 0},
		{"code inert", htmlHeader + "<code>" + annotation + "</code>", "", 0},
		{"pre inert", htmlHeader + "<pre>" + annotation + "</pre>", "", 0},
		{"textarea inert", htmlHeader + "<textarea>" + annotation + "</textarea>", "", 0},
		{"script inert", htmlHeader + "<script>" + annotation + "</script>", "", 0},
		{"style inert", htmlHeader + "<style>" + annotation + "</style>", "", 0},
		{"body marker", "<!doctype html><html><head></head><body><!-- uicl:html 1.0 -->", "HOST_MARKER_OUTSIDE_HEAD", 0},
		{"fake marker", "<html><head><script>\"<!-- uicl:html 1.0 -->\"</script></head><body>", "HOST_MARKER_COUNT", 0},
		{"duplicate marker", htmlHeader + "<!-- uicl:html 1.0 -->", "HOST_MARKER_COUNT", 0},
		{"wrong version", strings.Replace(htmlHeader, "1.0", "2.0", 1), "HOST_MODE_OR_VERSION_CONFLICT", 0},
		{"wrong encoding", htmlHeader + strings.Replace(data, "data-uicl", "data-uicl-encoding=other", 1), "HOST_ENCODING", 0},
		{"duplicate attr", htmlHeader + strings.Replace(data, "data-uicl", "data-uicl data-uicl", 1), "HOST_ATTRIBUTE_DUPLICATE", 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			d := Parse("test.html", []byte(tt.source), ParseOptions{Recover: true})
			if tt.code == "" && HasErrors(d.Diagnostics) {
				t.Fatalf("unexpected errors: %+v", d.Diagnostics)
			}
			if tt.code != "" && !hostCode(d, tt.code) {
				t.Fatalf("want %s got %+v", tt.code, d.Diagnostics)
			}
			if len(d.Islands) != tt.count {
				t.Fatalf("want %d islands got %d", tt.count, len(d.Islands))
			}
		})
	}
}

func TestHostEncodedAndSourceMaps(t *testing.T) {
	payload := "thing #one\n  title: \"中文🙂\"\n"
	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	d := mustHost(t, mdMarker+"<!-- uicl:base64url\n"+encoded+"\n-->\n")
	if len(d.Nodes) != 1 || d.Nodes[0].Range != d.Islands[0].Range {
		t.Fatal("encoded node not mapped to carrier")
	}
	ds := []Diagnostic{{File: d.File, Range: d.Islands[0].Range, Path: "thing#one.title"}, {File: d.File, Range: d.Islands[0].Range, Path: "unknown"}}
	d.AttachDecodedRanges(ds)
	if ds[0].DecodedRange == nil || *ds[0].DecodedRange != d.Islands[0].Document.Nodes[0].Get("title").Range {
		t.Fatal("downstream diagnostic lost decoded field range")
	}
	if ds[1].DecodedRange == nil || ds[1].DecodedRange.Start.Offset != 0 || ds[1].DecodedRange.End.Offset != len(payload) {
		t.Fatal("ambiguous downstream diagnostic must cover decoded payload")
	}
	for _, bad := range []string{"Zh", "Zg==", "A", "YWJj$", base64.RawURLEncoding.EncodeToString([]byte{0xff})} {
		t.Run(bad, func(t *testing.T) {
			d := Parse("bad.md", []byte(mdMarker+"<!-- uicl:base64url\n"+bad+"\n-->\n"), ParseOptions{})
			if !hostCode(d, "HOST_ENCODING") {
				t.Fatalf("expected strict encoding error: %+v", d.Diagnostics)
			}
		})
	}
	errPayload := base64.RawURLEncoding.EncodeToString([]byte("thing #a\n  title: [\n"))
	broken := Parse("bad.md", []byte(mdMarker+"<!-- uicl:base64url\n"+errPayload+"\n-->\n"), ParseOptions{Recover: true})
	if len(broken.Diagnostics) == 0 || broken.Diagnostics[0].DecodedRange == nil || broken.Diagnostics[0].Range != broken.Islands[0].Range {
		t.Fatalf("missing decoded diagnostic position: %+v", broken.Diagnostics)
	}
	plain := "\xef\xbb\xbf" + strings.ReplaceAll(mdMarker+"<!-- uicl\n"+payload+"-->\n", "\n", "\r\n")
	d = mustHost(t, plain)
	node := d.Nodes[0]
	if string(d.Source[node.KindRange.Start.Offset:node.KindRange.End.Offset]) != "thing" {
		t.Fatal("node range mapping")
	}
	value := node.Get("title")
	if string(d.Source[value.Range.Start.Offset:value.Range.End.Offset]) != "\"中文🙂\"" {
		t.Fatal("value range mapping")
	}
	if value.Range.End.Column-value.Range.Start.Column != 5 {
		t.Fatalf("Unicode scalar column mismatch: %+v", value.Range)
	}
}

func FuzzHostParsing(f *testing.F) {
	f.Add(mdMarker + "<!-- uicl\nthing #a\n-->\n")
	f.Add(htmlHeader + "<script type=text/plain data-uicl>uicl \"1.0\"\nthing</script>")
	f.Fuzz(func(t *testing.T, source string) {
		if len(source) > 65536 {
			t.Skip()
		}
		d := Parse("fuzz.uicl", []byte(source), ParseOptions{Recover: true})
		for _, diag := range d.Diagnostics {
			if diag.Range.Start.Offset < 0 || diag.Range.End.Offset > len(source) || diag.Range.End.Offset < diag.Range.Start.Offset {
				t.Fatalf("invalid diagnostic range: %+v", diag)
			}
		}
	})
}
