package uicl

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

type coreLine struct{ start, content, end, next, indent int }
type parseStop struct{}
type coreReader struct {
	d         *Document
	opts      ParseOptions
	lines     []coreLine
	i, depth  int
	rawRanges []Range
}

// ParseCore parses a structured document or an explicitly selected host fragment.
// In Recover mode it retains valid siblings and continues only at indentation
// boundaries; diagnostics always remain errors.
func ParseCore(file string, source []byte, opts ParseOptions) *Document {
	d := NewDocument(file, source)
	if len(source) > MaxBytes {
		d.Add("UICL-LIMIT", "syntax", "Source byte limit exceeded", d.Span(0, len(source)), "")
		return d
	}
	if !utf8.Valid(source) {
		d.Add("UICL-S005", "syntax", "Source must be valid UTF-8", d.Span(0, len(source)), "")
		return d
	}
	for i, c := range source {
		if c == 0 || (c == '\r' && (i+1 == len(source) || source[i+1] != '\n')) {
			d.Add("UICL-S005", "syntax", "NUL or bare CR is forbidden", d.Span(i, i+1), "")
			return d
		}
	}
	r := &coreReader{d: d, opts: opts}
	for start := 0; start < len(source); {
		end := start
		for end < len(source) && source[end] != '\n' {
			end++
		}
		next := end
		if next < len(source) {
			next++
		}
		if end > start && source[end-1] == '\r' {
			end--
		}
		content := start
		if start == 0 && bytes.HasPrefix(source, []byte{0xef, 0xbb, 0xbf}) {
			content = 3
		}
		base := content
		for content < end && source[content] == ' ' {
			content++
		}
		r.lines = append(r.lines, coreLine{start, content, end, next, content - base})
		start = next
	}
	r.run()
	r.collectTokens()
	return d
}

func (r *coreReader) report(f parseFault) {
	r.d.Add(f.code, "syntax", f.message, r.d.Span(f.start, f.end), "")
}
func (r *coreReader) attempt(fn func()) (ok bool) {
	defer func() {
		if e := recover(); e != nil {
			if f, yes := e.(parseFault); yes {
				r.report(f)
			} else if _, stop := e.(parseStop); stop {
				// The innermost entry already recorded its diagnostic.
			} else {
				panic(e)
			}
		}
	}()
	fn()
	return true
}
func (r *coreReader) lineText(l coreLine) string        { return string(r.d.Source[l.content:l.end]) }
func (r *coreReader) fail(code, msg string, l coreLine) { syntaxFail(code, msg, l.content, l.end) }
func (r *coreReader) row() (coreLine, bool) {
	for r.i < len(r.lines) {
		l := r.lines[r.i]
		text := r.lineText(l)
		if text == "" || strings.HasPrefix(text, "//") {
			r.i++
			continue
		}
		if strings.ContainsRune(text, '\t') {
			r.i++
			r.fail("UICL-LAYOUT", "Tab outside raw content", l)
		}
		if l.indent%2 != 0 {
			r.i++
			r.fail("UICL-LAYOUT", "Indent must use two-space units", l)
		}
		if l.indent/2 > MaxDepth {
			r.i++
			r.fail("UICL-LIMIT", "Indentation depth exceeds limit", l)
		}
		return l, true
	}
	return coreLine{}, false
}

// synchronize skips only the damaged entry's descendants. Raw bodies were
// consumed by raw() before this boundary and cannot become declarations.
func (r *coreReader) synchronize(start, indent int) {
	if r.i <= start {
		r.i = start + 1
	}
	for r.i < len(r.lines) {
		l := r.lines[r.i]
		text := r.lineText(l)
		if text == "" || strings.HasPrefix(text, "//") {
			r.i++
			continue
		}
		if l.indent <= indent {
			return
		}
		r.i++
	}
}
func (r *coreReader) run() {
	if !r.opts.Fragment {
		start := r.i
		ok := r.attempt(func() {
			l, exists := r.row()
			if !exists {
				syntaxFail("UICL-VERSION", "Missing version header", 0, 0)
			}
			if l.indent != 0 {
				r.fail("UICL-VERSION", "Version header must be at column zero", l)
			}
			p := newScalar(r.d, l.content, l.end)
			t := p.take("IDENT")
			if t.Text != "uicl" {
				r.fail("UICL-VERSION", "Expected uicl version header", l)
			}
			p.gap()
			v := p.take("STRING")
			version, _ := decodeString(v.Text)
			if version != CoreVersion {
				r.fail("UICL-VERSION", "Unsupported UICL version", l)
			}
			p.take("EOF")
			r.i++
		})
		if !ok {
			if !r.opts.Recover {
				return
			}
			r.synchronize(start, 0)
		}
	}
	for r.i < len(r.lines) {
		start := r.i
		ok := r.attempt(func() {
			l, exists := r.row()
			if !exists {
				return
			}
			start = r.i
			if l.indent != 0 {
				r.fail("UICL-LAYOUT", "Top-level declaration must be at column zero", l)
			}
			r.d.Nodes = append(r.d.Nodes, r.node(0))
		})
		if !ok {
			if !r.opts.Recover {
				return
			}
			r.synchronize(start, 0)
		}
	}
	if len(r.d.Nodes) == 0 && !HasErrors(r.d.Diagnostics) {
		r.d.Add("UICL-S003", "syntax", "At least one declaration required", r.d.Span(len(r.d.Source), len(r.d.Source)), "")
	}
}

// propertyPrefix scans only the key and colon. A raw marker is not an inline
// token and must not be scanned before we choose the value's layout mode.
func (r *coreReader) propertyPrefix(l coreLine) (name string, key Range, tail int, ok bool) {
	s := r.d.Source
	i := l.content
	if i >= l.end {
		return
	}
	if s[i] == '"' {
		i++
		closed := false
		for i < l.end {
			c := s[i]
			i++
			if c == '"' {
				closed = true
				break
			}
			if c == '\\' && i < l.end {
				i++
			}
		}
		if !closed {
			return
		}
	} else {
		if !identStart(s[i]) {
			return
		}
		i++
		for i < l.end && identPart(s[i]) {
			i++
		}
		for i < l.end && s[i] == '.' {
			i++
			if i >= l.end || !identStart(s[i]) {
				return
			}
			i++
			for i < l.end && identPart(s[i]) {
				i++
			}
		}
	}
	keyEnd := i
	for i < l.end && s[i] == ' ' {
		i++
	}
	if i >= l.end || s[i] != ':' || i+1 < l.end && s[i+1] == ':' {
		return
	}
	p := newScalar(r.d, l.content, keyEnd)
	name, key = p.key()
	p.take("EOF")
	i++
	for i < l.end && s[i] == ' ' {
		i++
	}
	return name, key, i, true
}
func (r *coreReader) node(indent int) *Node {
	l, ok := r.row()
	if !ok {
		syntaxFail("UICL-S003", "Expected node", len(r.d.Source), len(r.d.Source))
	}
	if l.indent != indent {
		r.fail("UICL-LAYOUT", "Node indentation mismatch", l)
	}
	p := newScalar(r.d, l.content, l.end)
	kind, kr := p.qname()
	n := &Node{Kind: kind, KindRange: kr, Range: r.d.Span(l.content, l.end), Props: []*Property{}, Children: []*Node{}}
	if p.at("#") {
		p.gap()
		begin := p.take("")
		p.adjacent()
		id := p.take("IDENT")
		n.ID = id.Text
		n.IDRange = r.d.Span(begin.Range.Start.Offset, id.Range.End.Offset)
	}
	if p.at("IDENT") || p.at("STRING") {
		p.gap()
		if p.at("STRING") {
			t := p.take("")
			label, _ := decodeString(t.Text)
			n.Label = &Value{Tag: "literal", Type: "text", Text: label, Range: t.Range}
		} else {
			text, rng := p.qname()
			n.Label = &Value{Tag: "literal", Type: "text", Text: text, Range: rng}
		}
	}
	seen := map[string]bool{}
	add := func(prop *Property) {
		if seen[prop.Name] {
			syntaxFail("UICL-DUPLICATE_KEY", "Duplicate property "+prop.Name, prop.NameRange.Start.Offset, prop.NameRange.End.Offset)
		}
		seen[prop.Name] = true
		n.Props = append(n.Props, prop)
		if prop.Range.End.Offset > n.Range.End.Offset {
			n.Range.End = prop.Range.End
		}
	}
	if p.at("(") {
		p.gap()
		p.take("")
		if !p.at(")") {
			for {
				add(p.pair(false))
				if !p.at(",") {
					break
				}
				p.take("")
				if p.at(")") {
					p.fail("Trailing comma is forbidden")
				}
			}
		}
		p.take(")")
	}
	p.take("EOF")
	r.i++
	for r.i < len(r.lines) {
		start := r.i
		done := false
		entryIndent := indent + 2
		ok := r.attempt(func() {
			child, exists := r.row()
			if !exists || child.indent <= indent {
				done = true
				return
			}
			start = r.i
			entryIndent = child.indent
			if child.indent != indent+2 {
				r.fail("UICL-LAYOUT", "Skipped node indentation", child)
			}
			if _, _, _, isProp := r.propertyPrefix(child); isProp {
				add(r.property(child))
			} else {
				c := r.node(indent + 2)
				n.Children = append(n.Children, c)
				n.Range.End = c.Range.End
			}
		})
		if done {
			break
		}
		if !ok {
			if !r.opts.Recover {
				panic(parseStop{})
			}
			r.synchronize(start, entryIndent)
		}
	}
	return n
}
func (r *coreReader) property(l coreLine) *Property {
	name, key, tail, ok := r.propertyPrefix(l)
	if !ok {
		r.fail("UICL-S003", "Expected record key and colon", l)
	}
	r.i++
	v := r.tailValue(l, tail, l.indent)
	return &Property{Name: name, NameRange: key, Value: v, Range: r.d.Span(key.Start.Offset, v.Range.End.Offset)}
}
func (r *coreReader) tailValue(l coreLine, tail, indent int) *Value {
	end := l.end
	for end > tail && r.d.Source[end-1] == ' ' {
		end--
	}
	if tail == end {
		next, ok := r.row()
		if !ok || next.indent != indent+2 {
			r.fail("UICL-MISSING_VALUE", "Missing block value", l)
		}
		return r.block(indent + 2)
	}
	text := string(r.d.Source[tail:end])
	if text == "|" || strings.HasPrefix(text, "```") {
		return r.raw(text, tail, indent, l)
	}
	p := newScalar(r.d, tail, end)
	v := p.value()
	p.take("EOF")
	return v
}
func (r *coreReader) block(indent int) *Value {
	r.depth++
	defer func() { r.depth-- }()
	if r.depth > MaxDepth {
		l, _ := r.row()
		r.fail("UICL-LIMIT", "Block nesting exceeds limit", l)
	}
	l, ok := r.row()
	if !ok || l.indent != indent {
		r.fail("UICL-LAYOUT", "Block indentation mismatch", l)
	}
	text := r.lineText(l)
	if text == "-" || strings.HasPrefix(text, "- ") {
		return r.blockList(indent)
	}
	return r.record(indent, nil)
}
func (r *coreReader) record(indent int, first *coreLine) *Value {
	v := &Value{Tag: "record", Fields: []*Property{}}
	seen := map[string]bool{}
	for {
		var l coreLine
		var ok bool
		if first != nil {
			l = *first
			ok = true
			first = nil
		} else {
			l, ok = r.row()
		}
		if !ok || l.indent < indent {
			break
		}
		if l.indent != indent {
			r.fail("UICL-LAYOUT", "Unexpected record indentation", l)
		}
		p := r.property(l)
		if seen[p.Name] {
			syntaxFail("UICL-DUPLICATE_KEY", "Duplicate record key "+p.Name, p.NameRange.Start.Offset, p.NameRange.End.Offset)
		}
		seen[p.Name] = true
		if len(v.Fields) == 0 {
			v.Range.Start = p.Range.Start
		}
		v.Range.End = p.Range.End
		v.Fields = append(v.Fields, p)
	}
	if len(v.Fields) == 0 {
		syntaxFail("UICL-MISSING_VALUE", "Empty block record", len(r.d.Source), len(r.d.Source))
	}
	return v
}
func (r *coreReader) blockList(indent int) *Value {
	v := &Value{Tag: "list", Items: []*Value{}}
	for {
		l, ok := r.row()
		if !ok || l.indent < indent {
			break
		}
		if l.indent != indent {
			r.fail("UICL-LAYOUT", "Unexpected list indentation", l)
		}
		text := r.lineText(l)
		if text != "-" && !strings.HasPrefix(text, "- ") {
			r.fail("UICL-LAYOUT", "Mixed list and record", l)
		}
		tail := l.content + 1
		for tail < l.end && r.d.Source[tail] == ' ' {
			tail++
		}
		virtual := l
		virtual.content = tail
		virtual.indent = indent + 2
		var item *Value
		if _, _, _, record := r.propertyPrefix(virtual); record {
			item = r.record(indent+2, &virtual)
		} else {
			r.i++
			item = r.tailValue(l, tail, indent)
		}
		if len(v.Items) == 0 {
			v.Range.Start = r.d.Position(l.content)
		}
		v.Range.End = item.Range.End
		v.Items = append(v.Items, item)
	}
	return v
}
func (r *coreReader) raw(marker string, start, indent int, opening coreLine) *Value {
	v := &Value{Tag: "raw", Form: "pipe", Range: r.d.Span(start, opening.end)}
	prefix := strings.Repeat(" ", indent+2)
	var content []string
	rawEnd := opening.next
	if marker == "|" {
		for r.i < len(r.lines) {
			l := r.lines[r.i]
			text := string(r.d.Source[l.start:l.end])
			if strings.Trim(text, " ") == "" {
				content = append(content, "")
				rawEnd = l.next
				r.i++
				continue
			}
			if !strings.HasPrefix(text, prefix) {
				break
			}
			content = append(content, text[len(prefix):])
			rawEnd = l.next
			r.i++
		}
		body := strings.Join(content, "\n")
		if strings.TrimRight(body, "\n") == "" {
			r.fail("UICL-RAW", "Use an empty string for empty text", opening)
		}
		v.Text = strings.TrimRight(body, "\n") + "\n"
	} else {
		v.Form = "fence"
		fenceLen := 0
		for fenceLen < len(marker) && marker[fenceLen] == '`' {
			fenceLen++
		}
		tag := marker[fenceLen:]
		var invalid *parseFault
		if fenceLen < 3 {
			r.fail("UICL-RAW", "Invalid raw fence", opening)
		}
		if tag != "" {
			if !identStart(tag[0]) {
				invalid = &parseFault{"UICL-RAW", "Invalid raw language tag", opening.content, opening.end}
			}
			for i := 1; i < len(tag); i++ {
				if !identPart(tag[i]) && tag[i] != '.' {
					invalid = &parseFault{"UICL-RAW", "Invalid raw language tag", opening.content, opening.end}
				}
			}
		}
		v.Language = tag
		closing := strings.Repeat(" ", indent) + marker[:fenceLen]
		closed := false
		for r.i < len(r.lines) {
			l := r.lines[r.i]
			text := string(r.d.Source[l.start:l.end])
			if text == closing {
				rawEnd = l.next
				r.i++
				closed = true
				break
			}
			if text == "" {
				content = append(content, "")
			} else if strings.HasPrefix(text, prefix) {
				content = append(content, text[len(prefix):])
			} else {
				if invalid == nil {
					invalid = &parseFault{"UICL-RAW", "Raw content or closing fence indentation mismatch", l.content, l.end}
				}
				// Keep malformed same-level payload inert until its recognizable
				// closing fence; an enclosing dedent remains a recovery boundary.
				if l.indent < indent {
					break
				}
			}
			rawEnd = l.next
			r.i++
		}
		if invalid != nil {
			r.rawRanges = append(r.rawRanges, r.d.Span(start, rawEnd))
			panic(*invalid)
		}
		if !closed {
			r.rawRanges = append(r.rawRanges, r.d.Span(start, rawEnd))
			r.fail("UICL-RAW", "Unclosed raw fence", opening)
		}
		if len(content) > 0 {
			v.Text = strings.Join(content, "\n") + "\n"
		}
	}
	v.Range.End = r.d.Position(rawEnd)
	r.rawRanges = append(r.rawRanges, v.Range)
	return v
}

// Tokens retain every source byte, including trivia. RAW is a protected token
// spanning its marker, body and closing fence, so a formatter cannot edit it.
func (r *coreReader) collectTokens() {
	emit := func(kind string, a, b int) {
		if b > a {
			r.d.Tokens = append(r.d.Tokens, Token{Kind: kind, Text: string(r.d.Source[a:b]), Range: r.d.Span(a, b)})
		}
	}
	i := 0
	rawIndex := 0
	for _, line := range r.lines {
		if i >= line.next {
			continue
		}
		if i < line.start {
			i = line.start
		}
		if i == 0 && bytes.HasPrefix(r.d.Source, []byte{0xef, 0xbb, 0xbf}) {
			emit("BOM", 0, 3)
			i = 3
		}
		if i >= line.end {
			emit("NL", i, line.next)
			i = line.next
			continue
		}
		if line.content >= i && strings.HasPrefix(r.lineText(line), "//") {
			emit("WS", i, line.content)
			emit("COMMENT", line.content, line.end)
			emit("NL", line.end, line.next)
			i = line.next
			continue
		}
		end := line.end
		for rawIndex < len(r.rawRanges) && r.rawRanges[rawIndex].End.Offset <= i {
			rawIndex++
		}
		if rawIndex < len(r.rawRanges) && r.rawRanges[rawIndex].Start.Offset < end {
			end = r.rawRanges[rawIndex].Start.Offset
		}
		var ts []Token
		func() {
			defer func() {
				if e := recover(); e != nil {
					if _, yes := e.(parseFault); !yes {
						panic(e)
					}
					ts = []Token{{Kind: "INVALID", Text: string(r.d.Source[i:end]), Range: r.d.Span(i, end)}}
				}
			}()
			ts = scanCore(r.d, i, end)
		}()
		r.d.Tokens = append(r.d.Tokens, ts...)
		i = end
		if rawIndex < len(r.rawRanges) && r.rawRanges[rawIndex].Start.Offset == i {
			rng := r.rawRanges[rawIndex]
			emit("RAW", i, rng.End.Offset)
			i = rng.End.Offset
			rawIndex++
			continue
		}
		emit("NL", i, line.next)
		i = line.next
	}
}
