package uicl

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func parserValue(t *testing.T, text string) *Value {
	t.Helper()
	d := ParseCore("test.uicl", []byte("uicl \"1.0\"\nnode\n  value: "+text+"\n"), ParseOptions{})
	if HasErrors(d.Diagnostics) {
		t.Fatalf("parse %q: %+v", text, d.Diagnostics)
	}
	return d.Nodes[0].Get("value")
}
func TestParserCollections(t *testing.T) {
	for _, pair := range [][2]string{
		{"[read, create]", "\n    - read\n    - create"},
		{"[{name: \"a\", value: 1}, {name: \"b\", value: 2}]", "\n    - name: \"a\"\n      value: 1\n    - name: \"b\"\n      value: 2"},
		{"{nested: {values: [1, 2]}}", "\n    nested:\n      values:\n        - 1\n        - 2"},
		{"[[1, 2], [3, 4]]", "\n    -\n      - 1\n      - 2\n    - [3, 4]"},
		{"[{nested: [1], next: true}]", "\n    - nested:\n        - 1\n      next: true"},
		{"[-10, 0]", "\n    - -10\n    - 0"},
	} {
		if a, b := Plain(parserValue(t, pair[0])), Plain(parserValue(t, pair[1])); !reflect.DeepEqual(a, b) {
			t.Fatalf("%s != %s: %#v %#v", pair[0], pair[1], a, b)
		}
	}
}
func TestParserRaw(t *testing.T) {
	for _, tt := range []struct{ source, want string }{
		{"|\n    @ref $data #id\n    - no", "@ref $data #id\n- no\n"},
		{"```python\n    print(\"@x\")\n  ```", "print(\"@x\")\n"},
		{"|\n    \tstill content\n\n", "\tstill content\n"},
		{"````go\n    ```\n    hello\n  ````", "```\nhello\n"},
	} {
		v := parserValue(t, tt.source)
		if v.Text != tt.want {
			t.Fatalf("raw %q: %q != %q", tt.source, v.Text, tt.want)
		}
	}
	v := parserValue(t, "\n    - |\n      first\n      second\n    - \"third\"")
	if got := Plain(v); !reflect.DeepEqual(got, []any{"first\nsecond\n", "third"}) {
		t.Fatal(got)
	}
}
func TestParserExpressions(t *testing.T) {
	for _, s := range []string{"= 1 + (not true)", "= 1 == (not false)", "= -(not true)", "= not not true"} {
		parserValue(t, s)
	}
	if v := parserValue(t, "= $x + 2 * 3").Tree; v.Op != "+" || v.Right.Op != "*" {
		t.Fatal(v)
	}
	if v := parserValue(t, "= not $a == $b").Tree; v.Op != "not" || v.Right.Op != "==" {
		t.Fatal(v)
	}
	if v := parserValue(t, "= $a ?? $b ?? $c").Tree; v.Op != "??" || v.Right.Op != "??" {
		t.Fatal(v)
	}
	if v := parserValue(t, "= not $a and $b").Tree; v.Op != "and" || v.Left.Op != "not" {
		t.Fatal(v)
	}
	if v := parserValue(t, "= ($a < $b) and ($b < $c)").Tree; v.Op != "and" {
		t.Fatal(v)
	}
	if v := parserValue(t, "= len([$item.name, $other?.value[0]])").Tree; v.Tag != "pure_call" || len(v.Args) != 1 {
		t.Fatal(v)
	}
	if v := parserValue(t, "$item?.name[1]"); v.Tag != "expression" || v.Tree.Tag != "index" || !v.Tree.Object.Optional {
		t.Fatal(v)
	}
	if v := parserValue(t, "@lib::node.output"); v.Module != "lib" || v.ID != "node" || len(v.Ports) != 1 {
		t.Fatal(v)
	}
	if v := parserValue(t, "@\"https://example.org/a\""); v.URI == nil || *v.URI != "https://example.org/a" {
		t.Fatal(v)
	}
	if v := parserValue(t, "\"@x\""); v.Tag != "literal" {
		t.Fatal(v)
	}
}
func TestParserNumbers(t *testing.T) {
	for _, s := range []string{"0", "-0", "9223372036854775807", "-9223372036854775808", "1e3", "1.234567890123456789012345678901234", "1e-6176", "1e6111", "1e6144", "100e-6178", "10000000000000000000000000000000000.0", "-0.00e99999999999999999999"} {
		v := parserValue(t, s)
		if v.Text != s {
			t.Fatalf("numeric lexeme changed: %q", v.Text)
		}
	}
	if v := parserValue(t, "= -9223372036854775808").Tree; v.Text != "-9223372036854775808" {
		t.Fatal(v)
	}
	if v := parserValue(t, "1e3"); v.Type != "decimal" {
		t.Fatal(v)
	}
	for _, s := range []string{"9223372036854775808", "-9223372036854775809", "1e-6177", "1e6145", "1.2345678901234567890123456789012345", "1e999999999999999999999", "1e-99999999999999999"} {
		d := ParseCore("number.uicl", []byte("uicl \"1.0\"\nn\n  x: "+s), ParseOptions{})
		if len(d.Diagnostics) != 1 || d.Diagnostics[0].Code != "UICL-NUMBER_RANGE" {
			t.Fatalf("expected numeric range error for %s: %+v", s, d.Diagnostics)
		}
	}
}
func TestParserLongExactCoefficient(t *testing.T) {
	// The value is exactly one despite its deliberately long lexical form.
	value := "1" + strings.Repeat("0", 1_000_001) + "e-1000001"
	d := ParseCore("large-number.uicl", []byte("uicl \"1.0\"\nn\n  x: "+value), ParseOptions{})
	if HasErrors(d.Diagnostics) {
		t.Fatal(d.Diagnostics)
	}
	if d.Nodes[0].Get("x").Text != value {
		t.Fatal("coefficient lexeme changed")
	}
	for _, value := range []string{"1e-9223372036854775808", "1e9223372036854775807", "1.0e-9223372036854775808"} {
		if decimalFits(value) {
			t.Errorf("overflow in exponent normalization: %s", value)
		}
	}
}
func TestParserInvalidSyntax(t *testing.T) {
	for _, body := range []string{
		"n\n  x: = $a < $b < $c", "n\n  x: = 1 / 3", "n\n  x: = status == \"active\"",
		"n\n  x: = 1 + not true", "n\n  x: = 1 == not false", "n\n  x: = -not true",
		"n\n  x: {a:1,a:2}", "n\n  x:\n    a: 1\n    a: 2", "n (x: 1)\n  x: 2",
		"n\n  x:", "n\n  x:\n    - 1\n    a: 2", "n\n  x:\n    name: a\n    ui.button x",
		"n\n    ui.text x", "n\n\tui.text x", "n\n  x: ```go\n    x = 1",
		"n\n  x: \"\\ud800\"", "n\n  x: \"\\udc00\"", "n\n  x: \"\\ud800\\u0041\"",
		"n\n  x: [1,]", "n\n  x: = len(1,)", "n\n  x: {a:1,}", "n (x:1,)",
		"n\n  x: 01", "n\n  x: 0x10", "n\n  x: NaN", "n\n  x: Infinity", "n\n  x: +1",
		"n#x", "n \"x\"(a:1)", "n\n  x: @ x", "n\n  x: @x .y", "n\n  x: $ x",
		"n\n  x:\n    - value\n      bad: field", "n\n  x:\n    -10", "n\n  x: [1,\n    2]",
		"n // end-of-line comment", "n\n  x: \"unterminated", "n\n  x: \"\\q\"",
	} {
		t.Run(body, func(t *testing.T) {
			d := ParseCore("bad.uicl", []byte("uicl \"1.0\"\n"+body+"\n"), ParseOptions{})
			if !HasErrors(d.Diagnostics) {
				t.Fatalf("accepted %q", body)
			}
			if len(d.Diagnostics) != 1 {
				t.Fatalf("strict mode should report once: %+v", d.Diagnostics)
			}
		})
	}
	for _, source := range []string{"aml \"1.0\"\nn", "uicl \"2.0\"\nn", "uicl\"1.0\"\nn", "uicl \"1.0\"", "uicl \"1.0\"\rn", "uicl \"1.0\"\nn\x00", string([]byte{0xff})} {
		d := ParseCore("bad.uicl", []byte(source), ParseOptions{})
		if !HasErrors(d.Diagnostics) {
			t.Errorf("accepted %q", source)
		}
	}
	if v := parserValue(t, "\"\\ud83d\\ude00\""); v.Text != "😀" {
		t.Fatal(v)
	}
}
func TestParserSourceRangesAndTokens(t *testing.T) {
	source := []byte("\xef\xbb\xbfuicl \"1.0\"\r\n// 注释\r\nnode #test \"中文😀\" (x: [1, 2])\r\n  text: ```go\r\n    \tprintln(1)\r\n  ```\r\n  next: @test\r\n")
	d := ParseCore("range.uicl", source, ParseOptions{})
	if HasErrors(d.Diagnostics) {
		t.Fatal(d.Diagnostics)
	}
	if !bytes.Equal(d.Source, source) {
		t.Fatal("source changed")
	}
	var reconstructed strings.Builder
	offset := 0
	raw := 0
	for _, token := range d.Tokens {
		if token.Range.Start.Offset != offset {
			t.Fatalf("token gap at %d: %+v", offset, token)
		}
		if string(source[token.Range.Start.Offset:token.Range.End.Offset]) != token.Text {
			t.Fatal("bad token text")
		}
		reconstructed.WriteString(token.Text)
		offset = token.Range.End.Offset
		if token.Kind == "RAW" {
			raw++
		}
	}
	if reconstructed.String() != string(source) || raw != 1 {
		t.Fatalf("incomplete token coverage %d", raw)
	}
	n := d.Nodes[0]
	if n.KindRange.Start.Line != 2 || n.KindRange.Start.Column != 0 || n.IDRange.Start.Column != 5 || n.Label.Range.Start.Column != 11 || n.Label.Range.End.Column != 16 {
		t.Fatalf("incorrect Unicode ranges: %+v %+v %+v", n.KindRange, n.IDRange, n.Label.Range)
	}
	if got := string(source[n.Get("x").Range.Start.Offset:n.Get("x").Range.End.Offset]); got != "[1, 2]" {
		t.Fatal(got)
	}
	if n.Range.End.Offset < n.Get("next").Range.End.Offset {
		t.Fatal("node excludes properties")
	}
	source[3] = 'X'
	if d.Source[3] != 'u' {
		t.Fatal("caller mutated document source")
	}
}
func TestParserRecovery(t *testing.T) {
	source := "uicl \"1.0\"\nnode #keep\n  bad: [1,]\n  good: yes\n  wrong: \"unfinished\n  child #ok\nother #last\n"
	d := ParseCore("editing.uicl", []byte(source), ParseOptions{Recover: true})
	if len(d.Diagnostics) != 2 || len(d.Nodes) != 2 || d.Nodes[0].Get("good") == nil || len(d.Nodes[0].Children) != 1 {
		t.Fatalf("recovery: %+v %#v", d.Diagnostics, d.Nodes)
	}
	d = ParseCore("editing.uicl", []byte("uicl \"1.0\"\nnode\n  code: ```go\n    pretend #inert\nother #safe\n"), ParseOptions{Recover: true})
	if len(d.Diagnostics) != 1 || len(d.Nodes) != 2 || d.Nodes[1].ID != "safe" || len(d.Nodes[0].Children) != 0 {
		t.Fatalf("raw recovery: %+v %#v", d.Diagnostics, d.Nodes)
	}
	for _, closing := range []string{"  ```\n", ""} {
		d = ParseCore("editing.uicl", []byte("uicl \"1.0\"\nnode #keep\n  code: ```go\n    valid raw line\n  node #leaked\n"+closing+"other #safe\n"), ParseOptions{Recover: true})
		if len(d.Diagnostics) != 1 || len(d.Nodes) != 2 || d.Nodes[1].ID != "safe" || len(d.Nodes[0].Children) != 0 {
			t.Fatalf("activated malformed raw payload: %+v %#v", d.Diagnostics, d.Nodes)
		}
	}
	d = ParseCore("editing.uicl", []byte("uicl \"1.0\"\n\n\tbad\n\nnode\n\n   bad\n  good: true\n"), ParseOptions{Recover: true})
	if len(d.Diagnostics) != 2 || len(d.Nodes) != 1 || d.Nodes[0].Get("good") == nil {
		t.Fatalf("layout recovery: %+v %#v", d.Diagnostics, d.Nodes)
	}
	fragment := ParseCore("fragment", []byte("ui.text \"hello\"\n"), ParseOptions{Fragment: true})
	if HasErrors(fragment.Diagnostics) || len(fragment.Nodes) != 1 {
		t.Fatal(fragment)
	}
}
func TestParserLimits(t *testing.T) {
	d := ParseCore("large", bytes.Repeat([]byte{' '}, MaxBytes+1), ParseOptions{})
	if len(d.Diagnostics) != 1 || d.Diagnostics[0].Code != "UICL-LIMIT" {
		t.Fatal(d.Diagnostics)
	}
	d = ParseCore("deep", []byte("uicl \"1.0\"\nn\n  x: "+strings.Repeat("[", MaxDepth+1)+"0"+strings.Repeat("]", MaxDepth+1)), ParseOptions{})
	if len(d.Diagnostics) != 1 || d.Diagnostics[0].Code != "UICL-LIMIT" {
		t.Fatal(d.Diagnostics)
	}
	for _, value := range []string{"$a" + strings.Repeat("?.b", MaxDepth+1), "= 1" + strings.Repeat(" + 1", MaxDepth+1)} {
		d = ParseCore("deep", []byte("uicl \"1.0\"\nn\n  x: "+value), ParseOptions{})
		if len(d.Diagnostics) != 1 || d.Diagnostics[0].Code != "UICL-LIMIT" {
			t.Fatal(d.Diagnostics)
		}
	}
}
func TestParserRepositoryCorpus(t *testing.T) {
	count := 0
	for _, dir := range []string{"profiles", "meta", "examples"} {
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".uicl") {
				return nil
			}
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			trimmed := bytes.TrimSpace(bytes.TrimPrefix(source, []byte{0xef, 0xbb, 0xbf}))
			if !bytes.HasPrefix(trimmed, []byte("uicl \"")) {
				return nil
			}
			count++
			d := ParseCore(path, source, ParseOptions{})
			if HasErrors(d.Diagnostics) {
				t.Errorf("%s: %+v", path, d.Diagnostics)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if count < 50 {
		t.Fatalf("unexpectedly small corpus: %d", count)
	}
}
func FuzzParseCore(f *testing.F) {
	for _, s := range []string{"uicl \"1.0\"\nnode\n  x: 1\n", "uicl \"1.0\"\nnode\n  x: ```go\n    x := 1\n  ```\n", "node\n  x: [1, {y: $a}]", "uicl \"1.0\"\n\n\tbad\n"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, source []byte) {
		if len(source) > 64*1024 {
			t.Skip()
		}
		d := ParseCore("fuzz", source, ParseOptions{Recover: true})
		for _, diag := range d.Diagnostics {
			if diag.Range.Start.Offset < 0 || diag.Range.End.Offset > len(source) {
				t.Fatal("invalid diagnostic range")
			}
		}
		if !HasErrors(d.Diagnostics) {
			var joined []byte
			for _, tok := range d.Tokens {
				joined = append(joined, []byte(tok.Text)...)
			}
			if !bytes.Equal(joined, source) {
				t.Fatal("lost source bytes")
			}
		}
	})
}
