package uicl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testCatalog(t *testing.T) *Catalog {
	t.Helper()
	c, ds, e := LoadCatalog(".", "", nil)
	if e != nil || HasErrors(ds) {
		t.Fatalf("catalog: %v %v", e, ds)
	}
	return c
}
func hasCode(ds []Diagnostic, code string) bool {
	for _, d := range ds {
		if d.Code == code {
			return true
		}
	}
	return false
}
func writeFixture(t *testing.T, root, name, body string) string {
	t.Helper()
	p := filepath.Join(root, name)
	if e := os.WriteFile(p, []byte("uicl \"1.0\"\n"+body), 0600); e != nil {
		t.Fatal(e)
	}
	return p
}

func TestWorkspaceLinks(t *testing.T) {
	cat := testCatalog(t)
	cases := []struct{ name, lib, main, code string }{
		{"export", "module #m\n  exports: [x]\nstate #x\n  value: 0\n", "use #lib\n  source: @\"./lib.uicl\"\nset\n  target: @lib::x\n  value: 1\n", ""},
		{"private", "module #m\n  exports: []\nstate #x\n  value: 0\n", "use #lib\n  source: @\"./lib.uicl\"\nset\n  target: @lib::x\n  value: 1\n", "NOT_EXPORTED"},
		{"missing", "app \"x\"\n", "set\n  target: @missing\n  value: 1\n", "UNRESOLVED_REF"},
		{"port", "app \"x\"\n", "state #x\n  value: 0\nset\n  target: @x.nope\n  value: 1\n", "UNKNOWN_PORT"},
		{"cycle", "use #main\n  source: @\"./main.uicl\"\n", "use #lib\n  source: @\"./lib.uicl\"\n", "IMPORT_CYCLE"},
		{"network", "app \"x\"\n", "use #remote\n  source: @\"https://example.org/m.uicl\"\n", "IMPORT_NONLOCAL_NOT_SUPPORTED"},
		{"escape", "app \"x\"\n", "use #bad\n  source: @\"../outside.uicl\"\n", "PATH_ESCAPE"},
		{"uri inert", "app \"x\"\n", "file #asset\n  media: \"text/plain\"\n  source: @\"https://invalid.example/no-network\"\n", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeFixture(t, root, "lib.uicl", tc.lib)
			main := writeFixture(t, root, "main.uicl", tc.main)
			w, e := NewWorkspace(root, cat)
			if e != nil {
				t.Fatal(e)
			}
			r := w.Check(main)
			if tc.code == "" {
				if r.ExitCode() != 0 {
					t.Fatalf("%+v", r.Diagnostics)
				}
			} else if !hasCode(r.Diagnostics, tc.code) {
				t.Fatalf("missing %s: %+v", tc.code, r.Diagnostics)
			}
		})
	}
}
func TestWorkspaceFreshBuffersAndDisk(t *testing.T) {
	root := t.TempDir()
	lib := writeFixture(t, root, "lib.uicl", "module #m\n  exports: [x]\nstate #x\n  value: 1\n")
	main := writeFixture(t, root, "main.uicl", "use #lib\n  source: @\"./lib.uicl\"\nset\n  target: @lib::x\n  value: 2\n")
	w, _ := NewWorkspace(root, testCatalog(t))
	if r := w.Check(main, main); r.ExitCode() != 0 || r.Documents != 2 {
		t.Fatalf("%+v", r)
	}
	if e := w.SetDocument(lib, []byte("uicl \"1.0\"\nmodule #m\n  exports: []\n")); e != nil {
		t.Fatal(e)
	}
	if !hasCode(w.Check(main).Diagnostics, "NOT_EXPORTED") {
		t.Fatal("unsaved import ignored")
	}
	w.CloseDocument(lib)
	if r := w.Check(main); r.ExitCode() != 0 {
		t.Fatal(r.Diagnostics)
	}
	os.WriteFile(lib, []byte("uicl \"1.0\"\nmodule #m\n  exports: []\n"), 0600)
	if !hasCode(w.Check(main).Diagnostics, "NOT_EXPORTED") {
		t.Fatal("disk change ignored")
	}
	newFile := filepath.Join(root, "unsaved.uicl")
	w.SetDocument(newFile, []byte("uicl \"1.0\"\napp \"draft\"\n"))
	if r := w.Check(newFile); r.ExitCode() != 0 {
		t.Fatal(r.Diagnostics)
	}
}
func TestWorkspaceRepository(t *testing.T) {
	cat := testCatalog(t)
	w, e := NewWorkspace(".", cat)
	if e != nil {
		t.Fatal(e)
	}
	var files []string
	for _, pattern := range []string{"profiles/*.uicl", "meta/*.uicl", "examples/*.uicl", "examples/fullstack/*.uicl", "examples/hosted/*.uicl", "spec/*.uicl", "guides/*.uicl"} {
		p, _ := filepath.Glob(pattern)
		files = append(files, p...)
	}
	files = append(files, "README.uicl", "UICL-1.0-Structured.uicl")
	r := w.Check(files...)
	if r.ExitCode() != 0 {
		for _, d := range r.Diagnostics {
			t.Errorf("%s:%d %s %s", d.File, d.Range.Start.Line+1, d.Code, d.Message)
		}
	}
	if r.Documents < 90 {
		t.Fatalf("unexpectedly few docs: %d", r.Documents)
	}
}
func TestResolveSymlinkEscape(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	target := writeFixture(t, outside, "x.uicl", "app \"x\"\n")
	link := filepath.Join(root, "link.uicl")
	if e := os.Symlink(target, link); e != nil {
		t.Skip(e)
	}
	if _, e := ResolvePath(root, link); e == nil {
		t.Fatal("symlink escape allowed")
	}
}
func TestCheckSourceLock(t *testing.T) {
	if r := CheckLock(".", "uicl.lock.json"); r.ExitCode() != 0 {
		t.Fatal(r.Diagnostics)
	}
	root := t.TempDir()
	src := writeFixture(t, root, "p.uicl", "profile #p \"P\"\n  id: \"p\"\n  version: \"1.0.0\"\n  core: \"1.0\"\n  requires: []\n")
	data, _ := os.ReadFile(src)
	lock := sourceLock{Language: "uicl", Core: "1.0", Status: "package-source-lock-not-production-plan", Catalog: []lockEntry{{ID: "p", Version: "1.0.0", Source: "p.uicl", SHA256: Digest(data), Requires: []string{}}}}
	b, _ := json.Marshal(lock)
	p := filepath.Join(root, "uicl.lock.json")
	os.WriteFile(p, b, 0600)
	if r := CheckLock(root, p); r.ExitCode() != 0 {
		t.Fatal(r.Diagnostics)
	}
	os.WriteFile(src, append(data, '\n'), 0600)
	if !hasCode(CheckLock(root, p).Diagnostics, "LOCK_DIGEST") {
		t.Fatal("digest drift ignored")
	}
	os.Remove(src)
	if !hasCode(CheckLock(root, p).Diagnostics, "LOCK_SOURCE") {
		t.Fatal("missing source ignored")
	}
	after, _ := os.ReadFile(p)
	if string(after) != string(b) {
		t.Fatal("lock modified")
	}
}
func TestWriteSourceConflictAndPermissions(t *testing.T) {
	root := t.TempDir()
	p := writeFixture(t, root, "x.uicl", "app \"x\"\n")
	b, _ := os.ReadFile(p)
	before, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if e := WriteSource(p, Digest([]byte("stale")), []byte("wrong")); e == nil || !strings.Contains(e.Error(), "CONFLICT") {
		t.Fatal(e)
	}
	if e := WriteSource(p, Digest(b), []byte("replacement")); e != nil {
		t.Fatal(e)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != before.Mode().Perm() {
		t.Fatalf("permissions changed: got %v, want %v", st.Mode().Perm(), before.Mode().Perm())
	}
	matches, _ := filepath.Glob(filepath.Join(root, ".uicl-*"))
	if len(matches) != 0 {
		t.Fatal("temporary files leaked")
	}
}

func TestCloseRemovedSymlinkBuffer(t *testing.T) {
	root := t.TempDir()
	target := writeFixture(t, root, "x.uicl", "app \"valid\"\n")
	link := filepath.Join(root, "link.uicl")
	if e := os.Symlink(target, link); e != nil {
		t.Skip(e)
	}
	w, _ := NewWorkspace(root, testCatalog(t))
	if e := w.SetDocument(link, []byte("uicl \"1.0\"\nunknown\n")); e != nil {
		t.Fatal(e)
	}
	if !hasCode(w.Check(target).Diagnostics, "UNKNOWN_NODE") {
		t.Fatal("buffer not applied")
	}
	os.Remove(link)
	if e := w.CloseDocument(link); e != nil {
		t.Fatal(e)
	}
	if r := w.Check(target); r.ExitCode() != 0 {
		t.Fatalf("stale alias buffer: %v", r.Diagnostics)
	}
}

func TestReportDoesNotPassSkippedStages(t *testing.T) {
	root := t.TempDir()
	p := writeFixture(t, root, "bad.uicl", "app (title: \"x\",)\n")
	w, _ := NewWorkspace(root, testCatalog(t))
	r := w.Check(p)
	if r.Stages["syntax"] != "FAIL" || r.Stages["shape"] != "NOT_EVALUATED" || r.Stages["link"] != "NOT_EVALUATED" {
		t.Fatal(r.Stages)
	}
	r = CheckLock(filepath.Join(root, "missing"), "uicl.lock.json")
	if r.ExitCode() != 2 || r.Stages["lock"] != "ERROR" {
		t.Fatal(r)
	}
}
