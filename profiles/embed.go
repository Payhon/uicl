// Package profiles embeds the canonical standard Profile sources.
package profiles

import "embed"

// Files is the same source catalog used by the specification and Python tools.
//
//go:embed *.uicl
var Files embed.FS
