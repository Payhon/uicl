package uicl

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"golang.org/x/net/html"
)

// Island keeps both the original carrier range and the decoded Core document.
// Plain payload offsets translate by PayloadRange.Start.Offset; encoded offsets
// deliberately map to the whole carrier instead of pretending to be byte-exact.
type Island struct {
	Range        Range     `json:"range"`
	PayloadRange Range     `json:"payloadRange"`
	Encoding     string    `json:"encoding,omitempty"`
	Complete     bool      `json:"complete"`
	TargetRange  *Range    `json:"targetRange,omitempty"`
	Document     *Document `json:"document"`
}

// AttachDecodedRanges adds honest decoded locations to diagnostics produced by
// downstream validators. Ambiguous semantic paths retain the payload range.
func (d *Document) AttachDecodedRanges(diagnostics []Diagnostic) {
	for _, is := range d.Islands {
		if is.Encoding == "" {
			continue
		}
		paths := map[string][]Range{}
		for _, n := range Walk(is.Document.Nodes) {
			path := n.Kind
			if n.ID != "" {
				path += "#" + n.ID
			}
			paths[path] = append(paths[path], n.Range)
			for _, p := range n.Props {
				r := p.Range
				if p.Value != nil {
					r = p.Value.Range
				}
				paths[path+"."+p.Name] = append(paths[path+"."+p.Name], r)
			}
		}
		for i := range diagnostics {
			diag := &diagnostics[i]
			if diag.File != d.File || diag.DecodedRange != nil || diag.Range != is.Range {
				continue
			}
			r := is.Document.Span(0, len(is.Document.Source))
			if matches := paths[diag.Path]; len(matches) == 1 {
				r = matches[0]
			}
			diag.DecodedRange = &r
		}
	}
}

// Parse detects the declared source mode. An absent or invalid marker is an
// error, never implicit permission to reinterpret damaged Core as Markdown.
func Parse(file string, source []byte, options ParseOptions) *Document {
	if options.Fragment {
		return ParseCore(file, source, options)
	}
	d := NewDocument(file, source)
	if len(source) > MaxBytes {
		d.Add("HOST_LIMIT", "syntax", "Document exceeds the 16 MiB limit", d.Span(0, len(source)), "")
		return d
	}
	if !utf8.Valid(source) {
		d.Add("SYNTAX", "syntax", "Source must be valid UTF-8", d.Span(0, len(source)), "")
		return d
	}
	bom := 0
	if bytes.HasPrefix(source, []byte("\xef\xbb\xbf")) {
		bom = 3
	}
	s := bytes.TrimSpace(source[bom:])
	if bytes.HasPrefix(s, []byte("uicl")) || !bytes.HasPrefix(s, []byte("<")) {
		return ParseCore(file, source, options)
	}
	if bytes.HasPrefix(source[bom:], []byte("<!-- uicl:markdown")) {
		d.Mode = "markdown"
		parseMarkdown(d, bom, options)
	} else {
		d.Mode = "html"
		parseHTML(d, options)
	}
	fillHostTokens(d)
	SortDiagnostics(d.Diagnostics)
	return d
}

func decodeHostPayload(source []byte) ([]byte, error) {
	compact := make([]byte, 0, len(source))
	for _, b := range source {
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			continue
		}
		if !(b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || b == '_' || b == '-') {
			return nil, fmt.Errorf("payload requires the unpadded base64url alphabet")
		}
		compact = append(compact, b)
	}
	if len(compact) > MaxBytes {
		return nil, fmt.Errorf("encoded payload exceeds the byte limit")
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(string(compact))
	if err != nil || base64.RawURLEncoding.EncodeToString(decoded) != string(compact) {
		return nil, fmt.Errorf("invalid or noncanonical base64url payload")
	}
	if !utf8.Valid(decoded) {
		return nil, fmt.Errorf("decoded payload must be valid UTF-8")
	}
	return decoded, nil
}

// commentPayload only recognizes explicit UICL annotation introducers.
func commentPayload(raw []byte) (offset, end int, encoding string, recognized bool, err error) {
	if !bytes.HasPrefix(raw, []byte("<!--")) {
		return
	}
	start := 4
	for start < len(raw) && (raw[start] == ' ' || raw[start] == '\t') {
		start++
	}
	rest := raw[start:]
	prefix := "uicl"
	if bytes.HasPrefix(rest, []byte("uicl:base64url")) {
		prefix = "uicl:base64url"
		encoding = "base64url"
	}
	if !bytes.HasPrefix(rest, []byte(prefix)) {
		return
	}
	start += len(prefix)
	if start < len(raw) && raw[start] != '\r' && raw[start] != '\n' {
		return
	}
	recognized = true
	if !bytes.HasSuffix(raw, []byte("-->")) || bytes.Count(raw, []byte("<!--")) != 1 || bytes.Count(raw, []byte("-->")) != 1 {
		err = fmt.Errorf("annotation must be one complete HTML comment")
		return
	}
	if start+1 < len(raw) && raw[start] == '\r' && raw[start+1] == '\n' {
		start += 2
	} else if start < len(raw) && raw[start] == '\n' {
		start++
	} else {
		err = fmt.Errorf("annotation introducer requires a newline")
		return
	}
	return start, len(raw) - 3, encoding, true, nil
}

func addIsland(d *Document, start, end, payloadStart, payloadEnd int, encoding string, complete bool, options ParseOptions) *Island {
	is := &Island{Range: d.Span(start, end), PayloadRange: d.Span(payloadStart, payloadEnd), Encoding: encoding, Complete: complete}
	src := d.Source[payloadStart:payloadEnd]
	if encoding != "" {
		var err error
		src, err = decodeHostPayload(src)
		if err != nil {
			d.Add("HOST_ENCODING", "syntax", err.Error(), is.Range, "")
			return nil
		}
	}
	is.Document = ParseCore(d.File, src, ParseOptions{Recover: options.Recover, Fragment: !complete})
	d.Islands = append(d.Islands, is)
	mapRange := func(r Range) Range {
		if encoding != "" {
			return is.Range
		}
		return d.Span(payloadStart+r.Start.Offset, payloadStart+r.End.Offset)
	}
	for _, diag := range is.Document.Diagnostics {
		if encoding != "" {
			r := diag.Range
			diag.DecodedRange = &r
		}
		diag.Range = mapRange(diag.Range)
		for i := range diag.Related {
			diag.Related[i].Range = mapRange(diag.Related[i].Range)
		}
		d.Diagnostics = append(d.Diagnostics, diag)
	}
	// Keep island-local ranges available to the formatter and source-map clients.
	encoded, _ := json.Marshal(is.Document.Nodes)
	var nodes []*Node
	_ = json.Unmarshal(encoded, &nodes)
	for _, node := range nodes {
		mapNodeRanges(node, mapRange)
	}
	d.Nodes = append(d.Nodes, nodes...)
	return is
}

func mapNodeRanges(n *Node, f func(Range) Range) {
	n.Range = f(n.Range)
	n.KindRange = f(n.KindRange)
	if n.ID != "" {
		n.IDRange = f(n.IDRange)
	}
	mapValueRanges(n.Label, f)
	for _, p := range n.Props {
		mapPropertyRanges(p, f)
	}
	for _, c := range n.Children {
		mapNodeRanges(c, f)
	}
}
func mapPropertyRanges(p *Property, f func(Range) Range) {
	p.Range = f(p.Range)
	p.NameRange = f(p.NameRange)
	mapValueRanges(p.Value, f)
}
func mapValueRanges(v *Value, f func(Range) Range) {
	if v == nil {
		return
	}
	v.Range = f(v.Range)
	for _, item := range v.Items {
		mapValueRanges(item, f)
	}
	for _, p := range v.Fields {
		mapPropertyRanges(p, f)
	}
	for _, x := range []*Value{v.Left, v.Right, v.Object, v.Index, v.Tree} {
		mapValueRanges(x, f)
	}
	for _, x := range v.Args {
		mapValueRanges(x, f)
	}
}

func markdownRange(n ast.Node, source []byte) (int, int) {
	if n.Lines().Len() > 0 {
		start := n.Lines().At(0).Start
		end := n.Lines().At(n.Lines().Len() - 1).Stop
		if h, ok := n.(*ast.HTMLBlock); ok && h.HasClosure() {
			end = h.ClosureLine.Stop
		}
		return start, end
	}
	if c := n.FirstChild(); c != nil {
		a, _ := markdownRange(c, source)
		_, b := markdownRange(n.LastChild(), source)
		return a, b
	}
	return 0, 0
}

func parseMarkdown(d *Document, bom int, options ParseOptions) {
	source := d.Source[bom:]
	const marker = "<!-- uicl:markdown 1.0 -->"
	if !bytes.HasPrefix(source, []byte(marker)) || len(source) > len(marker) && source[len(marker)] != '\r' && source[len(marker)] != '\n' {
		d.Add("HOST_MODE_OR_VERSION_CONFLICT", "syntax", "Expected the Markdown 1.0 marker on the first line", d.Span(bom, bom+min(len(source), len(marker))), "")
		return
	}
	tree := goldmark.New().Parser().Parse(text.NewReader(source))
	d.HostAST = tree
	var blocks []ast.Node
	for n := tree.FirstChild(); n != nil; n = n.NextSibling() {
		blocks = append(blocks, n)
	}
	annotations := map[int]bool{}
	islands := map[int]*Island{}
	headers := 0
	for i, n := range blocks {
		h, ok := n.(*ast.HTMLBlock)
		if !ok {
			continue
		}
		a, b := markdownRange(h, source)
		raw := bytes.TrimRight(source[a:b], " \t\r\n")
		b = a + len(raw)
		a += len(raw) - len(bytes.TrimLeft(raw, " \t"))
		raw = source[a:b]
		// goldmark reports the HTML token start after up to three leading spaces.
		lineStart := bytes.LastIndexByte(source[:a], '\n') + 1
		trim := bytes.TrimSpace(raw)
		if bytes.Equal(trim, []byte(marker)) {
			headers++
			annotations[i] = true
			if a != lineStart {
				d.Add("HOST_ANNOTATION_COLUMN", "syntax", "Markdown annotations must begin at column zero", d.Span(a+bom, b+bom), "")
			}
			continue
		}
		if bytes.HasPrefix(trim, []byte("<!-- uicl:")) && !bytes.HasPrefix(trim, []byte("<!-- uicl:base64url")) {
			d.Add("HOST_MODE_OR_VERSION_CONFLICT", "syntax", "Conflicting host mode or version marker", d.Span(a+bom, b+bom), "")
			annotations[i] = true
			continue
		}
		p, q, enc, recognized, err := commentPayload(raw)
		if !recognized {
			continue
		}
		annotations[i] = true
		if a != lineStart {
			d.Add("HOST_ANNOTATION_COLUMN", "syntax", "Markdown annotations must begin at column zero", d.Span(a+bom, b+bom), "")
			continue
		}
		if err != nil {
			d.Add("HOST_COMMENT_BOUNDARY", "syntax", err.Error(), d.Span(a+bom, b+bom), "")
			continue
		}
		islands[i] = addIsland(d, a+bom, b+bom, a+p+bom, a+q+bom, enc, false, options)
	}
	if headers != 1 {
		d.Add("HOST_DUPLICATE_MARKER", "syntax", "Expected exactly one Markdown mode marker", d.Span(bom, bom+len(marker)), "")
	}
	for index, is := range islands {
		if is == nil {
			continue
		}
		for j := index + 1; j < len(blocks); j++ {
			if !annotations[j] {
				a, b := markdownRange(blocks[j], source)
				if blocks[j].Pos() >= 0 {
					a = blocks[j].Pos()
				}
				if j+1 < len(blocks) && blocks[j+1].Pos() >= a {
					b = blocks[j+1].Pos()
				} else {
					b = len(source)
				}
				b = a + len(bytes.TrimRight(source[a:b], " \t\r\n"))
				r := d.Span(a+bom, b+bom)
				is.TargetRange = &r
				break
			}
		}
		for _, n := range Walk(is.Document.Nodes) {
			if n.Kind != "doc.block" {
				continue
			}
			r := is.Range
			if n.Text("target") != "next" {
				d.Add("MARKDOWN_TARGET_REQUIRES_NEXT", "shape", "Markdown doc.block target must be next", r, n.ID)
			} else if is.TargetRange == nil {
				d.Add("HOST_TARGET_MISSING", "link", "No following sibling content block exists", r, n.ID)
			}
		}
	}
}

type htmlSourceToken struct {
	token              html.Token
	start, end         int
	blocked            bool
	bodyStart, bodyEnd int
	closed             bool
}

var blockedHTML = map[string]bool{"pre": true, "code": true, "textarea": true, "template": true, "script": true, "style": true}
var voidHTML = map[string]bool{"area": true, "base": true, "br": true, "col": true, "embed": true, "hr": true, "img": true, "input": true, "link": true, "meta": true, "param": true, "source": true, "track": true, "wbr": true}

func htmlBlocked(n *html.Node) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && blockedHTML[p.Data] {
			return true
		}
	}
	return false
}
func htmlKey(t html.Token) string {
	b, _ := json.Marshal(t.Attr)
	return fmt.Sprintf("%d:%s:%s", t.Type, t.Data, b)
}
func htmlNodeKey(n *html.Node) string {
	kind := html.StartTagToken
	if n.Type == html.CommentNode {
		kind = html.CommentToken
	}
	return htmlKey(html.Token{Type: kind, Data: n.Data, Attr: n.Attr})
}

func parseHTML(d *Document, options ParseOptions) {
	// x/net/html decides actual tree ancestry; lexical blocked regions provide a
	// conservative second gate and retain original source slices for editing.
	tree, err := html.Parse(bytes.NewReader(d.Source))
	if err != nil {
		d.Add("HOST_HTML", "syntax", err.Error(), d.Span(0, len(d.Source)), "")
		return
	}
	d.HostAST = tree
	byKey := map[string][]*htmlSourceToken{}
	z := html.NewTokenizer(bytes.NewReader(d.Source))
	offset := 0
	var stack []string
	var capture *htmlSourceToken
	for {
		kind := z.Next()
		raw := z.Raw()
		start := offset
		offset += len(raw)
		if kind == html.ErrorToken {
			if z.Err() != io.EOF {
				d.Add("HOST_HTML", "syntax", z.Err().Error(), d.Span(start, offset), "")
			}
			break
		}
		token := z.Token()
		blocked := false
		for _, tag := range stack {
			blocked = blocked || blockedHTML[tag]
		}
		item := &htmlSourceToken{token: token, start: start, end: offset, blocked: blocked}
		if kind == html.StartTagToken || kind == html.SelfClosingTagToken {
			keyToken := token
			keyToken.Type = html.StartTagToken
			byKey[htmlKey(keyToken)] = append(byKey[htmlKey(keyToken)], item)
			if token.Data == "script" && !blocked {
				capture = item
				capture.bodyStart = offset
			}
			if !voidHTML[token.Data] {
				stack = append(stack, token.Data)
			}
		} else if kind == html.EndTagToken {
			if token.Data == "script" && capture != nil {
				capture.bodyEnd = start
				capture.closed = true
				capture = nil
			}
			for j := len(stack) - 1; j >= 0; j-- {
				if stack[j] == token.Data {
					stack = stack[:j]
					break
				}
			}
		} else if kind == html.CommentToken {
			byKey[htmlKey(token)] = append(byKey[htmlKey(token)], item)
		}
	}
	headers := 0
	ids := map[string][]Range{}
	var sources []*htmlSourceToken
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.CommentNode || n.Type == html.ElementNode {
			items := byKey[htmlNodeKey(n)]
			if len(items) > 0 {
				item := items[0]
				byKey[htmlNodeKey(n)] = items[1:]
				blocked := item.blocked || htmlBlocked(n)
				if !blocked && n.Type == html.ElementNode {
					for _, a := range n.Attr {
						if a.Key == "id" {
							ids[a.Val] = append(ids[a.Val], d.Span(item.start, item.end))
						}
					}
					if n.Data == "script" {
						sources = append(sources, item)
					}
				}
				if !blocked && n.Type == html.CommentNode {
					clean := strings.TrimSpace(n.Data)
					if clean == "uicl:html 1.0" {
						headers++
						if n.Parent == nil || n.Parent.Data != "head" {
							d.Add("HOST_MARKER_OUTSIDE_HEAD", "syntax", "HTML mode marker must be a real head comment", d.Span(item.start, item.end), "")
						}
					} else if strings.HasPrefix(clean, "uicl:") && !strings.HasPrefix(clean, "uicl:base64url") {
						d.Add("HOST_MODE_OR_VERSION_CONFLICT", "syntax", "Conflicting host mode or version marker", d.Span(item.start, item.end), "")
					} else {
						sources = append(sources, item)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(tree)
	if headers != 1 {
		d.Add("HOST_MARKER_COUNT", "syntax", "Expected exactly one HTML 1.0 marker in head", d.Span(0, min(len(d.Source), 1)), "")
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].start < sources[j].start })
	for _, item := range sources {
		if item.token.Type == html.CommentToken {
			p, q, enc, recognized, err := commentPayload(d.Source[item.start:item.end])
			if !recognized {
				continue
			}
			if err != nil {
				d.Add("HOST_COMMENT_BOUNDARY", "syntax", err.Error(), d.Span(item.start, item.end), "")
				continue
			}
			addIsland(d, item.start, item.end, item.start+p, item.start+q, enc, false, options)
		} else {
			attrs := map[string]string{}
			duplicate := false
			for _, a := range item.token.Attr {
				if _, ok := attrs[a.Key]; ok {
					duplicate = true
				}
				attrs[a.Key] = a.Val
			}
			_, plain := attrs["data-uicl"]
			enc, encoded := attrs["data-uicl-encoding"]
			if !plain && !encoded {
				continue
			}
			if duplicate {
				d.Add("HOST_ATTRIBUTE_DUPLICATE", "syntax", "UICL data block has duplicate attributes", d.Span(item.start, item.end), "")
				continue
			}
			if !strings.EqualFold(attrs["type"], "text/plain") {
				d.Add("HOST_SCRIPT_TYPE", "syntax", "UICL data blocks require type=text/plain", d.Span(item.start, item.end), "")
				continue
			}
			if encoded && enc != "base64url" {
				d.Add("HOST_ENCODING", "syntax", "Unsupported data block encoding", d.Span(item.start, item.end), "")
				continue
			}
			if !item.closed {
				d.Add("HOST_UNCLOSED_DATA_BLOCK", "syntax", "UICL data block requires a closing script tag", d.Span(item.start, len(d.Source)), "")
				continue
			}
			addIsland(d, item.start, item.bodyEnd, item.bodyStart, item.bodyEnd, enc, true, options)
		}
	}
	bound := map[string]Range{}
	for _, is := range d.Islands {
		for _, n := range Walk(is.Document.Nodes) {
			if n.Kind != "doc.block" {
				continue
			}
			target := n.Get("target")
			r := is.Range
			if target == nil || target.Tag != "record" || len(target.Fields) != 1 || target.Fields[0].Name != "html_id" || target.Fields[0].Value.Type != "text" {
				d.Add("HTML_TARGET_REQUIRES_ID", "shape", "HTML doc.block requires target: {html_id: ...}", r, n.ID)
				continue
			}
			key := target.Fields[0].Value.Text
			if len(ids[key]) != 1 {
				d.Add("HOST_TARGET_MISSING_OR_DUPLICATE", "link", "HTML target must resolve to exactly one ID", r, n.ID)
				continue
			}
			if previous, exists := bound[key]; exists {
				d.Add("HOST_TARGET_CONFLICT", "link", "HTML target is already bound by another contract", r, n.ID)
				d.Diagnostics[len(d.Diagnostics)-1].Related = []Location{{File: d.File, Range: previous, Message: "First binding"}}
			} else {
				bound[key] = r
				targetRange := ids[key][0]
				is.TargetRange = &targetRange
			}
		}
	}
}

func fillHostTokens(d *Document) {
	cursor := 0
	for _, is := range d.Islands {
		start, end := is.PayloadRange.Start.Offset, is.PayloadRange.End.Offset
		if start > cursor {
			d.Tokens = append(d.Tokens, Token{Kind: "HOST", Text: string(d.Source[cursor:start]), Range: d.Span(cursor, start)})
		}
		if is.Encoding != "" {
			d.Tokens = append(d.Tokens, Token{Kind: "ENCODED", Text: string(d.Source[start:end]), Range: is.PayloadRange})
		} else {
			for _, token := range is.Document.Tokens {
				token.Range = d.Span(start+token.Range.Start.Offset, start+token.Range.End.Offset)
				d.Tokens = append(d.Tokens, token)
			}
		}
		cursor = end
	}
	if cursor < len(d.Source) {
		d.Tokens = append(d.Tokens, Token{Kind: "HOST", Text: string(d.Source[cursor:]), Range: d.Span(cursor, len(d.Source))})
	}
}
