package uicl

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"unicode/utf8"
)

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// ApplyEdits applies byte ranges against exactly one expected source revision.
// Edits are simultaneous, not offsets into the result of preceding edits.
func ApplyEdits(source []byte, expectedDigest string, edits []TextEdit) ([]byte, error) {
	if expectedDigest == "" || Digest(source) != expectedDigest {
		return nil, fmt.Errorf("CONFLICT: source digest differs from the expected revision")
	}
	if !utf8.Valid(source) {
		return nil, fmt.Errorf("EDIT_UTF8: source must be valid UTF-8")
	}
	ordered := append([]TextEdit(nil), edits...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Range.Start.Offset == ordered[j].Range.Start.Offset {
			return ordered[i].Range.End.Offset < ordered[j].Range.End.Offset
		}
		return ordered[i].Range.Start.Offset < ordered[j].Range.Start.Offset
	})
	end, lastStart := 0, -1
	for _, edit := range ordered {
		a, b := edit.Range.Start.Offset, edit.Range.End.Offset
		if a < 0 || b < a || b > len(source) {
			return nil, fmt.Errorf("EDIT_RANGE: edit lies outside the source")
		}
		if a < end || a == lastStart {
			return nil, fmt.Errorf("EDIT_OVERLAP: overlapping edits are not allowed")
		}
		if a < len(source) && !utf8.RuneStart(source[a]) || b < len(source) && !utf8.RuneStart(source[b]) || !utf8.ValidString(edit.NewText) {
			return nil, fmt.Errorf("EDIT_UTF8: edits must preserve UTF-8 character boundaries")
		}
		end, lastStart = b, a
	}
	var out bytes.Buffer
	end = 0
	for _, edit := range ordered {
		out.Write(source[end:edit.Range.Start.Offset])
		out.WriteString(edit.NewText)
		end = edit.Range.End.Offset
	}
	out.Write(source[end:])
	return out.Bytes(), nil
}

// Format is conservative: only structural horizontal spacing changes. Valid
// indentation, trivia, comments, all literal lexemes and raw text stay intact.
func Format(doc *Document) ([]byte, []Diagnostic) {
	if HasErrors(doc.Diagnostics) {
		return nil, doc.Diagnostics
	}
	var edits []TextEdit
	if doc.Mode == "structured" {
		edits = formatCoreEdits(doc)
	} else {
		for _, is := range doc.Islands {
			formatted, err := ApplyEdits(is.Document.Source, Digest(is.Document.Source), formatCoreEdits(is.Document))
			if err != nil {
				return nil, []Diagnostic{formatDiagnostic(doc, err.Error())}
			}
			if bytes.Equal(formatted, is.Document.Source) {
				continue
			}
			if is.Encoding != "" {
				formatted = []byte(base64.RawURLEncoding.EncodeToString(formatted))
			}
			edits = append(edits, TextEdit{Range: is.PayloadRange, NewText: string(formatted)})
		}
	}
	result, err := ApplyEdits(doc.Source, Digest(doc.Source), edits)
	if err != nil {
		return nil, []Diagnostic{formatDiagnostic(doc, err.Error())}
	}
	if bytes.Equal(result, doc.Source) {
		return result, nil
	}
	var checked *Document
	if doc.Mode == "structured" && isFragment(doc) {
		checked = ParseCore(doc.File, result, ParseOptions{Fragment: true})
	} else {
		checked = Parse(doc.File, result, ParseOptions{})
	}
	if HasErrors(checked.Diagnostics) {
		return nil, []Diagnostic{formatDiagnostic(doc, "Formatter output failed syntax validation")}
	}
	if !bytes.Equal(semanticNodes(doc.Nodes), semanticNodes(checked.Nodes)) {
		return nil, []Diagnostic{formatDiagnostic(doc, "Formatter output changed the document semantics")}
	}
	return result, nil
}

func isFragment(d *Document) bool {
	for _, token := range d.Tokens {
		if token.Kind == "WS" || token.Kind == "NL" || token.Kind == "COMMENT" || token.Kind == "BOM" {
			continue
		}
		return token.Text != "uicl"
	}
	return true
}
func formatDiagnostic(d *Document, message string) Diagnostic {
	return Diagnostic{Code: "FORMAT_SEMANTICS", Phase: "format", Severity: "error", Message: message, File: d.File, Range: d.Span(0, 0)}
}
func protectedToken(t Token) bool {
	return t.Kind == "NL" || t.Kind == "COMMENT" || t.Kind == "BOM" || t.Kind == "RAW"
}
func formatCoreEdits(d *Document) []TextEdit {
	var edits []TextEdit
	tokens := d.Tokens
	for i, token := range tokens {
		if token.Kind == "WS" {
			// Keep line-leading/trailing whitespace and whitespace-only lines.
			if token.Range.Start.Offset == 0 || d.Source[token.Range.Start.Offset-1] == '\n' {
				continue
			}
			if i == 0 || i == len(tokens)-1 {
				continue
			}
			previous, next := tokens[i-1], tokens[i+1]
			if previous.Range.End.Line != next.Range.Start.Line || previous.Kind == "NL" || next.Kind == "NL" || previous.Kind == "BOM" || previous.Kind == "COMMENT" || next.Kind == "COMMENT" {
				continue
			}
			space := " "
			if previous.Kind == "[" || previous.Kind == "{" || previous.Kind == "(" || next.Kind == "]" || next.Kind == "}" || next.Kind == ")" || next.Kind == "," || next.Kind == ":" {
				space = ""
			}
			if token.Text != space {
				edits = append(edits, TextEdit{Range: token.Range, NewText: space})
			}
		} else if (token.Kind == ":" || token.Kind == ",") && i+1 < len(tokens) {
			next := tokens[i+1]
			if next.Kind != "WS" && !protectedToken(next) && next.Kind != "]" && next.Kind != "}" && next.Kind != ")" && token.Range.End.Offset == next.Range.Start.Offset {
				r := Range{Start: token.Range.End, End: token.Range.End}
				edits = append(edits, TextEdit{Range: r, NewText: " "})
			}
		}
	}
	return edits
}

func semanticNodes(nodes []*Node) []byte {
	// Ignore source ranges only; retain ordered properties/children, expression
	// operators, exact numbers, raw content and literal types in this safety gate.
	b, _ := json.Marshal(nodes)
	var value any
	_ = json.Unmarshal(b, &value)
	var strip func(any)
	strip = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			delete(x, "range")
			delete(x, "nameRange")
			delete(x, "kindRange")
			delete(x, "idRange")
			for _, child := range x {
				strip(child)
			}
		case []any:
			for _, child := range x {
				strip(child)
			}
		}
	}
	strip(value)
	b, _ = json.Marshal(value)
	return b
}
