// Package uicl parses, checks and edits UICL contracts without executing them.
package uicl

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"unicode/utf8"
)

const Version = "0.1.0"
const CoreVersion = "1.0"
const MaxBytes = 16 * 1024 * 1024
const MaxDepth = 96

// Position is zero-based; Column counts Unicode scalars, Offset counts bytes.
type Position struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}
type Location struct {
	File    string `json:"file"`
	Range   Range  `json:"range"`
	Message string `json:"message,omitempty"`
}
type Diagnostic struct {
	Code         string     `json:"code"`
	Phase        string     `json:"phase"`
	Severity     string     `json:"severity"`
	Message      string     `json:"message"`
	File         string     `json:"file"`
	Range        Range      `json:"range"`
	Path         string     `json:"path,omitempty"`
	Related      []Location `json:"relatedInformation,omitempty"`
	DecodedRange *Range     `json:"decodedRange,omitempty"`
}
type Token struct {
	Kind  string `json:"kind"`
	Text  string `json:"text"`
	Range Range  `json:"range"`
}
type Property struct {
	Name      string `json:"name"`
	Value     *Value `json:"value"`
	Range     Range  `json:"range"`
	NameRange Range  `json:"nameRange"`
}

// Numeric Text holds the exact decimal lexeme, never a binary floating value.
type Value struct {
	Tag      string      `json:"tag"`
	Type     string      `json:"type,omitempty"`
	Text     string      `json:"text,omitempty"`
	Bool     bool        `json:"bool,omitempty"`
	Range    Range       `json:"range"`
	Items    []*Value    `json:"items,omitempty"`
	Fields   []*Property `json:"fields,omitempty"`
	URI      *string     `json:"uri,omitempty"`
	Module   string      `json:"module,omitempty"`
	ID       string      `json:"id,omitempty"`
	Ports    []string    `json:"ports,omitempty"`
	Name     string      `json:"name,omitempty"`
	Op       string      `json:"op,omitempty"`
	Left     *Value      `json:"left,omitempty"`
	Right    *Value      `json:"right,omitempty"`
	Object   *Value      `json:"object,omitempty"`
	Index    *Value      `json:"index,omitempty"`
	Tree     *Value      `json:"tree,omitempty"`
	Args     []*Value    `json:"args,omitempty"`
	Optional bool        `json:"optional,omitempty"`
	Form     string      `json:"form,omitempty"`
	Language string      `json:"language,omitempty"`
}
type Node struct {
	Kind      string      `json:"kind"`
	ID        string      `json:"id,omitempty"`
	Label     *Value      `json:"label,omitempty"`
	Props     []*Property `json:"properties"`
	Children  []*Node     `json:"children"`
	Range     Range       `json:"range"`
	KindRange Range       `json:"kindRange"`
	IDRange   Range       `json:"idRange"`
}
type ParseOptions struct {
	Recover  bool
	Fragment bool
}
type Document struct {
	File        string       `json:"file"`
	Mode        string       `json:"mode"`
	Source      []byte       `json:"-"`
	Nodes       []*Node      `json:"nodes"`
	Tokens      []Token      `json:"tokens,omitempty"`
	Islands     []*Island    `json:"islands,omitempty"`
	HostAST     any          `json:"-"`
	Diagnostics []Diagnostic `json:"diagnostics"`
	lineStarts  []int
}

func NewDocument(file string, source []byte) *Document {
	d := &Document{File: file, Source: append([]byte(nil), source...), Mode: "structured", Nodes: []*Node{}, Diagnostics: []Diagnostic{}, lineStarts: []int{0}}
	for i, b := range source {
		if b == '\n' {
			d.lineStarts = append(d.lineStarts, i+1)
		}
	}
	return d
}
func (d *Document) Position(offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(d.Source) {
		offset = len(d.Source)
	}
	line := sort.Search(len(d.lineStarts), func(i int) bool { return d.lineStarts[i] > offset }) - 1
	start := d.lineStarts[line]
	column := utf8.RuneCount(d.Source[start:offset])
	// BOM is an encoding marker, not a character in the first logical line.
	if line == 0 && offset >= 3 && len(d.Source) >= 3 && string(d.Source[:3]) == "\xef\xbb\xbf" {
		column--
	}
	return Position{Offset: offset, Line: line, Column: column}
}
func (d *Document) Span(start, end int) Range { return Range{d.Position(start), d.Position(end)} }
func (d *Document) Add(code, phase, message string, r Range, path string) {
	d.Diagnostics = append(d.Diagnostics, Diagnostic{Code: code, Phase: phase, Severity: "error", Message: message, File: d.File, Range: r, Path: path})
}
func HasErrors(ds []Diagnostic) bool {
	for _, d := range ds {
		if d.Severity == "error" {
			return true
		}
	}
	return false
}
func SortDiagnostics(ds []Diagnostic) {
	sort.SliceStable(ds, func(i, j int) bool {
		a, b := ds[i], ds[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Range.Start.Offset != b.Range.Start.Offset {
			return a.Range.Start.Offset < b.Range.Start.Offset
		}
		return a.Code < b.Code
	})
}
func Digest(source []byte) string { sum := sha256.Sum256(source); return hex.EncodeToString(sum[:]) }
func (n *Node) Property(name string) *Property {
	for _, p := range n.Props {
		if p.Name == name {
			return p
		}
	}
	return nil
}
func (n *Node) Get(name string) *Value {
	if p := n.Property(name); p != nil {
		return p.Value
	}
	return nil
}
func (n *Node) Text(name string) string { return Text(n.Get(name)) }
func Text(v *Value) string {
	if v == nil {
		return ""
	}
	return v.Text
}
func Int(v *Value) (int64, bool) {
	if v == nil || v.Tag != "literal" || v.Type != "int" {
		return 0, false
	}
	n, e := strconv.ParseInt(v.Text, 10, 64)
	return n, e == nil
}
func Field(v *Value, name string) *Value {
	if v != nil {
		for _, p := range v.Fields {
			if p.Name == name {
				return p.Value
			}
		}
	}
	return nil
}
func Strings(v *Value) []string {
	var out []string
	if v != nil {
		for _, x := range v.Items {
			out = append(out, Text(x))
		}
	}
	return out
}
func Plain(v *Value) any {
	if v == nil {
		return nil
	}
	switch v.Tag {
	case "literal":
		switch v.Type {
		case "bool":
			return v.Bool
		case "null":
			return nil
		case "int", "decimal":
			return json.Number(v.Text)
		default:
			return v.Text
		}
	case "raw":
		return v.Text
	case "list":
		a := make([]any, 0, len(v.Items))
		for _, x := range v.Items {
			a = append(a, Plain(x))
		}
		return a
	case "record":
		a := map[string]any{}
		for _, x := range v.Fields {
			a[x.Name] = Plain(x.Value)
		}
		return a
	}
	return v
}
func Walk(nodes []*Node) []*Node {
	var out []*Node
	var visit func([]*Node)
	visit = func(ns []*Node) {
		for _, n := range ns {
			out = append(out, n)
			visit(n.Children)
		}
	}
	visit(nodes)
	return out
}
