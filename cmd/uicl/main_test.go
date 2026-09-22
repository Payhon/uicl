package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func call(t *testing.T, input string, args ...string) (int, string, string) {
	t.Helper()
	var out, err bytes.Buffer
	code := run(args, strings.NewReader(input), &out, &err)
	return code, out.String(), err.String()
}
func TestCLIJSONAndExitCodes(t *testing.T) {
	root := t.TempDir()
	name := filepath.Join(root, "draft.uicl")
	cases := []struct {
		source string
		code   int
	}{{"uicl \"1.0\"\napp \"你好\"\n", 0}, {"uicl \"1.0\"\nunknown\n", 1}, {"uicl \"1.0\"\napp (title: \"x\",)\n", 1}}
	for _, tc := range cases {
		code, out, err := call(t, tc.source, "check", "--root", root, "-", "--filename", name, "--json")
		if code != tc.code {
			t.Fatalf("code %d want %d: %s %s", code, tc.code, out, err)
		}
		var v map[string]any
		if e := json.Unmarshal([]byte(out), &v); e != nil || v["schemaVersion"] != float64(1) {
			t.Fatalf("not one versioned JSON object: %s (%v)", out, e)
		}
	}
	code, out, _ := call(t, "", "check", "--root", root, "--json", filepath.Join(root, "missing.uicl"))
	if code != 2 || !json.Valid([]byte(out)) {
		t.Fatalf("missing file: %d %s", code, out)
	}
	code, out, _ = call(t, "", "check", "--json")
	if code != 2 || !json.Valid([]byte(out)) {
		t.Fatal(code, out)
	}
}
func TestCLIFormatPreflight(t *testing.T) {
	root := t.TempDir()
	a, b := filepath.Join(root, "a.uicl"), filepath.Join(root, "b.uicl")
	original := "uicl \"1.0\"\napp   \"x\"\n"
	os.WriteFile(a, []byte(original), 0600)
	os.WriteFile(b, []byte("broken"), 0600)
	code, _, _ := call(t, "", "fmt", "--root", root, "--write", a, b)
	after, _ := os.ReadFile(a)
	if code != 1 || string(after) != original {
		t.Fatal("partial write despite preflight failure")
	}
	code, _, _ = call(t, "", "fmt", "--root", root, "--check", a)
	if code != 1 {
		t.Fatal(code)
	}
	code, out, err := call(t, "", "fmt", "--root", root, "--write", a, "--json")
	if code != 0 || !json.Valid([]byte(out)) {
		t.Fatal(code, out, err)
	}
	code, _, _ = call(t, "", "fmt", "--root", root, "--check", a)
	if code != 0 {
		t.Fatal(code)
	}
}
func TestCLIInspection(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{{"version", "--root", root, "--json"}, {"profile", "inspect", "uicl.ui", "--json", "--root", root}, {"explain", "PRIMARY_CONFLICT", "--json"}} {
		code, out, err := call(t, "", args...)
		if code != 0 || !json.Valid([]byte(out)) {
			t.Fatal(args, code, out, err)
		}
	}
	code, _, _ := call(t, "", "explain", "NONEXISTENT")
	if code != 2 {
		t.Fatal(code)
	}
}
