package game

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// omitzero_tag_guard_test.go is #1492's lint.
//
// `encoding/json`'s `omitzero` struct-tag option is honoured only from
// Go 1.24 onward; on an older toolchain it is silently ignored and the
// field is always written. When #1492 was found, go.mod's minimum and
// CI's `setup-go` were both 1.22, so a contributor building with a
// newer local Go would silently omit a zero-valued `omitzero` field
// that every CI-built binary, and every fixture CI's own
// `TestWriteSnapshotCorpus` run produced, always writes. That
// divergence is exactly what made
// TestSnapshotCorpusRestores fail locally and pass in CI: three
// GameSnapshot-reachable fields carried the tag, and three more
// structs reachable from it (CastBanRule.Filter, CastTimingRule.Filter,
// PlayerStatic.Timing, PlayerStatic.CastBan, CastPermission.Filter,
// CastPermission.Duration) carried it too, unnoticed because no
// corpus fixture happened to exercise a non-zero value on those six.
//
// The fix was to drop the tag everywhere in this module (always write
// the field, matching what every CI-built binary already did). This
// test is what stops it coming back: it parses every non-generated,
// non-test Go source under the server module and fails on any struct
// field tag containing the `omitzero` option, so the failure surfaces
// at the point of introduction rather than as a "works in CI, fails
// on my machine" report weeks later.
//
// Since ADR 0122 §1.1 the module declares Go 1.27, so every toolchain
// that can build it honours the option and that divergence cannot
// recur. The guard stays, by that ADR's call: every field it once
// caught is now always written, the frozen snapshot fixtures carry
// those keys, and adding `omitzero` to one drops a key at its zero
// value — a change to what is written, not an addition to it. Narrow
// it, or allowlist a site with its reason, rather than silence it.

// omitzeroTagAllowlist exempts one "<file>:<line>" struct field, keyed
// as the walk prints it, with the reason it is not the bug. An
// unexplained entry is how a lint stops meaning anything.
var omitzeroTagAllowlist = map[string]string{}

func TestNoOmitzeroStructTagBelowGo124(t *testing.T) {
	findings, scanned := scanTreeForOmitzeroTags(t, "../..")
	if scanned == 0 {
		t.Fatal("scanned no files — the walk is wrong, and a lint that reads nothing passes everything")
	}
	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Errorf("a struct field tag uses the `omitzero` JSON option, which stops writing the key at its "+
		"zero value — a change to what snapshots and the wire carry, not an addition (#1492, ADR 0122 "+
		"§1.1). Drop the option (or, if it is genuinely fine here, add the site to omitzeroTagAllowlist "+
		"with the reason).\n  %s",
		strings.Join(findings, "\n  "))
}

// scanTreeForOmitzeroTags walks root for Go sources, test files
// included (a test fixture struct is still marshaled by the same
// rules), and reports every struct field tag containing `omitzero`.
func scanTreeForOmitzeroTags(t *testing.T, root string) ([]string, int) {
	t.Helper()
	fset := token.NewFileSet()
	var findings []string
	scanned := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "testdata", "node_modules", "bin":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			t.Fatalf("parse %s: %v", path, perr)
		}
		scanned++
		findings = append(findings, omitzeroTagsInFile(fset, path, file)...)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return findings, scanned
}

// omitzeroTagsInFile reports every struct field in f whose tag's
// `json` option list contains `omitzero`, skipping anything in
// omitzeroTagAllowlist.
func omitzeroTagsInFile(fset *token.FileSet, file string, f *ast.File) []string {
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok || st.Fields == nil {
			return true
		}
		for _, field := range st.Fields.List {
			if field.Tag == nil {
				continue
			}
			raw, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				continue
			}
			jsonTag := reflect.StructTag(raw).Get("json")
			if jsonTag == "" {
				continue
			}
			parts := strings.Split(jsonTag, ",")
			hasOmitzero := false
			for _, opt := range parts[1:] {
				if opt == "omitzero" {
					hasOmitzero = true
					break
				}
			}
			if !hasOmitzero {
				continue
			}
			pos := fset.Position(field.Tag.Pos())
			key := fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
			if _, allowed := omitzeroTagAllowlist[key]; allowed {
				continue
			}
			out = append(out, key)
		}
		return true
	})
	return out
}
