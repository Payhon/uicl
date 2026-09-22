package uicl

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type parseFault struct {
	code, message string
	start, end    int
}

func syntaxFail(code, message string, start, end int) { panic(parseFault{code, message, start, end}) }

func identStart(c byte) bool { return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '_' }
func identPart(c byte) bool  { return identStart(c) || c >= '0' && c <= '9' }
func digit(c byte) bool      { return c >= '0' && c <= '9' }

// JSON accepts lone escaped surrogates and replaces them. UICL rejects them.
func decodeString(text string) (string, error) {
	for i := 1; i < len(text)-1; i++ {
		if text[i] != '\\' {
			continue
		}
		i++
		if i >= len(text)-1 || text[i] != 'u' {
			continue
		}
		if i+4 >= len(text) {
			return "", fmt.Errorf("incomplete Unicode escape")
		}
		n, e := strconv.ParseUint(text[i+1:i+5], 16, 16)
		if e != nil {
			return "", e
		}
		i += 4
		if n >= 0xdc00 && n <= 0xdfff {
			return "", fmt.Errorf("unpaired Unicode surrogate")
		}
		if n >= 0xd800 && n <= 0xdbff {
			if i+6 >= len(text) || text[i+1:i+3] != "\\u" {
				return "", fmt.Errorf("unpaired Unicode surrogate")
			}
			low, e := strconv.ParseUint(text[i+3:i+7], 16, 16)
			if e != nil || low < 0xdc00 || low > 0xdfff {
				return "", fmt.Errorf("unpaired Unicode surrogate")
			}
			i += 6
		}
	}
	var decoded string
	err := json.Unmarshal([]byte(text), &decoded)
	return decoded, err
}

func scanCore(d *Document, start, end int) []Token {
	var out []Token
	position := d.Position(start)
	emit := func(kind string, a, b int) {
		next := Position{Offset: b, Line: position.Line, Column: position.Column + utf8.RuneCount(d.Source[a:b])}
		out = append(out, Token{Kind: kind, Text: string(d.Source[a:b]), Range: Range{Start: position, End: next}})
		position = next
	}
	for i := start; i < end; {
		a := i
		c := d.Source[i]
		if c == ' ' {
			for i < end && d.Source[i] == ' ' {
				i++
			}
			emit("WS", a, i)
			continue
		}
		if c == '"' {
			i++
			closed := false
			for i < end {
				x := d.Source[i]
				i++
				if x == '"' {
					closed = true
					break
				}
				if x == '\\' && i < end {
					i++
				}
			}
			if !closed {
				syntaxFail("UICL-S005", "Unterminated string", a, i)
			}
			if _, e := decodeString(string(d.Source[a:i])); e != nil {
				syntaxFail("UICL-S005", e.Error(), a, i)
			}
			emit("STRING", a, i)
			continue
		}
		if digit(c) {
			i++
			if c != '0' {
				for i < end && digit(d.Source[i]) {
					i++
				}
			}
			if i < end && d.Source[i] == '.' {
				i++
				if i >= end || !digit(d.Source[i]) {
					syntaxFail("UICL-S003", "Decimal point requires digits", a, i)
				}
				for i < end && digit(d.Source[i]) {
					i++
				}
			}
			if i < end && (d.Source[i] == 'e' || d.Source[i] == 'E') {
				i++
				if i < end && (d.Source[i] == '+' || d.Source[i] == '-') {
					i++
				}
				if i >= end || !digit(d.Source[i]) {
					syntaxFail("UICL-S003", "Exponent requires digits", a, i)
				}
				for i < end && digit(d.Source[i]) {
					i++
				}
			}
			if i < end && (digit(d.Source[i]) || identStart(d.Source[i])) {
				syntaxFail("UICL-S003", "Invalid number or leading zero", a, i+1)
			}
			emit("NUMBER", a, i)
			continue
		}
		if identStart(c) {
			i++
			for i < end && identPart(d.Source[i]) {
				i++
			}
			emit("IDENT", a, i)
			continue
		}
		if i+1 < end {
			pair := string(d.Source[i : i+2])
			switch pair {
			case "::", "==", "!=", "<=", ">=", "??", "?.":
				i += 2
				emit(pair, a, i)
				continue
			}
		}
		if strings.ContainsRune("#@$():,.[]{}=+-*%<>", rune(c)) {
			i++
			emit(string(c), a, i)
			continue
		}
		unexpected, size := utf8.DecodeRune(d.Source[a:end])
		syntaxFail("UICL-S003", fmt.Sprintf("Unexpected character %q", unexpected), a, a+size)
	}
	return out
}

// decimalFits asks whether some exact c * 10^e representation fits the
// specification, rather than incorrectly rejecting equivalent trailing zeros.
func decimalFits(text string) bool {
	t := strings.TrimPrefix(text, "-")
	mantissa := t
	exponent := int64(0)
	if i := strings.IndexAny(t, "eE"); i >= 0 {
		mantissa = t[:i]
		exp, e := strconv.ParseInt(t[i+1:], 10, 64)
		if e != nil {
			return strings.Trim(mantissa, "0.") == ""
		}
		exponent = exp
	}
	if strings.Trim(mantissa, "0.") == "" {
		return true
	}
	// Only coefficient digits can offset an exponent. Bound by the actual
	// source length before arithmetic, including at int64 exponent extremes.
	if exponent < -6176-int64(len(mantissa)) || exponent > 6144+int64(len(mantissa)) {
		return false
	}
	if i := strings.IndexByte(mantissa, '.'); i >= 0 {
		fraction := int64(len(mantissa) - i - 1)
		exponent -= fraction
		mantissa = mantissa[:i] + mantissa[i+1:]
	}
	mantissa = strings.TrimLeft(mantissa, "0")
	if mantissa == "" {
		return true
	}
	trimmed := strings.TrimRight(mantissa, "0")
	zeros := len(mantissa) - len(trimmed)
	exponent += int64(zeros)
	digits := len(trimmed)
	if digits > 34 || exponent < -6176 {
		return false
	}
	// A coefficient may retain zeros to bring a high exponent down to 6111.
	return exponent <= 6111+int64(34-digits)
}

type scalarParser struct {
	d          *Document
	ts         []Token
	pos, depth int
}

func newScalar(d *Document, start, end int) *scalarParser {
	all := scanCore(d, start, end)
	ts := make([]Token, 0, len(all)+1)
	for _, t := range all {
		if t.Kind != "WS" {
			ts = append(ts, t)
		}
	}
	ts = append(ts, Token{Kind: "EOF", Range: d.Span(end, end)})
	return &scalarParser{d: d, ts: ts}
}
func (p *scalarParser) peek() Token         { return p.ts[p.pos] }
func (p *scalarParser) at(kind string) bool { return p.peek().Kind == kind }
func (p *scalarParser) fail(message string) {
	t := p.peek()
	syntaxFail("UICL-S003", message, t.Range.Start.Offset, t.Range.End.Offset)
}
func (p *scalarParser) take(kind string) Token {
	t := p.peek()
	if kind != "" && t.Kind != kind {
		p.fail("Expected " + kind + ", got " + t.Kind)
	}
	if p.pos < len(p.ts)-1 {
		p.pos++
	}
	return t
}
func (p *scalarParser) gap() {
	if p.pos > 0 && p.ts[p.pos-1].Range.End.Offset == p.peek().Range.Start.Offset {
		p.fail("Header fields require a space")
	}
}
func (p *scalarParser) adjacent() {
	if p.pos > 0 && p.ts[p.pos-1].Range.End.Offset != p.peek().Range.Start.Offset {
		p.fail("No whitespace inside a qualified name, anchor or reference")
	}
}
func (p *scalarParser) enter() {
	p.depth++
	if p.depth > MaxDepth {
		t := p.peek()
		syntaxFail("UICL-LIMIT", "Value or expression nesting exceeds limit", t.Range.Start.Offset, t.Range.End.Offset)
	}
}
func (p *scalarParser) qname() (string, Range) {
	t := p.take("IDENT")
	r := t.Range
	for p.at(".") {
		p.adjacent()
		p.take("")
		p.adjacent()
		t = p.take("IDENT")
		r.End = t.Range.End
	}
	return string(p.d.Source[r.Start.Offset:r.End.Offset]), r
}
func (p *scalarParser) key() (string, Range) {
	if p.at("STRING") {
		t := p.take("")
		s, _ := decodeString(t.Text)
		return s, t.Range
	}
	return p.qname()
}
func (p *scalarParser) reference() *Value {
	first := p.take("@")
	p.adjacent()
	v := &Value{Tag: "ref", Range: first.Range}
	if p.at("STRING") {
		t := p.take("")
		s, _ := decodeString(t.Text)
		v.URI = &s
		v.Range.End = t.Range.End
		return v
	}
	t := p.take("IDENT")
	v.ID = t.Text
	v.Range.End = t.Range.End
	if p.at("::") {
		p.adjacent()
		p.take("")
		p.adjacent()
		t = p.take("IDENT")
		v.Module = v.ID
		v.ID = t.Text
		v.Range.End = t.Range.End
	}
	for p.at(".") {
		p.adjacent()
		p.take("")
		p.adjacent()
		t = p.take("IDENT")
		v.Ports = append(v.Ports, t.Text)
		v.Range.End = t.Range.End
	}
	return v
}
func (p *scalarParser) binding() *Value {
	t := p.take("$")
	p.adjacent()
	name := p.take("IDENT")
	return p.postfix(&Value{Tag: "binding", Name: name.Text, Range: Range{Start: t.Range.Start, End: name.Range.End}})
}
func (p *scalarParser) postfix(v *Value) *Value {
	count := 0
	for p.at(".") || p.at("?.") || p.at("[") {
		count++
		if count > MaxDepth {
			t := p.peek()
			syntaxFail("UICL-LIMIT", "Member or index nesting exceeds limit", t.Range.Start.Offset, t.Range.End.Offset)
		}
		t := p.take("")
		if t.Kind == "[" {
			idx := p.expr(0)
			end := p.take("]")
			v = &Value{Tag: "index", Object: v, Index: idx, Range: Range{Start: v.Range.Start, End: end.Range.End}}
		} else {
			name := p.take("IDENT")
			v = &Value{Tag: "member", Object: v, Name: name.Text, Optional: t.Kind == "?.", Range: Range{Start: v.Range.Start, End: name.Range.End}}
		}
	}
	return v
}
func (p *scalarParser) number(negative bool) *Value {
	startPosition := p.peek().Range.Start
	start := startPosition.Offset
	if negative {
		p.take("-")
	}
	t := p.take("NUMBER")
	s := t.Text
	if negative {
		s = "-" + s
	}
	typ := "int"
	if strings.ContainsAny(s, ".eE") {
		typ = "decimal"
		if !decimalFits(s) {
			syntaxFail("UICL-NUMBER_RANGE", "Decimal exceeds exact 34-digit coefficient or exponent bounds", start, t.Range.End.Offset)
		}
	} else {
		if _, e := strconv.ParseInt(s, 10, 64); e != nil {
			syntaxFail("UICL-NUMBER_RANGE", "Integer exceeds signed 64-bit range", start, t.Range.End.Offset)
		}
	}
	return &Value{Tag: "literal", Type: typ, Text: s, Range: Range{Start: startPosition, End: t.Range.End}}
}
func (p *scalarParser) pair(expression bool) *Property {
	name, r := p.key()
	p.take(":")
	var v *Value
	if expression {
		v = p.expr(0)
	} else {
		v = p.value()
	}
	return &Property{Name: name, NameRange: r, Value: v, Range: Range{Start: r.Start, End: v.Range.End}}
}
func (p *scalarParser) aggregate(record, expression bool) *Value {
	open, close, tag := "[", "]", "list"
	if record {
		open, close, tag = "{", "}", "record"
	}
	start := p.take(open)
	v := &Value{Tag: tag, Range: start.Range}
	seen := map[string]bool{}
	if !p.at(close) {
		for {
			if record {
				prop := p.pair(expression)
				if seen[prop.Name] {
					syntaxFail("UICL-DUPLICATE_KEY", "Duplicate record key "+prop.Name, prop.NameRange.Start.Offset, prop.NameRange.End.Offset)
				}
				seen[prop.Name] = true
				v.Fields = append(v.Fields, prop)
			} else {
				var item *Value
				if expression {
					item = p.expr(0)
				} else {
					item = p.value()
				}
				v.Items = append(v.Items, item)
			}
			if !p.at(",") {
				break
			}
			p.take("")
			if p.at(close) {
				p.fail("Trailing comma is forbidden")
			}
		}
	}
	v.Range.End = p.take(close).Range.End
	return v
}
func (p *scalarParser) value() *Value {
	p.enter()
	defer func() { p.depth-- }()
	t := p.peek()
	switch t.Kind {
	case "=":
		p.take("")
		tree := p.expr(0)
		return &Value{Tag: "expression", Tree: tree, Range: Range{Start: t.Range.Start, End: tree.Range.End}}
	case "$":
		tree := p.binding()
		return &Value{Tag: "expression", Tree: tree, Range: tree.Range}
	case "@":
		return p.reference()
	case "STRING":
		p.take("")
		s, _ := decodeString(t.Text)
		return &Value{Tag: "literal", Type: "text", Text: s, Range: t.Range}
	case "NUMBER", "-":
		return p.number(t.Kind == "-")
	case "IDENT":
		if t.Text == "true" || t.Text == "false" || t.Text == "null" {
			p.take("")
			typ := "bool"
			if t.Text == "null" {
				typ = "null"
			}
			return &Value{Tag: "literal", Type: typ, Bool: t.Text == "true", Range: t.Range}
		}
		if t.Text == "NaN" || t.Text == "Infinity" {
			p.fail("NaN and Infinity are not Core values")
		}
		s, r := p.qname()
		return &Value{Tag: "literal", Type: "text", Text: s, Range: r}
	case "[":
		return p.aggregate(false, false)
	case "{":
		return p.aggregate(true, false)
	}
	p.fail("Expected value")
	return nil
}
func binaryPower(s string) (int, bool) {
	switch s {
	case "??":
		return 10, false
	case "or":
		return 20, false
	case "and":
		return 30, false
	case "==", "!=", "<", "<=", ">", ">=", "in":
		return 40, true
	case "+", "-":
		return 50, false
	case "*", "%":
		return 60, false
	}
	return 0, false
}
func (p *scalarParser) expr(minimum int) *Value {
	p.enter()
	defer func() { p.depth-- }()
	t := p.peek()
	var left *Value
	switch {
	case t.Kind == "-":
		// A signed numeric literal includes MinInt64; never validate its positive
		// magnitude as an independent int64 before applying the lexical minus.
		if p.pos+1 < len(p.ts) && p.ts[p.pos+1].Kind == "NUMBER" {
			left = p.number(true)
		} else {
			p.take("")
			operand := p.expr(70)
			left = &Value{Tag: "unary", Op: "-", Right: operand, Range: Range{Start: t.Range.Start, End: operand.Range.End}}
		}
	case t.Text == "not":
		if minimum > 35 {
			p.fail("Boolean not requires parentheses in an arithmetic or comparison operand")
		}
		p.take("")
		operand := p.expr(35)
		left = &Value{Tag: "unary", Op: "not", Right: operand, Range: Range{Start: t.Range.Start, End: operand.Range.End}}
	case t.Kind == "(":
		p.take("")
		left = p.expr(0)
		end := p.take(")")
		left.Range = Range{Start: t.Range.Start, End: end.Range.End}
	case t.Kind == "$":
		left = p.binding()
	case t.Kind == "@":
		left = p.reference()
	case t.Kind == "[":
		left = p.aggregate(false, true)
	case t.Kind == "{":
		left = p.aggregate(true, true)
	case t.Kind == "STRING" || t.Kind == "NUMBER" || t.Text == "true" || t.Text == "false" || t.Text == "null":
		left = p.value()
	case t.Kind == "IDENT" && t.Text != "and" && t.Text != "or" && t.Text != "in":
		name, r := p.qname()
		p.take("(")
		args := []*Value{}
		if !p.at(")") {
			for {
				args = append(args, p.expr(0))
				if !p.at(",") {
					break
				}
				p.take("")
				if p.at(")") {
					p.fail("Trailing comma is forbidden")
				}
			}
		}
		end := p.take(")")
		left = &Value{Tag: "pure_call", Name: name, Args: args, Range: Range{Start: r.Start, End: end.Range.End}}
	default:
		p.fail("Expected expression; variables require $")
	}
	left = p.postfix(left)
	compared := false
	operators := 0
	for {
		t = p.peek()
		bp, comparison := binaryPower(t.Text)
		if bp == 0 || bp < minimum {
			break
		}
		operators++
		if operators > MaxDepth {
			syntaxFail("UICL-LIMIT", "Binary expression nesting exceeds limit", t.Range.Start.Offset, t.Range.End.Offset)
		}
		if comparison {
			if compared {
				p.fail("Comparison chains require explicit boolean operators")
			}
			compared = true
		}
		p.take("")
		next := bp + 1
		if t.Text == "??" {
			next = bp
		}
		right := p.expr(next)
		left = &Value{Tag: "binary", Op: t.Text, Left: left, Right: right, Range: Range{Start: left.Range.Start, End: right.Range.End}}
	}
	return left
}
