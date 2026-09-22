package uicl

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ResolvePath confines reads (including symlink targets) to root. A nonexistent
// final component is allowed for unsaved documents; ReadSource still reports IO.
func ResolvePath(root, path string) (string, error) {
	r, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	r, err = filepath.EvalSymlinks(r)
	if err != nil {
		return "", err
	}
	p, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var missing []string
	for {
		_, e := os.Lstat(p)
		if e == nil {
			p, err = filepath.EvalSymlinks(p)
			if err != nil {
				return "", err
			}
			break
		}
		if !errors.Is(e, fs.ErrNotExist) {
			return "", e
		}
		parent := filepath.Dir(p)
		if parent == p {
			return "", e
		}
		missing = append(missing, filepath.Base(p))
		p = parent
	}
	for i := len(missing) - 1; i >= 0; i-- {
		p = filepath.Join(p, missing[i])
	}
	rel, err := filepath.Rel(r, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("PATH_ESCAPE: %s is outside workspace", path)
	}
	return p, nil
}

func ReadSource(path string) ([]byte, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("source must be a regular file: %s", path)
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, MaxBytes+1))
	if e != nil {
		return nil, e
	}
	if len(b) > MaxBytes {
		return nil, fmt.Errorf("LIMIT: source exceeds %d bytes", MaxBytes)
	}
	return b, nil
}

type ExternalResource struct {
	File   string `json:"file"`
	URI    string `json:"uri"`
	Status string `json:"status"`
}
type Report struct {
	SchemaVersion     int                `json:"schemaVersion"`
	Status            string             `json:"status"`
	Stages            map[string]string  `json:"stages"`
	Diagnostics       []Diagnostic       `json:"diagnostics"`
	Documents         int                `json:"documents"`
	Profiles          int                `json:"profiles"`
	NodeSchemas       int                `json:"nodeSchemas"`
	RecordSchemas     int                `json:"recordSchemas"`
	References        int                `json:"references"`
	ExternalResources []ExternalResource `json:"externalResources"`
	NotEvaluated      []string           `json:"notEvaluated"`
}

func NewReport() Report {
	return Report{SchemaVersion: 1, Status: "PASS", Stages: map[string]string{"syntax": "NOT_EVALUATED", "shape": "NOT_EVALUATED", "link": "NOT_EVALUATED", "selectedStaticRules": "NOT_EVALUATED"}, Diagnostics: []Diagnostic{}, ExternalResources: []ExternalResource{}, NotEvaluated: []string{"runtime binding types and generic port compatibility", "full domain semantics, permissions and effects", "recipe expansion and execution", "real build, deployment and artifact verification", "browser rendering and complete host conformance"}}
}
func (r *Report) Finish() {
	SortDiagnostics(r.Diagnostics)
	r.Status = "PASS"
	for _, d := range r.Diagnostics {
		if d.Severity != "error" {
			continue
		}
		if r.Status != "ERROR" {
			r.Status = "FAIL"
		}
		if d.Phase == "io" || d.Phase == "tool" {
			r.Status = "ERROR"
		}
		phase := d.Phase
		if phase == "static" {
			phase = "selectedStaticRules"
		}
		if _, ok := r.Stages[phase]; ok {
			r.Stages[phase] = "FAIL"
		}
	}
	// A skipped prerequisite cannot make subsequent stages pass globally.
	if r.Status == "ERROR" || r.Stages["syntax"] == "FAIL" {
		for _, phase := range []string{"shape", "link", "selectedStaticRules"} {
			if r.Stages[phase] == "PASS" {
				r.Stages[phase] = "NOT_EVALUATED"
			}
		}
	}
	if r.Status == "ERROR" && r.Stages["syntax"] == "PASS" {
		r.Stages["syntax"] = "NOT_EVALUATED"
	}
	if r.Stages["shape"] == "FAIL" {
		for _, phase := range []string{"link", "selectedStaticRules"} {
			if r.Stages[phase] == "PASS" {
				r.Stages[phase] = "NOT_EVALUATED"
			}
		}
	}
	if _, ok := r.Stages["lock"]; ok && r.Status != "PASS" {
		r.Stages["lock"] = r.Status
	}
}
func (r Report) ExitCode() int {
	if r.Status == "ERROR" {
		return 2
	}
	if r.Status == "FAIL" {
		return 1
	}
	return 0
}

// Workspace keeps only explicit in-memory overrides. Each Check rebuilds its
// dependency graph so disk changes and removed imports cannot leave stale data.
type Workspace struct {
	Root      string
	Catalog   *Catalog
	Documents map[string][]byte
	aliases   map[string]string
}

func NewWorkspace(root string, catalog *Catalog) (*Workspace, error) {
	p, e := filepath.Abs(root)
	if e != nil {
		return nil, e
	}
	p, e = filepath.EvalSymlinks(p)
	if e != nil {
		return nil, e
	}
	st, e := os.Stat(p)
	if e != nil {
		return nil, e
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("workspace root is not a directory")
	}
	return &Workspace{Root: p, Catalog: catalog, Documents: map[string][]byte{}, aliases: map[string]string{}}, nil
}
func (w *Workspace) SetDocument(file string, source []byte) error {
	alias, e := filepath.Abs(file)
	if e != nil {
		return e
	}
	p, e := ResolvePath(w.Root, file)
	if e != nil {
		return e
	}
	if len(source) > MaxBytes {
		return fmt.Errorf("LIMIT: document too large")
	}
	if w.aliases == nil {
		w.aliases = map[string]string{}
	}
	if old := w.aliases[alias]; old != "" && old != p {
		delete(w.Documents, old)
	}
	w.aliases[alias] = p
	w.Documents[p] = append([]byte(nil), source...)
	return nil
}
func (w *Workspace) CloseDocument(file string) error {
	alias, e := filepath.Abs(file)
	if e != nil {
		return e
	}
	if p, ok := w.aliases[alias]; ok {
		delete(w.Documents, p)
		for name, target := range w.aliases {
			if target == p {
				delete(w.aliases, name)
			}
		}
		return nil
	}
	p, e := ResolvePath(w.Root, file)
	if e == nil {
		delete(w.Documents, p)
	}
	return e
}

type linkedDocument struct {
	doc     *Document
	ids     map[string]*Node
	exports map[string]bool
	imports map[string]*linkedDocument
}

func (w *Workspace) Check(files ...string) Report {
	r := NewReport()
	if w.Catalog == nil {
		r.Diagnostics = append(r.Diagnostics, Diagnostic{Code: "CATALOG_REQUIRED", Phase: "tool", Severity: "error", Message: "No Profile catalog loaded"})
		r.Finish()
		return r
	}
	r.Profiles = len(w.Catalog.Profiles)
	r.NodeSchemas = len(w.Catalog.Schemas)
	r.RecordSchemas = len(w.Catalog.Records)
	for k := range r.Stages {
		r.Stages[k] = "PASS"
	}
	cache := map[string]*linkedDocument{}
	visiting := map[string]bool{}
	add := func(doc *Document, code, phase, message string, span Range, path string) {
		r.Diagnostics = append(r.Diagnostics, Diagnostic{Code: code, Phase: phase, Severity: "error", Message: message, File: doc.File, Range: span, Path: path})
	}
	var load func(string, *Document, Range) *linkedDocument
	load = func(file string, origin *Document, where Range) *linkedDocument {
		p, e := ResolvePath(w.Root, file)
		if e != nil {
			code, phase := "PATH_ESCAPE", "link"
			if !strings.Contains(e.Error(), code) {
				code, phase = "IO", "io"
			}
			add(origin, code, phase, e.Error(), where, "")
			return nil
		}
		if visiting[p] {
			add(origin, "IMPORT_CYCLE", "link", "Circular module import: "+p, where, "")
			return nil
		}
		if d, ok := cache[p]; ok {
			return d
		}
		if len(visiting) >= MaxDepth {
			add(origin, "LIMIT", "link", "Module nesting limit exceeded", where, "")
			return nil
		}
		b, ok := w.Documents[p]
		if !ok {
			b, e = ReadSource(p)
			if e != nil {
				code, phase := "IO", "io"
				if strings.HasPrefix(e.Error(), "LIMIT:") {
					code, phase = "LIMIT", "syntax"
				}
				add(origin, code, phase, e.Error(), where, "")
				cache[p] = nil
				return nil
			}
		}
		d := Parse(p, b, ParseOptions{Recover: true})
		r.Documents++
		r.Diagnostics = append(r.Diagnostics, d.Diagnostics...)
		info := &linkedDocument{doc: d, ids: map[string]*Node{}, exports: map[string]bool{}, imports: map[string]*linkedDocument{}}
		cache[p] = info
		if HasErrors(d.Diagnostics) {
			for _, phase := range []string{"shape", "link", "selectedStaticRules"} {
				if r.Stages[phase] == "PASS" {
					r.Stages[phase] = "NOT_EVALUATED"
				}
			}
			return info
		}
		r.Diagnostics = append(r.Diagnostics, w.Catalog.Check(d)...)
		for _, n := range Walk(d.Nodes) {
			if n.ID != "" {
				if _, exists := info.ids[n.ID]; !exists {
					info.ids[n.ID] = n
				}
			}
		}
		var modules []*Node
		for _, n := range d.Nodes {
			if n.Kind == "module" {
				modules = append(modules, n)
			}
		}
		if len(modules) > 1 {
			add(d, "DUPLICATE_MODULE", "link", "A document may declare only one module", modules[1].Range, "")
		}
		for _, n := range modules {
			if v := n.Get("exports"); v != nil {
				for _, x := range v.Items {
					id := Text(x)
					if info.ids[id] == nil {
						add(d, "UNKNOWN_EXPORT", "link", "Export does not identify a declaration: "+id, x.Range, "module.exports")
					}
					info.exports[id] = true
				}
			}
		}
		visiting[p] = true
		for _, n := range d.Nodes {
			if n.Kind != "use" {
				continue
			}
			v := n.Get("source")
			if v == nil {
				continue
			}
			if v.Tag != "ref" || v.URI == nil {
				add(d, "IMPORT_REQUIRES_URI", "link", "Imports require a local URI reference", v.Range, "use.source")
				continue
			}
			u, err := url.Parse(*v.URI)
			if err != nil || u.Scheme != "" || u.Host != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(*v.URI, "#") {
				add(d, "IMPORT_NONLOCAL_NOT_SUPPORTED", "link", "Only explicit local module URIs are supported", v.Range, "use.source")
				continue
			}
			if u.Path == "" {
				add(d, "IMPORT_REQUIRES_URI", "link", "Import URI must name a file", v.Range, "use.source")
				continue
			}
			target := filepath.FromSlash(u.Path)
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(p), target)
			}
			info.imports[n.ID] = load(target, d, v.Range)
		}
		delete(visiting, p)
		return info
	}
	for _, f := range files {
		origin := NewDocument(f, nil)
		load(f, origin, Range{})
	}
	paths := make([]string, 0, len(cache))
	for p := range cache {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		info := cache[p]
		if info == nil || HasErrors(info.doc.Diagnostics) {
			continue
		}
		d := info.doc
		for _, n := range Walk(d.Nodes) {
			schema := w.Catalog.Schemas[n.Kind]
			if schema == nil {
				continue
			}
			for _, property := range w.Catalog.Properties(n) {
				field := schema.Fields[property.Name]
				if field == nil {
					continue
				}
				for _, typed := range w.Catalog.References(property.Value, field) {
					ref := typed.Value
					r.References++
					path := n.Kind
					if n.ID != "" {
						path += "#" + n.ID
					}
					path += "." + property.Name
					if ref.URI != nil {
						if n.Kind != "use" {
							r.ExternalResources = append(r.ExternalResources, ExternalResource{d.File, *ref.URI, "NOT_RESOLVED_BY_CHECKER"})
						}
						continue
					}
					dest := info
					if ref.Module != "" {
						var ok bool
						dest, ok = info.imports[ref.Module]
						if !ok {
							add(d, "UNKNOWN_IMPORT", "link", "Unknown import alias: "+ref.Module, ref.Range, path)
							continue
						}
						if dest == nil || HasErrors(dest.doc.Diagnostics) {
							continue
						}
						if !dest.exports[ref.ID] {
							add(d, "NOT_EXPORTED", "link", "Identity is not exported: "+ref.ID, ref.Range, path)
							continue
						}
					}
					target := dest.ids[ref.ID]
					if target == nil {
						add(d, "UNRESOLVED_REF", "link", "Unknown identity: "+ref.ID, ref.Range, path)
						continue
					}
					if len(ref.Ports) > 0 {
						s := w.Catalog.Schemas[target.Kind]
						if s == nil {
							continue
						}
						_, ok := s.Ports[ref.Ports[0]]
						if len(ref.Ports) != 1 || !ok {
							add(d, "UNKNOWN_PORT", "link", "Unknown port on "+target.Kind, ref.Range, path)
						}
					} else if len(typed.Targets) > 0 {
						valid := false
						for _, k := range typed.Targets {
							if k == target.Kind {
								valid = true
								break
							}
						}
						if !valid {
							add(d, "REF_TARGET_KIND", "link", "Reference target has incompatible kind: "+target.Kind, ref.Range, path)
						}
					}
				}
			}
		}
	}
	if len(files) == 0 {
		r.Diagnostics = append(r.Diagnostics, Diagnostic{Code: "INPUT_REQUIRED", Phase: "tool", Severity: "error", Message: "At least one explicit file is required"})
	}
	for _, info := range cache {
		if info != nil {
			info.doc.AttachDecodedRanges(r.Diagnostics)
		}
	}
	r.Finish()
	return r
}

func References(v *Value) []*Value {
	var out []*Value
	var visit func(*Value)
	visit = func(x *Value) {
		if x == nil {
			return
		}
		if x.Tag == "ref" {
			out = append(out, x)
		}
		for _, a := range x.Items {
			visit(a)
		}
		for _, p := range x.Fields {
			visit(p.Value)
		}
		for _, a := range x.Args {
			visit(a)
		}
		visit(x.Left)
		visit(x.Right)
		visit(x.Object)
		visit(x.Index)
		visit(x.Tree)
	}
	visit(v)
	return out
}
