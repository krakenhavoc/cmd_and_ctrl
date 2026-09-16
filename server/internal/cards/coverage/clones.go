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

// CloneGroup is one body that appears more than once.
type CloneGroup struct {
	// Hash identifies the body text; it is what the baseline records.
	Hash string
	// Lines is the body's length, for the report.
	Lines int
	// Members names each occurrence: "file.go:func Name" or
	// "file.go:closure@line".
	Members []string
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
