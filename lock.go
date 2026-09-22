package uicl

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type lockEntry struct {
	ID       string   `json:"id"`
	Version  string   `json:"version"`
	Source   string   `json:"source"`
	SHA256   string   `json:"sha256"`
	Requires []string `json:"requires"`
}
type sourceLock struct {
	Language string      `json:"language"`
	Core     string      `json:"core"`
	Status   string      `json:"status"`
	Catalog  []lockEntry `json:"catalog"`
}

// CheckLock verifies the existing package source lock without rewriting it or
// interpreting its historical verification environment as a Go requirement.
func CheckLock(root, file string) Report {
	r := NewReport()
	r.Stages = map[string]string{"lock": "PASS"}
	r.NotEvaluated = []string{"production plan, runtime adapters and credentials"}
	add := func(code, phase, message, file string) {
		r.Diagnostics = append(r.Diagnostics, Diagnostic{Code: code, Phase: phase, Severity: "error", Message: message, File: file})
	}
	p, e := ResolvePath(root, file)
	if e != nil {
		code, phase := "IO", "io"
		if strings.Contains(e.Error(), "PATH_ESCAPE") {
			code, phase = "PATH_ESCAPE", "lock"
		}
		add(code, phase, e.Error(), file)
		r.Finish()
		return r
	}
	b, e := ReadSource(p)
	if e != nil {
		add("IO", "io", e.Error(), p)
		r.Finish()
		return r
	}
	var lock sourceLock
	if e = json.Unmarshal(b, &lock); e != nil {
		add("LOCK_FORMAT", "lock", e.Error(), p)
		r.Finish()
		return r
	}
	if lock.Language != "uicl" || lock.Core != CoreVersion || lock.Status != "package-source-lock-not-production-plan" {
		add("LOCK_FORMAT", "lock", "Expected a UICL 1.0 package source lock", p)
	}
	if len(lock.Catalog) == 0 {
		add("LOCK_FORMAT", "lock", "Lock catalog must not be empty", p)
	}
	entries := map[string]lockEntry{}
	sources := map[string]bool{}
	for _, entry := range lock.Catalog {
		if entry.ID == "" || entry.Version == "" {
			add("LOCK_FORMAT", "lock", "Profile ID and version are required", p)
		}
		if _, ok := entries[entry.ID]; ok {
			add("PROFILE_CONFLICT", "lock", "Duplicate locked Profile: "+entry.ID, p)
		}
		entries[entry.ID] = entry
		if entry.Source == "" || filepath.IsAbs(entry.Source) {
			add("LOCK_SOURCE", "lock", "Lock source must be a relative path", p)
			continue
		}
		source, e := ResolvePath(root, filepath.Join(filepath.Dir(p), filepath.FromSlash(entry.Source)))
		if e != nil {
			add("PATH_ESCAPE", "lock", e.Error(), p)
			continue
		}
		if sources[source] {
			add("LOCK_SOURCE", "lock", "Repeated Profile source: "+entry.Source, p)
		}
		sources[source] = true
		data, e := ReadSource(source)
		if e != nil {
			add("LOCK_SOURCE", "lock", "Locked source is unavailable: "+e.Error(), source)
			continue
		}
		if Digest(data) != entry.SHA256 {
			add("LOCK_DIGEST", "lock", "Source digest differs for "+entry.ID, source)
		}
		d := ParseCore(source, data, ParseOptions{Recover: true})
		r.Documents++
		r.Diagnostics = append(r.Diagnostics, d.Diagnostics...)
		if HasErrors(d.Diagnostics) {
			continue
		}
		if len(d.Nodes) != 1 || d.Nodes[0].Kind != "profile" {
			add("LOCK_SOURCE", "lock", "Each locked source must define exactly one Profile", source)
			continue
		}
		n := d.Nodes[0]
		if n.Text("id") != entry.ID || n.Text("version") != entry.Version || n.Text("core") != lock.Core {
			add("LOCK_PROFILE", "lock", "Profile identity, version or Core differs from lock", source)
		}
		actual := Strings(n.Get("requires"))
		expected := append([]string(nil), entry.Requires...)
		sort.Strings(actual)
		sort.Strings(expected)
		if strings.Join(actual, "\x00") != strings.Join(expected, "\x00") {
			add("LOCK_DEPENDENCY", "lock", "Declared dependencies differ from lock", source)
		}
	}
	state := map[string]int{}
	var visit func(string)
	visit = func(id string) {
		if state[id] == 2 {
			return
		}
		if state[id] == 1 {
			add("PROFILE_CYCLE", "lock", "Circular locked Profile dependency: "+id, p)
			return
		}
		state[id] = 1
		for _, dep := range entries[id].Requires {
			name, version, ok := strings.Cut(dep, "@")
			other, exists := entries[name]
			if !ok || !exists || other.Version != version {
				add("PROFILE_DEPENDENCY", "lock", fmt.Sprintf("Unresolved locked dependency %q of %s", dep, id), p)
				continue
			}
			visit(name)
		}
		state[id] = 2
	}
	ids := make([]string, 0, len(entries))
	for id := range entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		visit(id)
	}
	r.Profiles = len(entries)
	r.Finish()
	return r
}
