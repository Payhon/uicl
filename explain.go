package uicl

import "strings"

type Explanation struct {
	Code    string `json:"code"`
	Phase   string `json:"phase"`
	Summary string `json:"summary"`
	Fix     string `json:"fix"`
}

// Descriptions group related diagnostics; the precise source-specific reason is
// always carried by Diagnostic.Message. No model or network lookup is involved.
func Explain(code string) (Explanation, bool) {
	type group struct{ codes, phase, summary, fix string }
	groups := []group{
		{"DUPLICATE_PROPERTY", "shape", "A property is declared more than once.", "Keep exactly one declaration, including the primary text shorthand."},
		{"SCHEMA_PROPERTY_NAME SCHEMA_PORT_NAME DUPLICATE_SCHEMA_PORT SCHEMA_PORT_TYPE SCHEMA_DEFAULT_TYPE SCHEMA_ENUM_TYPE SCHEMA_DEFAULT_ENUM SCHEMA_CHILD", "profile", "A schema contains an invalid property, port, default, enum or child definition.", "Check the Profile authoring contract and register every referenced kind with its exact dependency version."},
		{"UICL-S003 UICL-S005 UICL-LIMIT UICL-VERSION UICL-LAYOUT UICL-DUPLICATE_KEY UICL-MISSING_VALUE UICL-RAW UICL-NUMBER_RANGE", "syntax", "The input violates UICL syntax, Unicode, layout or exact numeric limits.", "Use the marked source range to correct the input according to Core 1.0 and Layout-1. Raw contents are never interpreted as declarations."},
		{"SYNTAX VERSION LAYOUT MISSING_VALUE DUPLICATE_KEY RAW STRING NUMBER NUMBER_RANGE INT_RANGE DECIMAL_RANGE LIMIT INVALID_UNICODE", "syntax", "The document violates a Core syntax, encoding or resource constraint.", "Check the indicated token against UICL 1.0 and Layout-1; use two-space indentation and keep raw block boundaries intact."},
		{"UNKNOWN_NODE UNKNOWN_PROPERTY MISSING_PROPERTY TYPE_SHAPE ENUM EXPRESSION_FORBIDDEN PRIMARY_CONFLICT PRIMARY_FORBIDDEN MISSING_ID FORBIDDEN_ID DUPLICATE_ID CHILD_FORBIDDEN", "shape", "The declaration does not match its loaded Profile schema.", "Inspect the owning Profile; correct the indicated property or explicitly load the required extension. Do not silently discard fields."},
		{"MUTUALLY_EXCLUSIVE TYPE_RECORD_OR_ENUM SERVER_ACCESS_REQUIRED GRAMMAR_REPEAT_BOUNDS DIMENSION_PAIR VIDEO_TIMEBASE TIMELINE_BOUNDS TIMELINE_OVERLAP DUPLICATE_COLUMN PG_DEFAULT_CONFLICT PG_UNKNOWN_KEY_COLUMN PG_UNKNOWN_COLUMN PG_TRANSACTION_CONFLICT SECRET_EMPTY_SCOPE", "static", "An implemented domain constraint is violated.", "Correct the fields named in the diagnostic. Preserve access rules and use an explicit migration or execution strategy where required."},
		{"PROFILE_ROOT PROFILE_ID PROFILE_CORE PROFILE_VERSION PROFILE_CONFLICT SCHEMA_CONFLICT SCHEMA_NAME DUPLICATE_SCHEMA_PROPERTY DUPLICATE_PORT PROFILE_DEPENDENCY PROFILE_CYCLE UNKNOWN_SHAPE_TYPE UNKNOWN_RECORD_TYPE PROFILE_META PROFILE_SCHEMA PROFILE_NAMESPACE SCHEMA_PRIMARY SCHEMA_IDENTITY SCHEMA_CHILDREN SCHEMA_PROPERTY SCHEMA_PORT SCHEMA_DEFAULT SCHEMA_ENUM SCHEMA_TARGET PROPERTY_TYPE", "profile", "A Profile definition or its dependency graph is inconsistent.", "Use exact dependency versions, unique names and valid Meta schemas. Query profile inspect to review the loaded definitions."},
		{"DUPLICATE_MODULE UNKNOWN_EXPORT IMPORT_REQUIRES_URI IMPORT_NONLOCAL_NOT_SUPPORTED IMPORT_CYCLE UNKNOWN_IMPORT NOT_EXPORTED UNRESOLVED_REF UNKNOWN_PORT REF_TARGET_KIND PATH_ESCAPE", "link", "A local module or reference cannot be resolved within the workspace contract.", "Check --root, relative use.source, exports, identity and port names. Network modules are not fetched automatically."},
		{"HOST_LIMIT HOST_MODE_OR_VERSION_CONFLICT HOST_ANNOTATION_COLUMN HOST_COMMENT_BOUNDARY HOST_DUPLICATE_MARKER HOST_ENCODING HOST_HTML HOST_MARKER_OUTSIDE_HEAD HOST_MARKER_COUNT HOST_ATTRIBUTE_DUPLICATE HOST_SCRIPT_TYPE HOST_UNCLOSED_DATA_BLOCK MARKDOWN_TARGET_REQUIRES_NEXT HTML_TARGET_REQUIRES_ID HOST_TARGET_MISSING HOST_TARGET_MISSING_OR_DUPLICATE HOST_TARGET_CONFLICT", "host", "A host marker, contract carrier or content binding is invalid.", "Use the UICL 1.0 host marker and real annotation boundaries; check target uniqueness and canonical unpadded base64url encoding."},
		{"LOCK_FORMAT LOCK_SOURCE LOCK_DIGEST LOCK_PROFILE LOCK_DEPENDENCY", "lock", "The package source lock no longer describes the local Profile sources.", "Compare the locked source files, IDs, versions, dependencies and SHA-256 digests. This command never updates the lock."},
		{"CONFLICT EDIT_RANGE EDIT_OVERLAP EDIT_UTF8 WRITE_TARGET FORMAT_SEMANTICS", "edit", "The requested edit cannot safely preserve the source contract.", "Reload the current source, verify edit ranges and expected digest, then regenerate the candidate edit. No automatic conflict overwrite is performed."},
		{"USAGE INPUT_REQUIRED CATALOG_REQUIRED IO", "tool", "The command could not be invoked or completed with the supplied inputs.", "Check command help, explicit file paths, root directory and file readability."},
	}
	for _, g := range groups {
		for _, c := range strings.Fields(g.codes) {
			if c == code {
				return Explanation{code, g.phase, g.summary, g.fix}, true
			}
		}
	}
	return Explanation{}, false
}
