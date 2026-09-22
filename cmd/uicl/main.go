package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Payhon/uicl"
)

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

type stringsFlag []string

func (s *stringsFlag) String() string     { return strings.Join(*s, ",") }
func (s *stringsFlag) Set(v string) error { *s = append(*s, v); return nil }

type options struct {
	root, profileDir, filename string
	extra                      stringsFlag
	json                       bool
}

func flags(name string, args []string, errOut io.Writer, configure func(*flag.FlagSet)) (*flag.FlagSet, *options, error) {
	o := &options{}
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(errOut)
	f.StringVar(&o.root, "root", ".", "workspace root")
	f.BoolVar(&o.json, "json", false, "emit one JSON result")
	f.StringVar(&o.profileDir, "profile-dir", "", "replace the embedded standard Profile directory")
	f.Var(&o.extra, "extra-profile", "load an explicit local extension (repeatable)")
	f.StringVar(&o.filename, "filename", "", "file identity for standard input")
	if configure != nil {
		configure(f)
	}
	return f, o, f.Parse(interspersed(f, args))
}

// flag.FlagSet handles flag syntax; this only allows flags after file arguments.
func interspersed(f *flag.FlagSet, args []string) []string {
	var switches, files []string
	for i := 0; i < len(args); i++ {
		s := args[i]
		if s == "--" {
			files = append(files, args[i+1:]...)
			break
		}
		if s == "-" || !strings.HasPrefix(s, "-") {
			files = append(files, s)
			continue
		}
		switches = append(switches, s)
		name := strings.TrimLeft(s, "-")
		if strings.Contains(name, "=") {
			continue
		}
		if v := f.Lookup(name); v != nil {
			if b, ok := v.Value.(interface{ IsBoolFlag() bool }); ok && b.IsBoolFlag() {
				continue
			}
			if i+1 < len(args) {
				i++
				switches = append(switches, args[i])
			}
		}
	}
	return append(append(switches, "--"), files...)
}
func wantsJSON(args []string) bool {
	for _, a := range args {
		if a == "--json" || a == "-json" || a == "--json=true" {
			return true
		}
	}
	return false
}
func jsonOut(out io.Writer, v any) int {
	e := json.NewEncoder(out)
	e.SetIndent("", "  ")
	e.SetEscapeHTML(false)
	if e.Encode(v) != nil {
		return 2
	}
	return 0
}
func diagnostics(out io.Writer, ds []uicl.Diagnostic) {
	for _, d := range ds {
		fmt.Fprintf(out, "%s:%d:%d: %s: %s\n", d.File, d.Range.Start.Line+1, d.Range.Start.Column+1, d.Code, d.Message)
	}
}
func failure(out, errOut io.Writer, asJSON bool, code, phase, message, file string) int {
	r := uicl.NewReport()
	r.Diagnostics = append(r.Diagnostics, uicl.Diagnostic{Code: code, Phase: phase, Severity: "error", Message: message, File: file})
	r.Finish()
	if asJSON {
		jsonOut(out, r)
	} else {
		diagnostics(errOut, r.Diagnostics)
	}
	return r.ExitCode()
}
func report(out, errOut io.Writer, asJSON bool, r uicl.Report) int {
	if asJSON {
		if jsonOut(out, r) != 0 {
			return 2
		}
	} else {
		diagnostics(errOut, r.Diagnostics)
		fmt.Fprintf(out, "%s: %d documents, %d Profiles, %d references (syntax/shape/link/selected checks only)\n", r.Status, r.Documents, r.Profiles, r.References)
	}
	return r.ExitCode()
}
func catalog(o *options) (*uicl.Catalog, []uicl.Diagnostic, error) {
	return uicl.LoadCatalog(o.root, o.profileDir, []string(o.extra))
}
func catalogFailure(out, errOut io.Writer, o *options, ds []uicl.Diagnostic, e error) int {
	if e != nil {
		if strings.Contains(e.Error(), "PATH_ESCAPE") {
			return failure(out, errOut, o.json, "PATH_ESCAPE", "profile", e.Error(), "")
		}
		return failure(out, errOut, o.json, "IO", "io", e.Error(), "")
	}
	r := uicl.NewReport()
	r.Diagnostics = ds
	r.Finish()
	return report(out, errOut, o.json, r)
}

func run(args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) == 0 {
		return failure(out, errOut, false, "USAGE", "tool", "Usage: uicl <check|fmt|profile inspect|explain|lock|version> [options]", "")
	}
	command, rest := args[0], args[1:]
	if command == "help" || command == "--help" || command == "-h" {
		fmt.Fprintln(out, "Usage: uicl <check|fmt|profile inspect|explain|lock|version> [options]\nUse uicl COMMAND --help for flags. check and fmt require explicit files. No code is executed.")
		return 0
	}
	if command == "profile" {
		if len(rest) == 0 || rest[0] != "inspect" {
			return failure(out, errOut, wantsJSON(rest), "USAGE", "tool", "Usage: uicl profile inspect ID [--json]", "")
		}
		rest = rest[1:]
	}
	var check, write bool
	f, o, e := flags(command, rest, errOut, func(f *flag.FlagSet) {
		if command == "fmt" || command == "lock" {
			f.BoolVar(&check, "check", false, "check without writing")
		}
		if command == "fmt" {
			f.BoolVar(&write, "write", false, "write changed files")
		}
	})
	if errors.Is(e, flag.ErrHelp) {
		return 0
	}
	if e != nil {
		return failure(out, errOut, wantsJSON(rest), "USAGE", "tool", e.Error(), "")
	}
	files := f.Args()
	usage := func(message string) int { return failure(out, errOut, o.json, "USAGE", "tool", message, "") }
	switch command {
	case "check":
		if len(files) == 0 {
			return usage("check requires at least one explicit file")
		}
		cat, ds, e := catalog(o)
		if e != nil || uicl.HasErrors(ds) {
			return catalogFailure(out, errOut, o, ds, e)
		}
		w, e := uicl.NewWorkspace(o.root, cat)
		if e != nil {
			return failure(out, errOut, o.json, "IO", "io", e.Error(), o.root)
		}
		for _, p := range files {
			if p != "-" {
				continue
			}
			if len(files) != 1 || o.filename == "" {
				return usage("standard input requires a single '-' and --filename")
			}
			b, e := io.ReadAll(io.LimitReader(in, uicl.MaxBytes+1))
			if e != nil {
				return failure(out, errOut, o.json, "IO", "io", e.Error(), o.filename)
			}
			if len(b) > uicl.MaxBytes {
				return failure(out, errOut, o.json, "LIMIT", "syntax", "Input exceeds byte limit", o.filename)
			}
			if e = w.SetDocument(o.filename, b); e != nil {
				return failure(out, errOut, o.json, "PATH_ESCAPE", "link", e.Error(), o.filename)
			}
			files = []string{o.filename}
		}
		return report(out, errOut, o.json, w.Check(files...))
	case "fmt":
		return formatCommand(o, files, check, write, in, out, errOut)
	case "profile":
		if len(files) != 1 {
			return usage("profile inspect requires one Profile ID")
		}
		cat, ds, e := catalog(o)
		if e != nil || uicl.HasErrors(ds) {
			return catalogFailure(out, errOut, o, ds, e)
		}
		p := cat.Profiles[files[0]]
		if p == nil {
			return usage("Unknown Profile: " + files[0])
		}
		if o.json {
			return jsonOut(out, map[string]any{"schemaVersion": 1, "status": "PASS", "profile": p})
		}
		fmt.Fprintf(out, "%s@%s\nSource: %s\nSHA-256: %s\n\n%s", p.ID, p.Version, p.Path, p.Digest, p.Source)
		return 0
	case "explain":
		if len(files) != 1 {
			return usage("explain requires one diagnostic code")
		}
		x, ok := uicl.Explain(files[0])
		if !ok {
			return usage("Unknown diagnostic code: " + files[0])
		}
		if o.json {
			return jsonOut(out, map[string]any{"schemaVersion": 1, "status": "PASS", "explanation": x})
		}
		fmt.Fprintf(out, "%s (%s)\n%s\nFix: %s\n", x.Code, x.Phase, x.Summary, x.Fix)
		return 0
	case "lock":
		if !check || len(files) > 1 {
			return usage("Usage: uicl lock --check [FILE]")
		}
		p := filepath.Join(o.root, "uicl.lock.json")
		if len(files) == 1 {
			p = files[0]
		}
		return report(out, errOut, o.json, uicl.CheckLock(o.root, p))
	case "version":
		if len(files) > 0 {
			return usage("version takes no file arguments")
		}
		cat, ds, e := catalog(o)
		if e != nil || uicl.HasErrors(ds) {
			return catalogFailure(out, errOut, o, ds, e)
		}
		versions := map[string]string{}
		for id, p := range cat.Profiles {
			versions[id] = p.Version
		}
		if o.json {
			return jsonOut(out, map[string]any{"schemaVersion": 1, "status": "PASS", "version": uicl.Version, "core": uicl.CoreVersion, "profiles": versions})
		}
		fmt.Fprintf(out, "uicl %s (Core %s)\n", uicl.Version, uicl.CoreVersion)
		ids := make([]string, 0, len(versions))
		for id := range versions {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			fmt.Fprintf(out, "%s@%s\n", id, versions[id])
		}
		return 0
	default:
		return usage("Unknown command: " + command)
	}
}

type formattedFile struct {
	File    string  `json:"file"`
	Changed bool    `json:"changed"`
	Written bool    `json:"written"`
	Content *string `json:"content,omitempty"`
}
type formatResult struct {
	SchemaVersion int               `json:"schemaVersion"`
	Status        string            `json:"status"`
	Diagnostics   []uicl.Diagnostic `json:"diagnostics"`
	Files         []formattedFile   `json:"files"`
}
type pendingFormat struct {
	path, digest string
	source       []byte
}

func formatCommand(o *options, files []string, check, write bool, in io.Reader, out, errOut io.Writer) int {
	bad := func(s string) int { return failure(out, errOut, o.json, "USAGE", "tool", s, "") }
	if len(files) == 0 {
		return bad("fmt requires explicit files")
	}
	if check && write {
		return bad("--check and --write are mutually exclusive")
	}
	if !check && !write && len(files) != 1 {
		return bad("stdout formatting accepts only one file")
	}
	r := formatResult{SchemaVersion: 1, Status: "PASS", Diagnostics: []uicl.Diagnostic{}, Files: []formattedFile{}}
	var pending []pendingFormat
	exit := 0
	seen := map[string]bool{}
	add := func(code, phase, message, file string) {
		r.Diagnostics = append(r.Diagnostics, uicl.Diagnostic{Code: code, Phase: phase, Severity: "error", Message: message, File: file})
		if phase == "io" {
			exit = 2
		} else if exit == 0 {
			exit = 1
		}
	}
	for _, name := range files {
		var b []byte
		var e error
		p := name
		if name == "-" {
			if len(files) != 1 || o.filename == "" || write {
				return bad("stdin requires a single '-', --filename, and no --write")
			}
			p, e = uicl.ResolvePath(o.root, o.filename)
			if e == nil {
				b, e = io.ReadAll(io.LimitReader(in, uicl.MaxBytes+1))
			}
		} else {
			p, e = uicl.ResolvePath(o.root, name)
			if e == nil {
				b, e = uicl.ReadSource(p)
			}
		}
		if e != nil {
			phase, code := "io", "IO"
			if strings.Contains(e.Error(), "PATH_ESCAPE") {
				phase, code = "link", "PATH_ESCAPE"
			}
			if strings.HasPrefix(e.Error(), "LIMIT:") {
				phase, code = "syntax", "LIMIT"
			}
			add(code, phase, e.Error(), name)
			continue
		}
		if seen[p] {
			continue
		}
		seen[p] = true
		d := uicl.Parse(p, b, uicl.ParseOptions{Recover: true})
		formatted, ds := uicl.Format(d)
		r.Diagnostics = append(r.Diagnostics, ds...)
		if uicl.HasErrors(ds) {
			if exit == 0 {
				exit = 1
			}
			for _, d := range ds {
				if d.Phase == "format" && d.Severity == "error" {
					exit = 2
				}
			}
			continue
		}
		changed := !bytes.Equal(b, formatted)
		entry := formattedFile{File: p, Changed: changed}
		if !check && !write {
			s := string(formatted)
			entry.Content = &s
		}
		r.Files = append(r.Files, entry)
		pending = append(pending, pendingFormat{p, uicl.Digest(b), formatted})
		if changed && check && exit == 0 {
			exit = 1
		}
	}
	// Preflight every file before any mutation. Runtime write failures can still
	// leave earlier files written; their Written fields make that explicit.
	if write && exit == 0 {
		for i, p := range pending {
			if !r.Files[i].Changed {
				continue
			}
			if e := uicl.WriteSource(p.path, p.digest, p.source); e != nil {
				code, phase := "IO", "io"
				if strings.HasPrefix(e.Error(), "CONFLICT:") {
					code, phase = "CONFLICT", "edit"
				}
				add(code, phase, e.Error(), p.path)
				exit = 2 // A refused or failed write is a tool execution error.
				break
			}
			r.Files[i].Written = true
		}
	}
	if exit == 1 {
		r.Status = "FAIL"
	}
	if exit == 2 {
		r.Status = "ERROR"
	}
	uicl.SortDiagnostics(r.Diagnostics)
	if o.json {
		if jsonOut(out, r) != 0 {
			return 2
		}
	} else {
		diagnostics(errOut, r.Diagnostics)
		for _, f := range r.Files {
			if f.Content != nil {
				if exit == 0 {
					if _, e := io.WriteString(out, *f.Content); e != nil {
						return 2
					}
				}
			} else if f.Changed {
				action := "needs formatting"
				if f.Written {
					action = "formatted"
				}
				fmt.Fprintf(out, "%s: %s\n", f.File, action)
			}
		}
	}
	return exit
}
