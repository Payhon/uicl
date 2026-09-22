package uicl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Payhon/uicl/profiles"
)

func catalogForTest(t *testing.T) *Catalog {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	c, ds, err := LoadCatalog(root, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if HasErrors(ds) {
		t.Fatalf("standard Profile errors: %+v", ds)
	}
	return c
}

func catalogHasCode(ds []Diagnostic, code string) bool {
	for _, d := range ds {
		if d.Code == code {
			return true
		}
	}
	return false
}

func TestCatalogStandardAndExamples(t *testing.T) {
	c := catalogForTest(t)
	if len(c.Profiles) != 20 || len(c.Schemas) != 162 || len(c.Records) != 20 {
		t.Fatalf("counts: %d profiles, %d schemas, %d records", len(c.Profiles), len(c.Schemas), len(c.Records))
	}
	for _, glob := range []string{"examples/*.uicl", "examples/fullstack/*.uicl", "meta/*.uicl"} {
		files, err := filepath.Glob(glob)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			t.Run(file, func(t *testing.T) {
				source, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				d := ParseCore(file, source, ParseOptions{})
				if HasErrors(d.Diagnostics) {
					t.Fatalf("parse: %+v", d.Diagnostics)
				}
				if ds := c.Check(d); HasErrors(ds) {
					t.Fatalf("shape: %+v", ds)
				}
			})
		}
	}
}

func TestCatalogShapeAndStaticRules(t *testing.T) {
	c := catalogForTest(t)
	cases := []struct{ name, body, code string }{
		{"primary", "ui.text \"你好\"\n", ""},
		{"primary conflict", "ui.text \"你好\"\n  value: \"duplicate\"\n", "PRIMARY_CONFLICT"},
		{"missing id", "state\n  value: 0\n", "MISSING_ID"},
		{"duplicate id", "state #x\n  value: 0\nstate #x\n  value: 1\n", "DUPLICATE_ID"},
		{"unknown field", "ui.text \"x\"\n  typo: true\n", "UNKNOWN_PROPERTY"},
		{"unknown node", "mystery \"x\"\n", "UNKNOWN_NODE"},
		{"required", "state #x\n", "MISSING_PROPERTY"},
		{"shape", "ui.text \"x\"\n  size: true\n", "TYPE_SHAPE"},
		{"enum", "secret #x\n  source: imaginary\n  scope: [ssh]\n", "ENUM"},
		{"expression", "secret #x\n  source: $mode\n  scope: [ssh]\n", "EXPRESSION_FORBIDDEN"},
		{"child", "ui.text \"x\"\n  ui.text \"y\"\n", "CHILD_FORBIDDEN"},
		{"file missing source", "file #x\n  media: \"text/plain\"\n", "MUTUALLY_EXCLUSIVE"},
		{"file source conflict", "file #x\n  media: \"text/plain\"\n  source: @\"x\"\n  body: text\n", "MUTUALLY_EXCLUSIVE"},
		{"empty type", "type #x\n", "TYPE_RECORD_OR_ENUM"},
		{"server access", "collection #x\n  storage: server\n  field \"name\"\n    type: text\n", "SERVER_ACCESS_REQUIRED"},
		{"secret scope", "secret #x\n  source: runtime_prompt\n  scope: []\n", "SECRET_EMPTY_SCOPE"},
		{"image dimensions", "image #x\n  prompt: test\n  size: [0, 1200]\n", "DIMENSION_PAIR"},
		{"grammar repeat", "g.repeat\n  min: 2\n  max: 1\n  g.literal\n    text: x\n", "GRAMMAR_REPEAT_BOUNDS"},
		{"malformed record", "video #x\n  prompt: test\n  size: [100, 100]\n  fps: {num: nope}\n  frames: 24\n", "TYPE_SHAPE"},
		{"zero fps", "video #x\n  size: [100, 100]\n  fps: {num: 24, den: 0}\n  frames: 24\n", "VIDEO_TIMEBASE"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			d := ParseCore("test.uicl", []byte("uicl \"1.0\"\n"+tt.body), ParseOptions{})
			if HasErrors(d.Diagnostics) {
				t.Fatalf("parse: %+v", d.Diagnostics)
			}
			ds := c.Check(d)
			if tt.code == "" && HasErrors(ds) || tt.code != "" && !catalogHasCode(ds, tt.code) {
				t.Fatalf("expected %q, got %+v", tt.code, ds)
			}
		})
	}
	// Schema normalization must not change the source AST or synthesize defaults.
	d := ParseCore("test.uicl", []byte("uicl \"1.0\"\nui.text \"你好\"\n"), ParseOptions{})
	n := d.Nodes[0]
	_ = c.Check(d)
	if n.Label == nil || len(n.Props) != 0 || len(c.Properties(n)) != 1 {
		t.Fatal("primary normalization mutated AST")
	}
}

func TestCatalogTimelineAndDatabase(t *testing.T) {
	c := catalogForTest(t)
	cases := []struct{ file, old, replacement, code string }{
		{"examples/14-aigc-video.uicl", "start_frame: 72", "start_frame: 20", "TIMELINE_OVERLAP"},
		{"examples/14-aigc-video.uicl", "start_frame: 72", "start_frame: 9223372036854775807", "TIMELINE_BOUNDS"},
		{"examples/09-postgresql.uicl", "atomic: split_allowed", "atomic: required", "PG_TRANSACTION_CONFLICT"},
		{"examples/09-postgresql.uicl", "columns: [customer_id]", "columns: [missing]", "PG_UNKNOWN_COLUMN"},
	}
	for _, tt := range cases {
		t.Run(tt.code, func(t *testing.T) {
			source, err := os.ReadFile(tt.file)
			if err != nil {
				t.Fatal(err)
			}
			text := strings.Replace(string(source), tt.old, tt.replacement, 1)
			if text == string(source) {
				t.Fatalf("fixture no longer contains %q", tt.old)
			}
			d := ParseCore(tt.file, []byte(text), ParseOptions{})
			if HasErrors(d.Diagnostics) {
				t.Fatalf("parse: %+v", d.Diagnostics)
			}
			if ds := c.Check(d); !catalogHasCode(ds, tt.code) {
				t.Fatalf("expected %s: %+v", tt.code, ds)
			}
		})
	}
}

func TestCatalogExtensionsAndInvalidDeclarations(t *testing.T) {
	source, err := os.ReadFile("examples/profile-authoring/measurement-profile.uicl")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, old, replacement, code string }{
		{"valid", "", "", ""},
		{"unknown type", "type: \"decimal\"", "type: \"banana\"", "UNKNOWN_SHAPE_TYPE"},
		{"malformed type", "type: \"decimal\"", "type: \"list<decimal\"", "UNKNOWN_SHAPE_TYPE"},
		{"missing record", "type: \"decimal\"", "type: \"record<Missing>\"", "UNKNOWN_SHAPE_TYPE"},
		{"missing dependency", "uicl.core@1.0.0", "missing@1.0.0", "PROFILE_DEPENDENCY"},
		{"version mismatch", "uicl.core@1.0.0", "uicl.core@9.0.0", "PROFILE_DEPENDENCY"},
		{"cycle", "uicl.core@1.0.0", "example.measurement@1.0.0", "PROFILE_CYCLE"},
		{"unsupported core", "core: \"1.0\"", "core: \"2.0\"", "PROFILE_CORE"},
		{"primary missing", "primary: \"label\"", "primary: \"missing\"", "SCHEMA_PRIMARY"},
		{"duplicate field", "property \"unit\"", "property \"value\"", "DUPLICATE_SCHEMA_PROPERTY"},
		{"duplicate port", "  rule #measurement_rule", "    port \"value\"\n      type: \"decimal\"\n      phase: \"compile\"\n  rule #measurement_rule", "DUPLICATE_SCHEMA_PORT"},
		{"invalid required flag", "required: true", "required: nope", "TYPE_SHAPE"},
		{"malformed targets", "      type: \"decimal\"", "      targets: 4\n      type: \"decimal\"", "TYPE_SHAPE"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			file := filepath.Join(root, "extension.uicl")
			text := string(source)
			if tt.old != "" {
				text = strings.Replace(text, tt.old, tt.replacement, 1)
			}
			if err := os.WriteFile(file, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			c, ds, err := LoadCatalog(root, "", []string{file})
			if err != nil {
				t.Fatal(err)
			}
			if tt.code == "" {
				if HasErrors(ds) {
					t.Fatalf("extension: %+v", ds)
				}
				if c.Schemas["lab.sample"] == nil {
					t.Fatal("extension not registered")
				}
			} else if !catalogHasCode(ds, tt.code) {
				t.Fatalf("expected %s, got %+v", tt.code, ds)
			}
		})
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	c, ds, err := LoadCatalog(root, "", []string{"examples/profile-authoring/measurement-profile.uicl"})
	if err != nil || HasErrors(ds) {
		t.Fatalf("extension: %v %+v", err, ds)
	}
	example, err := os.ReadFile("examples/profile-authoring/measurement-example.uicl")
	if err != nil {
		t.Fatal(err)
	}
	d := ParseCore("measurement-example.uicl", example, ParseOptions{})
	if ds := c.Check(d); HasErrors(ds) {
		t.Fatalf("example: %+v", ds)
	}
	_, ds, err = LoadCatalog(root, "", []string{"profiles/00-meta.uicl"})
	if err != nil || !catalogHasCode(ds, "PROFILE_CONFLICT") {
		t.Fatalf("duplicate: %v %+v", err, ds)
	}
}

func TestCatalogExplicitSourcesReplaceEmbedded(t *testing.T) {
	root := t.TempDir()
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "profiles")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	entries, err := profiles.Files.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		data, err := profiles.Files.ReadFile(entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		// A harmless source change must be visible in inspect's digest and source.
		data = append(data, []byte("\n// explicit catalog source\n")...)
		if err := os.WriteFile(filepath.Join(dir, entry.Name()), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	c, ds, err := LoadCatalog(root, dir, nil)
	if err != nil || HasErrors(ds) {
		t.Fatalf("explicit: %v %+v", err, ds)
	}
	for _, p := range c.Profiles {
		if !strings.HasPrefix(p.Path, root) || !strings.Contains(p.Source, "explicit catalog source") || p.Digest != Digest([]byte(p.Source)) {
			t.Fatalf("not explicit: %+v", p)
		}
	}
	// The embedded catalog works from an otherwise empty directory.
	if c, ds, err := LoadCatalog(t.TempDir(), "", nil); err != nil || HasErrors(ds) || len(c.Profiles) != 20 {
		t.Fatalf("embedded: %v %+v", err, ds)
	}
	// Replacement metadata cannot mark its own mandatory fields as optional.
	metaFile := filepath.Join(dir, "00-meta.uicl")
	meta, err := os.ReadFile(metaFile)
	if err != nil {
		t.Fatal(err)
	}
	meta = []byte(strings.ReplaceAll(string(meta), "      required: true\n", ""))
	if err := os.WriteFile(metaFile, meta, 0600); err != nil {
		t.Fatal(err)
	}
	_, ds, err = LoadCatalog(root, dir, nil)
	if err != nil || !catalogHasCode(ds, "MISSING_PROPERTY") {
		t.Fatalf("metamodel bypass: %v %+v", err, ds)
	}
}

func TestCatalogNestedReferences(t *testing.T) {
	c := catalogForTest(t)
	d := ParseCore("nested.uicl", []byte("uicl \"1.0\"\npg.table #a\n  foreign_keys: [{columns: [id], target: @b, target_columns: [id], on_delete: restrict}]\n"), ParseOptions{})
	if HasErrors(d.Diagnostics) {
		t.Fatalf("parse: %+v", d.Diagnostics)
	}
	got := c.References(d.Nodes[0].Get("foreign_keys"), c.Schemas["pg.table"].Fields["foreign_keys"])
	if len(got) != 1 || got[0].Value.ID != "b" || len(got[0].Targets) != 1 || got[0].Targets[0] != "pg.table" {
		t.Fatalf("nested reference constraints: %+v", got)
	}
	// Malformed typed data still returns references for independent diagnostics.
	got = c.References(&Value{Tag: "ref", ID: "x"}, &FieldSpec{Type: "record<Missing>"})
	if len(got) != 1 || got[0].Value.ID != "x" {
		t.Fatalf("malformed reference: %+v", got)
	}
}

func TestCatalogEnumValueEquality(t *testing.T) {
	for _, pair := range [][2]string{{"1.0", "1.00"}, {"-0", "0"}, {"1e3", "1000.0"}, {"1e-6176", "10e-6177"}} {
		a := &Value{Tag: "literal", Type: "decimal", Text: pair[0]}
		b := &Value{Tag: "literal", Type: "decimal", Text: pair[1], Range: Range{Start: Position{Offset: 9}}}
		if !enumContains([]*Value{a}, b) {
			t.Fatalf("equivalent decimal values differ: %v", pair)
		}
	}
	a := &Value{Tag: "ref", ID: "x"}
	b := &Value{Tag: "ref", ID: "x", Range: Range{Start: Position{Offset: 9}}}
	if !enumContains([]*Value{a}, b) {
		t.Fatal("reference ranges affect equality")
	}
	x := &Value{Tag: "literal", Type: "int", Text: "1"}
	y := &Value{Tag: "literal", Type: "int", Text: "2"}
	left := &Value{Tag: "record", Fields: []*Property{{Name: "a", Value: x}, {Name: "b", Value: y}}}
	right := &Value{Tag: "record", Fields: []*Property{{Name: "b", Value: y}, {Name: "a", Value: x}}}
	if !enumContains([]*Value{left}, right) || enumContains([]*Value{x}, y) {
		t.Fatal("record or distinct number equality")
	}
}
