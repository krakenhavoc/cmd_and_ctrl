package coverage

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// clones.go — the exact-duplicate detector behind
// TestNoNewExactClonesInTheCatalog (#584, Discussion #558).
//
// The September 2026 review found that the catalog's largest problem
// was not the engine but copy-paste in the card layer: hundreds of
// function and closure bodies that were the same text under different
// names, grown one batch at a time. The constructors (#590), the tally
// (#591), OncePerBatch (#594) and the token table (#611) removed the
// biggest families; this keeps the count from growing back. It looks
// only for EXACT duplicates — whitespace-normalised bodies that are
// byte-identical — because those are never deliberate. Structural
// near-clones are a judgement call and are left to review.
//
// #895: the baseline is keyed on the body HASH and the set of
// declaring FILES, never on a line number, so an edit ABOVE a listed
// closure no longer rewrites the file. See CloneGroup.Files.

// CloneGroup is one body that appears more than once.
type CloneGroup struct {
	// Hash identifies the body text; it is what the baseline is keyed
	// on and the only field the gate reads back.
	Hash string
	// Lines is the body's LENGTH in lines, for the report. Not a line
	// NUMBER: it does not move when something above the body does.
	Lines int
	// Members names each occurrence: "file.go:func Name" or
	// "file.go:closure@line". The line number is for a human reading
	// a failure — it is deliberately NOT part of the baseline (#895).
	Members []string
}

// Files is the set of files this body is declared in, sorted and
// deduplicated — the baseline's location column since #895.
//
// The line number is dropped on purpose. A baseline row keyed on
// "batch17_helpers.go:closure@214" is rewritten by `-update` whenever
// anything ABOVE that closure grows or shrinks a line, so an unrelated
// edit turns into a diff in a file nobody touched; three agents hit it
// in one week (#636, #870/#877 in PR #878). The gate only ever needs
// the body hash, and a reader only ever needs to know which files to
// open — neither is a line number.
func (g CloneGroup) Files() []string {
	seen := make(map[string]bool, len(g.Members))
	out := make([]string, 0, len(g.Members))
	for _, m := range g.Members {
		file := m
		if i := strings.Index(m, ":"); i >= 0 {
			file = m[:i]
		}
		if seen[file] {
			continue
		}
		seen[file] = true
		out = append(out, file)
	}
	sort.Strings(out)
	return out
}

// CloneBaselineHeader is the two comment lines the baseline opens
// with. Exported so the writer and the format test agree on them
// without either restating them.
const CloneBaselineHeader = "# exact-duplicate bodies in server/internal/cards/effects, one per line:\n" +
	"# <hash> <lines> <members> <files>. Keyed on the hash and the declaring files, never on\n" +
	"# a line number (#895). Regenerate with -update; never add a line by hand.\n"

// FormatCloneBaseline renders the baseline file for a set of groups.
//
// ONE writer, used both by the `-update` path and by the test that
// proves the format is line-insensitive — two renderings could not
// disagree about what the baseline says if only one of them exists.
//
// Deterministic: ExactClones returns the groups hash-sorted, Files is
// sorted, and nothing here reads the filesystem or a line number.
func FormatCloneBaseline(groups []CloneGroup) string {
	var sb strings.Builder
	sb.WriteString(CloneBaselineHeader)
	for _, g := range groups {
		fmt.Fprintf(&sb, "%s %d %d %s\n", g.Hash, g.Lines, len(g.Members), strings.Join(g.Files(), ","))
	}
	return sb.String()
}

// ExactClones parses every non-test Go file in dir and returns the
// function and closure bodies of at least minLines lines that occur
// two or more times, sorted by hash.
func ExactClones(dir string, minLines int) ([]CloneGroup, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	type member struct {
		name  string
		lines int
	}
	byHash := map[string][]member{}
	for _, f := range files {
		base := filepath.Base(f)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		af, err := parser.ParseFile(fset, f, src, 0)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", base, err)
		}
		add := func(name string, body *ast.BlockStmt) {
			if body == nil {
				return
			}
			lines := fset.Position(body.Rbrace).Line - fset.Position(body.Lbrace).Line + 1
			if lines < minLines {
				return
			}
			var buf bytes.Buffer
			if err := printer.Fprint(&buf, fset, body); err != nil {
				return
			}
			norm := strings.Join(strings.Fields(buf.String()), " ")
			sum := sha1.Sum([]byte(norm))
			h := hex.EncodeToString(sum[:8])
			byHash[h] = append(byHash[h], member{name, lines})
		}
		for _, d := range af.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			add(base+":func "+fd.Name.Name, fd.Body)
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if fl, ok := n.(*ast.FuncLit); ok {
					add(fmt.Sprintf("%s:closure@%d", base, fset.Position(fl.Pos()).Line), fl.Body)
				}
				return true
			})
		}
	}
	var out []CloneGroup
	for h, ms := range byHash {
		if len(ms) < 2 {
			continue
		}
		g := CloneGroup{Hash: h, Lines: ms[0].lines}
		for _, m := range ms {
			g.Members = append(g.Members, m.name)
		}
		sort.Strings(g.Members)
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Hash < out[j].Hash })
	return out, nil
}
