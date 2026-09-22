package uicl

import (
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Payhon/uicl/profiles"
)

// FieldSpec comes directly from a Profile property declaration.
type FieldSpec struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Expression  bool     `json:"expression"`
	Default     *Value   `json:"default,omitempty"`
	Enum        []*Value `json:"enum,omitempty"`
	Targets     []string `json:"targets,omitempty"`
	Definition  *Node    `json:"definition"`
}

type Schema struct {
	Name       string                `json:"name"`
	Profile    string                `json:"profile"`
	Primary    string                `json:"primary,omitempty"`
	Identity   string                `json:"identity,omitempty"`
	Children   []string              `json:"children"`
	Fields     map[string]*FieldSpec `json:"fields"`
	Ports      map[string]string     `json:"ports"`
	Definition *Node                 `json:"definition"`
	File       string                `json:"file"`
}

type Profile struct {
	ID         string   `json:"id"`
	Version    string   `json:"version"`
	Core       string   `json:"core"`
	Path       string   `json:"path"`
	Digest     string   `json:"sha256"`
	Requires   []string `json:"requires"`
	Definition *Node    `json:"definition"`
	Source     string   `json:"source"`
}

type Catalog struct {
	Schemas  map[string]*Schema  `json:"schemas"`
	Records  map[string]*Schema  `json:"records"`
	Profiles map[string]*Profile `json:"profiles"`
}

// LoadCatalog loads embedded standards unless profileDir explicitly replaces
// them. Extensions are opt-in, local and subject to the same root boundary.
// Invalid source/schema definitions are returned as diagnostics, I/O as errors.
func LoadCatalog(root, profileDir string, extra []string) (*Catalog, []Diagnostic, error) {
	c := &Catalog{Schemas: map[string]*Schema{}, Records: map[string]*Schema{}, Profiles: map[string]*Profile{}}
	var docs []*Document
	var ds []Diagnostic
	load := func(file string, source []byte) {
		d := ParseCore(file, source, ParseOptions{Recover: true})
		docs = append(docs, d)
		ds = append(ds, d.Diagnostics...)
	}
	var files []string
	if profileDir == "" {
		names, err := fs.Glob(profiles.Files, "*.uicl")
		if err != nil {
			return nil, nil, err
		}
		for _, name := range names {
			source, err := profiles.Files.ReadFile(name)
			if err != nil {
				return nil, nil, err
			}
			load("embedded:profiles/"+name, source)
		}
	} else {
		dir, err := ResolvePath(root, profileDir)
		if err != nil {
			return nil, nil, err
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".uicl") {
				files = append(files, filepath.Join(dir, entry.Name()))
			}
		}
		if len(files) == 0 {
			return nil, nil, fmt.Errorf("no .uicl Profile sources in %s", dir)
		}
	}
	files = append(files, extra...)
	for _, name := range files {
		file, err := ResolvePath(root, name)
		if err != nil {
			return nil, nil, err
		}
		info, err := os.Stat(file)
		if err != nil {
			return nil, nil, err
		}
		if !info.Mode().IsRegular() || info.Size() > MaxBytes {
			return nil, nil, fmt.Errorf("Profile source is not a regular file within %d bytes: %s", MaxBytes, file)
		}
		source, err := ReadSource(file)
		if err != nil {
			return nil, nil, err
		}
		load(file, source)
	}
	add := func(file, code, message string, r Range) {
		ds = append(ds, Diagnostic{Code: code, Phase: "profile", Severity: "error", Message: message, File: file, Range: r})
	}
	for _, d := range docs {
		for _, n := range d.Nodes {
			if n.Kind != "profile" {
				add(d.File, "PROFILE_ROOT", "Profile sources may only contain profile declarations", n.KindRange)
				continue
			}
			id := n.Text("id")
			if id == "" {
				add(d.File, "PROFILE_ID", "Profile id must be nonempty text", n.Range)
				continue
			}
			if old := c.Profiles[id]; old != nil {
				add(d.File, "PROFILE_CONFLICT", "duplicate Profile "+id+" (already defined in "+old.Path+")", n.Range)
				continue
			}
			p := &Profile{ID: id, Version: n.Text("version"), Core: n.Text("core"), Path: d.File, Digest: Digest(d.Source), Requires: Strings(n.Get("requires")), Definition: n, Source: string(d.Source)}
			c.Profiles[id] = p
			if p.Core != CoreVersion {
				add(d.File, "PROFILE_CORE", "Profile requires unsupported Core "+p.Core, n.Range)
			}
			if p.Version == "" {
				add(d.File, "PROFILE_VERSION", "Profile version must be nonempty text", n.Range)
			}
			for _, def := range n.Children {
				if def.Kind != "schema" && def.Kind != "record_schema" {
					continue
				}
				name := declarationName(def, "kind")
				dest := c.Schemas
				if def.Kind == "record_schema" {
					name, dest = declarationName(def, "name"), c.Records
				}
				if name == "" {
					add(d.File, "SCHEMA_NAME", "schema name must be nonempty text", def.Range)
					continue
				}
				if old := dest[name]; old != nil {
					add(d.File, "SCHEMA_CONFLICT", "duplicate schema "+name+" (already defined in "+old.File+")", def.Range)
					continue
				}
				s := &Schema{Name: name, Profile: id, Primary: def.Text("primary"), Identity: def.Text("identity"), Children: Strings(def.Get("children")), Fields: map[string]*FieldSpec{}, Ports: map[string]string{}, Definition: def, File: d.File}
				dest[name] = s
				for _, child := range def.Children {
					key := declarationName(child, "name")
					switch child.Kind {
					case "property":
						if key == "" {
							add(d.File, "SCHEMA_PROPERTY_NAME", "property name must be nonempty text", child.Range)
							continue
						}
						if s.Fields[key] != nil {
							add(d.File, "DUPLICATE_SCHEMA_PROPERTY", "duplicate property "+name+"."+key, child.Range)
							continue
						}
						f := &FieldSpec{Name: key, Type: child.Text("type"), Description: child.Text("description"), Required: boolValue(child.Get("required")), Expression: boolValue(child.Get("expression")), Default: child.Get("default"), Targets: Strings(child.Get("targets")), Definition: child}
						if enum := child.Get("enum"); enum != nil {
							f.Enum = enum.Items
						}
						s.Fields[key] = f
					case "port":
						if key == "" {
							add(d.File, "SCHEMA_PORT_NAME", "port name must be nonempty text", child.Range)
							continue
						}
						if _, exists := s.Ports[key]; exists {
							add(d.File, "DUPLICATE_SCHEMA_PORT", "duplicate port "+name+"."+key, child.Range)
							continue
						}
						s.Ports[key] = child.Text("type")
						if s.Ports[key] == "" {
							add(d.File, "SCHEMA_PORT_TYPE", "port business type must be nonempty", child.Range)
						}
					}
				}
			}
		}
	}
	// All declarations are collected before resolving types or dependencies.
	for _, group := range []map[string]*Schema{c.Schemas, c.Records} {
		for _, s := range group {
			for _, f := range s.Fields {
				if !c.validType(f.Type, 0) {
					add(s.File, "UNKNOWN_SHAPE_TYPE", "invalid or unknown shape type "+f.Type, f.Definition.Range)
					continue
				}
				if f.Default != nil && !c.TypeOK(f.Default, f.Type) {
					add(s.File, "SCHEMA_DEFAULT_TYPE", "default does not match "+f.Type, f.Default.Range)
				}
				for _, v := range f.Enum {
					if !c.TypeOK(v, f.Type) {
						add(s.File, "SCHEMA_ENUM_TYPE", "enum member does not match "+f.Type, v.Range)
					}
				}
				if f.Default != nil && f.Definition.Get("enum") != nil && !enumContains(f.Enum, f.Default) {
					add(s.File, "SCHEMA_DEFAULT_ENUM", "default is outside enum", f.Default.Range)
				}
				for _, target := range f.Targets {
					if c.Schemas[target] == nil {
						add(s.File, "SCHEMA_TARGET", "unknown target node kind "+target, f.Definition.Range)
					}
				}
			}
			if s.Primary != "" {
				f := s.Fields[s.Primary]
				if f == nil || !c.TypeOK(&Value{Tag: "literal", Type: "text", Text: ""}, f.Type) {
					add(s.File, "SCHEMA_PRIMARY", "primary must name a field accepting text: "+s.Primary, s.Definition.Range)
				}
			}
			for _, pattern := range s.Children {
				found := false
				for name := range c.Schemas {
					if schemaMatch(pattern, name) {
						found = true
						break
					}
				}
				if !found {
					add(s.File, "SCHEMA_CHILD", "child pattern matches no registered node: "+pattern, s.Definition.Range)
				}
			}
		}
	}
	state := map[string]int{}
	var visit func(string)
	visit = func(id string) {
		p := c.Profiles[id]
		if state[id] == 2 {
			return
		}
		if state[id] == 1 {
			add(p.Path, "PROFILE_CYCLE", "Profile dependency cycle at "+id, p.Definition.Range)
			return
		}
		state[id] = 1
		seen := map[string]bool{}
		for _, dep := range p.Requires {
			i := strings.LastIndex(dep, "@")
			if i <= 0 || i == len(dep)-1 {
				add(p.Path, "PROFILE_DEPENDENCY", "dependency must name an exact version: "+dep, p.Definition.Range)
				continue
			}
			name, version := dep[:i], dep[i+1:]
			if seen[name] {
				add(p.Path, "PROFILE_DEPENDENCY", "duplicate dependency "+name, p.Definition.Range)
				continue
			}
			seen[name] = true
			other := c.Profiles[name]
			if other == nil || other.Version != version {
				add(p.Path, "PROFILE_DEPENDENCY", "missing dependency or version mismatch: "+dep, p.Definition.Range)
				continue
			}
			visit(name)
		}
		state[id] = 2
	}
	for _, id := range sortedKeys(c.Profiles) {
		visit(id)
	}
	// Profile declarations are checked against the canonical Core metamodel,
	// independently of replaceable domain definitions. A replacement cannot
	// disable its own validation by weakening schema/property declarations.
	meta, err := profileMetamodel()
	if err != nil {
		return nil, nil, err
	}
	for _, d := range docs {
		ds = append(ds, meta.Check(d)...)
	}
	SortDiagnostics(ds)
	return c, ds, nil
}

func profileMetamodel() (*Catalog, error) {
	source, err := profiles.Files.ReadFile("00-meta.uicl")
	if err != nil {
		return nil, err
	}
	d := ParseCore("embedded:profiles/00-meta.uicl", source, ParseOptions{})
	if HasErrors(d.Diagnostics) {
		return nil, fmt.Errorf("invalid embedded Profile metamodel: %v", d.Diagnostics)
	}
	c := &Catalog{Schemas: map[string]*Schema{}, Records: map[string]*Schema{}, Profiles: map[string]*Profile{}}
	for _, p := range d.Nodes {
		for _, n := range p.Children {
			if n.Kind != "schema" {
				continue
			}
			s := &Schema{Name: declarationName(n, "kind"), Primary: n.Text("primary"), Identity: n.Text("identity"), Children: Strings(n.Get("children")), Fields: map[string]*FieldSpec{}, Ports: map[string]string{}}
			for _, field := range n.Children {
				if field.Kind != "property" {
					continue
				}
				f := &FieldSpec{Name: declarationName(field, "name"), Type: field.Text("type"), Required: boolValue(field.Get("required")), Expression: boolValue(field.Get("expression")), Default: field.Get("default"), Definition: field}
				if enum := field.Get("enum"); enum != nil {
					f.Enum = enum.Items
				}
				s.Fields[f.Name] = f
			}
			c.Schemas[s.Name] = s
		}
	}
	return c, nil
}

func sortedKeys[V any](m map[string]V) []string {
	a := make([]string, 0, len(m))
	for k := range m {
		a = append(a, k)
	}
	sort.Strings(a)
	return a
}
func declarationName(n *Node, key string) string {
	if n.Label != nil {
		return Text(n.Label)
	}
	return n.Text(key)
}
func boolValue(v *Value) bool { return v != nil && v.Tag == "literal" && v.Type == "bool" && v.Bool }
func schemaMatch(pattern, name string) bool {
	ok, err := path.Match(pattern, name)
	return err == nil && ok
}

// Properties provides primary-slot normalization without changing the AST.
func (c *Catalog) Properties(n *Node) []*Property {
	props := append([]*Property(nil), n.Props...)
	s := c.Schemas[n.Kind]
	if n.Label != nil && s != nil && s.Primary != "" && n.Property(s.Primary) == nil {
		props = append(props, &Property{Name: s.Primary, Value: n.Label, Range: n.Label.Range, NameRange: n.Label.Range})
	}
	return props
}

// TypedReference carries the constraints of the actual nested Profile field.
type TypedReference struct {
	Value   *Value
	Targets []string
}

// References follows named record/list types so nested reference fields retain
// their own target constraints. Untyped values still yield all literal refs.
func (c *Catalog) References(v *Value, field *FieldSpec) []TypedReference {
	var out []TypedReference
	var visit func(*Value, string, []string, int)
	visit = func(value *Value, typ string, targets []string, depth int) {
		if value == nil {
			return
		}
		fallback := func() {
			for _, ref := range References(value) {
				out = append(out, TypedReference{ref, targets})
			}
		}
		if depth > MaxDepth || value.Tag == "expression" {
			fallback()
			return
		}
		branches, ok := typeBranches(typ)
		if !ok {
			fallback()
			return
		}
		if len(branches) > 1 {
			for _, branch := range branches {
				if c.TypeOK(value, branch) {
					visit(value, branch, targets, depth+1)
					return
				}
			}
			fallback()
			return
		}
		if strings.HasPrefix(typ, "list<") && strings.HasSuffix(typ, ">") && value.Tag == "list" {
			for _, item := range value.Items {
				visit(item, typ[5:len(typ)-1], targets, depth+1)
			}
			return
		}
		if strings.HasPrefix(typ, "record<") && strings.HasSuffix(typ, ">") && value.Tag == "record" {
			if schema := c.Records[typ[7:len(typ)-1]]; schema != nil {
				for _, prop := range value.Fields {
					childType, childTargets := "value", targets
					if child := schema.Fields[prop.Name]; child != nil {
						childType = child.Type
						if len(child.Targets) > 0 {
							childTargets = child.Targets
						}
					}
					visit(prop.Value, childType, childTargets, depth+1)
				}
				return
			}
		}
		fallback()
	}
	if field == nil {
		visit(v, "value", nil, 0)
	} else {
		visit(v, field.Type, field.Targets, 0)
	}
	return out
}

func typeBranches(t string) ([]string, bool) {
	depth, start := 0, 0
	var out []string
	for i, ch := range t {
		switch ch {
		case '<':
			depth++
		case '>':
			depth--
			if depth < 0 {
				return nil, false
			}
		case '|':
			if depth == 0 {
				if i == start {
					return nil, false
				}
				out = append(out, t[start:i])
				start = i + 1
			}
		}
	}
	if depth != 0 || start == len(t) {
		return nil, false
	}
	return append(out, t[start:]), true
}
func (c *Catalog) validType(t string, depth int) bool {
	if depth > MaxDepth {
		return false
	}
	branches, ok := typeBranches(t)
	if !ok {
		return false
	}
	if len(branches) > 1 {
		for _, b := range branches {
			if !c.validType(b, depth+1) {
				return false
			}
		}
		return true
	}
	switch t {
	case "value", "ref", "text", "number", "int", "decimal", "bool", "null", "record":
		return true
	}
	if strings.HasPrefix(t, "list<") && strings.HasSuffix(t, ">") {
		return c.validType(t[5:len(t)-1], depth+1)
	}
	if strings.HasPrefix(t, "record<") && strings.HasSuffix(t, ">") {
		return c.Records[t[7:len(t)-1]] != nil
	}
	return false
}

func (c *Catalog) TypeOK(v *Value, t string) bool { return c.typeOK(v, t, 0) }
func (c *Catalog) typeOK(v *Value, t string, depth int) bool {
	if v == nil || depth > MaxDepth {
		return false
	}
	branches, ok := typeBranches(t)
	if !ok {
		return false
	}
	if len(branches) > 1 {
		for _, b := range branches {
			if c.typeOK(v, b, depth+1) {
				return true
			}
		}
		return false
	}
	switch t {
	case "value":
		return true
	case "ref":
		return v.Tag == "ref"
	case "text":
		return v.Tag == "raw" || v.Tag == "literal" && v.Type == "text"
	case "number":
		return v.Tag == "literal" && (v.Type == "int" || v.Type == "decimal")
	case "int", "decimal", "bool", "null":
		return v.Tag == "literal" && v.Type == t
	case "record":
		return v.Tag == "record"
	}
	if strings.HasPrefix(t, "list<") && strings.HasSuffix(t, ">") {
		if v.Tag != "list" {
			return false
		}
		for _, item := range v.Items {
			if !c.typeOK(item, t[5:len(t)-1], depth+1) {
				return false
			}
		}
		return true
	}
	if strings.HasPrefix(t, "record<") && strings.HasSuffix(t, ">") {
		s := c.Records[t[7:len(t)-1]]
		if v.Tag != "record" || s == nil {
			return false
		}
		return len(c.checkFields("", v.Fields, s.Fields, "", v.Range, depth+1)) == 0
	}
	return false
}

func enumContains(values []*Value, v *Value) bool {
	for _, candidate := range values {
		if equalProfileValue(candidate, v) {
			return true
		}
	}
	return false
}

// Compare values independently of source ranges, raw spelling and record order.
func equalProfileValue(a, b *Value) bool {
	if a == nil || b == nil {
		return a == b
	}
	textual := func(v *Value) bool { return v.Tag == "raw" || v.Tag == "literal" && v.Type == "text" }
	if textual(a) && textual(b) {
		return a.Text == b.Text
	}
	if a.Tag != b.Tag {
		return false
	}
	switch a.Tag {
	case "literal":
		number := func(v *Value) bool { return v.Type == "int" || v.Type == "decimal" }
		if number(a) && number(b) {
			x, xok := new(big.Rat).SetString(a.Text)
			y, yok := new(big.Rat).SetString(b.Text)
			return xok && yok && x.Cmp(y) == 0
		}
		if a.Type != b.Type {
			return false
		}
		if a.Type == "bool" {
			return a.Bool == b.Bool
		}
		if a.Type == "null" {
			return true
		}
		return a.Text == b.Text
	case "list":
		if len(a.Items) != len(b.Items) {
			return false
		}
		for i := range a.Items {
			if !equalProfileValue(a.Items[i], b.Items[i]) {
				return false
			}
		}
		return true
	case "record":
		if len(a.Fields) != len(b.Fields) {
			return false
		}
		for _, p := range a.Fields {
			if !equalProfileValue(p.Value, Field(b, p.Name)) {
				return false
			}
		}
		return true
	case "ref":
		if a.URI == nil != (b.URI == nil) {
			return false
		}
		if a.URI != nil && *a.URI != *b.URI {
			return false
		}
		if a.Module != b.Module || a.ID != b.ID || len(a.Ports) != len(b.Ports) {
			return false
		}
		for i := range a.Ports {
			if a.Ports[i] != b.Ports[i] {
				return false
			}
		}
		return true
	}
	// Enum declarations accept values, not executable expression comparisons.
	return false
}

func (c *Catalog) checkFields(file string, actual []*Property, expected map[string]*FieldSpec, semanticPath string, r Range, depth int) []Diagnostic {
	var ds []Diagnostic
	add := func(code, message, field string, span Range) {
		ds = append(ds, Diagnostic{Code: code, Phase: "shape", Severity: "error", Message: message, File: file, Range: span, Path: semanticPath + "." + field})
	}
	seen := map[string]bool{}
	for _, p := range actual {
		if seen[p.Name] {
			add("DUPLICATE_PROPERTY", "duplicate property "+p.Name, p.Name, p.NameRange)
			continue
		}
		seen[p.Name] = true
		f := expected[p.Name]
		if f == nil {
			add("UNKNOWN_PROPERTY", "unknown property "+p.Name, p.Name, p.NameRange)
			continue
		}
		v := p.Value
		if v == nil {
			add("TYPE_SHAPE", "missing property value", p.Name, p.Range)
			continue
		}
		if v.Tag == "expression" {
			if !f.Expression {
				add("EXPRESSION_FORBIDDEN", "expressions are not allowed for "+p.Name, p.Name, v.Range)
			}
			continue
		}
		if !c.typeOK(v, f.Type, depth+1) {
			add("TYPE_SHAPE", p.Name+" requires "+f.Type, p.Name, v.Range)
			continue
		}
		if f.Definition != nil && f.Definition.Get("enum") != nil && !enumContains(f.Enum, v) {
			add("ENUM", "value is outside the declared enum", p.Name, v.Range)
		}
	}
	for _, name := range sortedKeys(expected) {
		f := expected[name]
		if !seen[name] && f.Required && f.Default == nil {
			add("MISSING_PROPERTY", "missing required property "+name, name, r)
		}
	}
	return ds
}

// Check evaluates registered shapes and selected domain rules only.
func (c *Catalog) Check(d *Document) []Diagnostic {
	var ds []Diagnostic
	ids := map[string]*Node{}
	add := func(code, message string, n *Node, r Range) {
		p := n.Kind
		if n.ID != "" {
			p += "#" + n.ID
		}
		ds = append(ds, Diagnostic{Code: code, Phase: "shape", Severity: "error", Message: message, File: d.File, Range: r, Path: p})
	}
	for _, n := range Walk(d.Nodes) {
		s := c.Schemas[n.Kind]
		if s == nil {
			add("UNKNOWN_NODE", "unknown node kind "+n.Kind, n, n.KindRange)
			continue
		}
		if n.ID != "" {
			if old := ids[n.ID]; old != nil {
				add("DUPLICATE_ID", "duplicate identity "+n.ID, n, n.IDRange)
				ds[len(ds)-1].Related = []Location{{File: d.File, Range: old.IDRange, Message: "first declaration"}}
			} else {
				ids[n.ID] = n
			}
		}
		if s.Identity == "required" && n.ID == "" {
			add("MISSING_ID", n.Kind+" requires an identity", n, n.KindRange)
		}
		if s.Identity == "forbidden" && n.ID != "" {
			add("FORBIDDEN_ID", n.Kind+" does not accept an identity", n, n.IDRange)
		}
		if n.Label != nil {
			if s.Primary == "" {
				add("PRIMARY_FORBIDDEN", n.Kind+" has no primary text slot", n, n.Label.Range)
			} else if n.Property(s.Primary) != nil {
				add("PRIMARY_CONFLICT", "primary text conflicts with "+s.Primary, n, n.Label.Range)
			}
		}
		p := n.Kind
		if n.ID != "" {
			p += "#" + n.ID
		}
		ds = append(ds, c.checkFields(d.File, c.Properties(n), s.Fields, p, n.KindRange, 0)...)
		for _, child := range n.Children {
			ok := false
			for _, pattern := range s.Children {
				if schemaMatch(pattern, child.Kind) {
					ok = true
					break
				}
			}
			if !ok {
				add("CHILD_FORBIDDEN", n.Kind+" does not accept child "+child.Kind, child, child.KindRange)
			}
		}
		ds = append(ds, c.selected(d.File, n)...)
	}
	SortDiagnostics(ds)
	return ds
}

func (c *Catalog) selected(file string, n *Node) []Diagnostic {
	var ds []Diagnostic
	p := map[string]*Value{}
	for _, f := range c.Properties(n) {
		p[f.Name] = f.Value
	}
	add := func(code, message string, value *Value) {
		r := n.KindRange
		if value != nil {
			r = value.Range
		}
		ds = append(ds, Diagnostic{Code: code, Phase: "static", Severity: "error", Message: message, File: file, Range: r, Path: n.Kind + "#" + n.ID})
	}
	xor := func(a, b string, required bool) {
		count := 0
		if p[a] != nil {
			count++
		}
		if p[b] != nil {
			count++
		}
		if count > 1 || required && count != 1 {
			add("MUTUALLY_EXCLUSIVE", "expected "+a+" or "+b+" exclusively", nil)
		}
	}
	switch n.Kind {
	case "file":
		xor("source", "body", true)
	case "app":
		xor("target", "targets", false)
	case "endpoint":
		xor("action", "query", true)
	case "type":
		enum := p["enum"]
		hasEnum := enum != nil && len(enum.Items) > 0
		if hasEnum == (len(n.Children) > 0) {
			add("TYPE_RECORD_OR_ENUM", "type must define either nonempty enum or record fields", nil)
		}
	case "collection":
		if Text(p["storage"]) == "server" && p["access"] == nil {
			policy := false
			for _, child := range n.Children {
				if child.Kind == "policy" {
					policy = true
				}
			}
			if !policy {
				add("SERVER_ACCESS_REQUIRED", "server collection requires access or a policy", nil)
			}
		}
	case "g.repeat":
		lo, lok := Int(p["min"])
		hi, hok := Int(p["max"])
		if len(n.Children) != 1 || lok && lo < 0 || lok && hok && hi < lo {
			add("GRAMMAR_REPEAT_BOUNDS", "repeat requires one child and ordered nonnegative bounds", nil)
		}
	case "secret":
		if scope := p["scope"]; scope != nil && scope.Tag == "list" && len(scope.Items) == 0 {
			add("SECRET_EMPTY_SCOPE", "secret scope cannot be empty", scope)
		}
	case "pg.migration":
		if Text(p["atomic"]) == "required" && p["steps"] != nil {
			for _, step := range p["steps"].Items {
				if Text(Field(step, "build")) == "concurrently" {
					add("PG_TRANSACTION_CONFLICT", "concurrent index building cannot run in a required atomic migration", step)
				}
			}
		}
	case "pg.table":
		names := map[string]bool{}
		if cols := p["columns"]; cols != nil {
			for _, col := range cols.Items {
				name := Text(Field(col, "name"))
				if name == "" {
					continue
				}
				if names[name] {
					add("DUPLICATE_COLUMN", "duplicate column "+name, col)
				}
				names[name] = true
				if Field(col, "default") != nil && Field(col, "default_sql") != nil {
					add("PG_DEFAULT_CONFLICT", "column default and default_sql are mutually exclusive", col)
				}
			}
		}
		if keys := p["primary_key"]; keys != nil {
			for _, key := range keys.Items {
				if !names[Text(key)] {
					add("PG_UNKNOWN_KEY_COLUMN", "unknown primary key column "+Text(key), key)
				}
			}
		}
		for _, group := range []string{"indexes", "unique", "foreign_keys"} {
			if values := p[group]; values != nil {
				for _, entry := range values.Items {
					if cols := Field(entry, "columns"); cols != nil {
						for _, col := range cols.Items {
							if !names[Text(col)] {
								add("PG_UNKNOWN_COLUMN", "unknown column "+Text(col), col)
							}
						}
					}
				}
			}
		}
	}
	if n.Kind == "image" || n.Kind == "video" || n.Kind == "design.canvas" {
		if size := p["size"]; size != nil && size.Tag == "list" {
			ok := len(size.Items) == 2
			for _, v := range size.Items {
				number, valid := Int(v)
				if !valid || number <= 0 {
					ok = false
				}
			}
			if !ok {
				add("DIMENSION_PAIR", "size must contain two positive integers", size)
			}
		}
	}
	if n.Kind == "video" {
		frames, validFrames := Int(p["frames"])
		num, validNum := Int(Field(p["fps"], "num"))
		den, validDen := Int(Field(p["fps"], "den"))
		if validFrames && frames <= 0 || validNum && num <= 0 || validDen && den <= 0 {
			add("VIDEO_TIMEBASE", "frames and both fps components must be positive", nil)
		}
		for _, track := range n.Children {
			if track.Kind != "track" {
				continue
			}
			type segment struct {
				start, end int64
				node       *Node
			}
			var segments []segment
			for _, shot := range track.Children {
				if shot.Kind != "shot" && shot.Kind != "cue" {
					continue
				}
				start, okStart := Int(shot.Get("start_frame"))
				length, okLength := Int(shot.Get("frames"))
				if !okStart || !okLength || !validFrames {
					continue
				}
				// Subtraction keeps the bounds check safe at signed integer limits.
				if start < 0 || length <= 0 || start > frames || length > frames-start {
					add("TIMELINE_BOUNDS", "segment is outside the video timeline", shot.Get("start_frame"))
					continue
				}
				segments = append(segments, segment{start, start + length, shot})
			}
			sort.SliceStable(segments, func(i, j int) bool { return segments[i].start < segments[j].start })
			if track.Text("overlap") != "allow" {
				for i := 1; i < len(segments); i++ {
					if segments[i-1].end > segments[i].start {
						add("TIMELINE_OVERLAP", "segments overlap in a track that forbids overlap", segments[i].node.Get("start_frame"))
					}
				}
			}
		}
	}
	return ds
}
