#!/usr/bin/env bash
# prototype.sh — THROWAWAY prototype for #5906 (OPTIONS.md, option A).
#
# Run from the root of a scratch worktree of develop. Not a migration script:
# it compiles, but the tests it moves are not all green (see OPTIONS.md).
#
# Needs:
#   PROTO    optional: a built proto2 binary. If unset, the script builds the
#            go/types rewriter embedded at the end of this file (proto2/main.go)
#            with go1.27 into a temp dir.
#   EXPORTS  export-data map of effects' dependencies, made with:
#              cd server && go list -export -deps -test \
#                -f '{{.ImportPath}} {{.Export}}' ./internal/cards/effects/ \
#                | grep -v cards/effects > exports.txt
#            (build the dependencies first; listing the test variant of effects
#             itself would compile it, which is what OOMs.)
#
# What it does, in one type-checked pass over package effects (sources + tests):
#   - card files (an init(), no exported declaration) go to effects/cards/<x>,
#     at most 700 per shard, named by first letter (s1..s4 when a letter is big)
#   - framework files, and any card file that another package would need, stay
#     in effects (core)
#   - each top-level declaration of a _test.go file is placed on its own:
#     helpers another file uses go to effects/effectstest (exported, non-test),
#     the rest go to effects/tests/<x> (at most 150 test files per shard), which
#     blank-import the aggregator effects/cards
#   - every unexported name that now crosses a package boundary is exported
#     (first letter upper-cased, X-prefixed on a collision) at every use site
#   - moved files dot-import the packages they use
set -euo pipefail
EXPORTS=${EXPORTS:?export map}
if [ -z "${PROTO:-}" ]; then
  tmp=$(mktemp -d); mkdir -p "$tmp/proto2"
  printf 'module proto2\n\ngo 1.27.0\n' > "$tmp/proto2/go.mod"
  sed -n '/^: <<.PROTO_GO_EOF.$/,/^PROTO_GO_EOF$/p' "$0" | sed '1d;$d' > "$tmp/proto2/main.go"
  (cd "$tmp/proto2" && go build -o proto2 .); PROTO=$tmp/proto2/proto2
fi
cd server/internal/cards/effects
"$PROTO" . "$EXPORTS" apply "${MAX_CARDS:-700}" "${MAX_TESTS:-150}"
# Hand fix-up the tool does not do: an anonymous struct type with unexported
# fields crosses a package boundary in one test helper (HornTopFive).
grep -rl 'HornTopFive(' effectstest tests | xargs sed -i 's/struct{ *name, typeLine string *}/struct{ Name, TypeLine string }/g'
sed -i 's/names\[i\]\.name, TypeLine: names\[i\]\.typeLine/names[i].Name, TypeLine: names[i].TypeLine/' effectstest/horn_of_the_mark_h.go
# Every other file that imports effects also blank-imports the aggregator, so a
# binary or test that reads the catalog still sees every card (a real migration
# would do this only where the catalog is read, plus a guard; see OPTIONS.md).
cd ../../..
grep -rl --include=*.go '"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"$' cmd internal \
  | grep -v '^internal/cards/effects/' \
  | xargs sed -i -E \
      -e 's#^import ((\S+ )?"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects")$#import (\n\t\1\n\t_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards"\n)#' \
      -e 's#^(\s+)(\S+ )?"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"$#&\n\1_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards"#'

exit 0

# The rewriter (never executed by bash; extracted and built above).
: <<'PROTO_GO_EOF'
// proto2: throwaway prototype for issue #5906, option A (refined).
//
//	effects            core: framework, shared helpers, every card file another file needs
//	effects/cards/<x>  card shards (no tests), ~<=MAX files each, named by first letters
//	effects/cards      aggregator: blank-imports every shard
//	effects/effectstest  test helpers (declarations lifted out of _test.go files, exported)
//	effects/tests/<x>  test-only packages, by test file name; each blank-imports the aggregator
//
// Usage: proto2 <effects dir> <exports map> <plan|apply> <maxCardFilesPerShard> <maxTestFilesPerShard>
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const modPath = "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
const core, tst = "core", "effectstest"

type unit struct {
	file string
	idx  int // -1: whole file
	decl ast.Decl
}

func main() {
	dir, exp, mode := os.Args[1], os.Args[2], os.Args[3]
	maxPer, _ := strconv.Atoi(os.Args[4])
	exports := map[string]string{}
	f, _ := os.Open(exp)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		p := strings.Fields(sc.Text())
		if len(p) == 2 {
			exports[p[0]] = p[1]
		}
	}
	fset := token.NewFileSet()
	matches, _ := filepath.Glob(filepath.Join(dir, "*.go"))
	sort.Strings(matches)
	var files []*ast.File
	src := map[string][]byte{}
	byName := map[string]*ast.File{}
	var names []string
	for _, m := range matches {
		b, _ := os.ReadFile(m)
		af, err := parser.ParseFile(fset, m, b, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			panic(err)
		}
		n := filepath.Base(m)
		files = append(files, af)
		src[n] = b
		byName[n] = af
		names = append(names, n)
	}
	imp := importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) { return os.Open(exports[path]) })
	info := &types.Info{Uses: map[*ast.Ident]types.Object{}, Defs: map[*ast.Ident]types.Object{}, Types: map[ast.Expr]types.TypeAndValue{}}
	nerr := 0
	conf := types.Config{Importer: imp, Error: func(err error) { nerr++ }}
	pkg, _ := conf.Check("effects", fset, files, info)
	fmt.Fprintf(os.Stderr, "type errors: %d\n", nerr)
	isTest := func(n string) bool { return strings.HasSuffix(n, "_test.go") }

	// units
	var units []*unit
	unitAt := map[string][]*unit{} // per file, ordered
	for _, n := range names {
		af := byName[n]
		if !isTest(n) {
			u := &unit{file: n, idx: -1}
			units = append(units, u)
			unitAt[n] = []*unit{u}
			continue
		}
		for i, d := range af.Decls {
			if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
				continue
			}
			u := &unit{file: n, idx: i, decl: d}
			units = append(units, u)
			unitAt[n] = append(unitAt[n], u)
		}
	}
	unitOf := func(p token.Pos) *unit {
		n := filepath.Base(fset.Position(p).Filename)
		us := unitAt[n]
		if len(us) == 1 && us[0].idx == -1 {
			return us[0]
		}
		for _, u := range us {
			if p >= u.decl.Pos() && p <= u.decl.End() {
				return u
			}
		}
		return nil
	}
	hasInit := map[string]bool{}
	declExported := map[string]bool{}
	for n, af := range byName {
		if isTest(n) {
			continue
		}
		for _, d := range af.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil && d.Name.Name == "init" {
					hasInit[n] = true
				}
				if d.Recv == nil && ast.IsExported(d.Name.Name) {
					declExported[n] = true
				}
			case *ast.GenDecl:
				for _, s := range d.Specs {
					switch s := s.(type) {
					case *ast.ValueSpec:
						for _, id := range s.Names {
							if ast.IsExported(id.Name) {
								declExported[n] = true
							}
						}
					case *ast.TypeSpec:
						if ast.IsExported(s.Name.Name) {
							declExported[n] = true
						}
					}
				}
			}
		}
	}
	inScope := func(o types.Object) bool {
		if o == nil || o.Pkg() != pkg || !o.Pos().IsValid() {
			return false
		}
		if o.Parent() == pkg.Scope() {
			return true
		}
		switch x := o.(type) {
		case *types.Var:
			return x.IsField()
		case *types.Func:
			return x.Type().(*types.Signature).Recv() != nil
		}
		return false
	}
	type edge struct {
		use, decl *unit
		obj       types.Object
	}
	var edges []edge
	for id, o := range info.Uses {
		if !inScope(o) {
			continue
		}
		u, d := unitOf(id.Pos()), unitOf(o.Pos())
		if u == nil || d == nil {
			continue
		}
		if u != d {
			edges = append(edges, edge{u, d, o})
		}
	}
	coloc := map[*unit][]*unit{}
	for id, o := range info.Defs {
		fn, ok := o.(*types.Func)
		if !ok || fn.Pkg() != pkg {
			continue
		}
		sig := fn.Type().(*types.Signature)
		if sig.Recv() == nil {
			continue
		}
		t := sig.Recv().Type()
		if p, ok := t.(*types.Pointer); ok {
			t = p.Elem()
		}
		if nt, ok := t.(*types.Named); ok && nt.Obj().Pkg() == pkg {
			a, b := unitOf(id.Pos()), unitOf(nt.Obj().Pos())
			if a != b {
				coloc[a] = append(coloc[a], b)
				coloc[b] = append(coloc[b], a)
			}
		}
	}
	// shard naming: by first letter; letters with > maxPer files split by second letter into ranges
	shardNames := func(list []string, prefix string) map[string]string {
		byL := map[string][]string{}
		for _, n := range list {
			c := rune(n[0])
			l := "0"
			if unicode.IsLetter(c) {
				l = string(unicode.ToLower(c))
			}
			byL[l] = append(byL[l], n)
		}
		out := map[string]string{}
		for l, ns := range byL {
			sort.Strings(ns)
			if maxPer <= 0 || len(ns) <= maxPer {
				for _, n := range ns {
					out[n] = prefix + l
				}
				continue
			}
			parts := (len(ns) + maxPer - 1) / maxPer
			per := (len(ns) + parts - 1) / parts
			k := 1
			for i, n := range ns {
				// do not split one second-letter group across shards when possible
				if i > 0 && i%per == 0 {
					k++
				}
				out[n] = fmt.Sprintf("%s%s%d", prefix, l, k)
			}
		}
		return out
	}
	assign := map[*unit]string{}
	var cardFiles, testFiles []string
	for _, n := range names {
		if isTest(n) {
			testFiles = append(testFiles, n)
		} else if hasInit[n] && !declExported[n] {
			cardFiles = append(cardFiles, n)
		}
	}
	isTestPkg := func(p string) bool { return strings.HasPrefix(p, "tests") }
	isShard := func(p string) bool { return strings.HasPrefix(p, "cards") }
	for _, u := range units {
		if !isTest(u.file) {
			if hasInit[u.file] && !declExported[u.file] {
				assign[u] = "?"
			} else {
				assign[u] = core
			}
		}
	}
	cs := shardNames(cardFiles, "cards")
	maxPer, _ = strconv.Atoi(os.Args[5])
	ts := shardNames(testFiles, "tests")
	for _, u := range units {
		if assign[u] == "?" {
			assign[u] = cs[u.file]
			if os.Args[4] == "0" { // option B: cards stay in effects, only tests move
				assign[u] = core
			}
		} else if isTest(u.file) {
			assign[u] = ts[u.file]
		}
	}
	isTestFunc := func(u *unit) bool {
		fd, ok := u.decl.(*ast.FuncDecl)
		if !ok || fd.Recv != nil {
			return false
		}
		return strings.HasPrefix(fd.Name.Name, "Test") || strings.HasPrefix(fd.Name.Name, "Benchmark") || strings.HasPrefix(fd.Name.Name, "Fuzz") || strings.HasPrefix(fd.Name.Name, "Example") || fd.Name.Name == "init" || fd.Name.Name == "TestMain"
	}
	// dependency rule: may pkg A use pkg B?
	allowed := func(a, b string) bool {
		if a == b || b == core {
			return true
		}
		switch {
		case isShard(b):
			return a == tst || isTestPkg(a)
		case b == tst:
			return isTestPkg(a)
		}
		return false
	}
	bad := 0
	for changed := true; changed; {
		changed = false
		set := func(u *unit, p string) {
			if assign[u] != p {
				assign[u] = p
				changed = true
			}
		}
		for _, e := range edges {
			pu, pd := assign[e.use], assign[e.decl]
			if allowed(pu, pd) {
				continue
			}
			if isTest(e.decl.file) {
				if isTestFunc(e.decl) {
					continue // reported below
				}
				// test decl needed elsewhere: it becomes a helper; a helper needed by core is impossible
				set(e.decl, tst)
			} else {
				set(e.decl, core)
			}
		}
		for a, bs := range coloc {
			for _, b := range bs {
				if assign[a] != assign[b] {
					if isTest(a.file) {
						set(a, tst)
						set(b, tst)
					} else {
						set(a, core)
						set(b, core)
					}
				}
			}
		}
	}
	for _, e := range edges {
		if !allowed(assign[e.use], assign[e.decl]) {
			bad++
			fmt.Fprintf(os.Stderr, "UNRESOLVED %s -> %s (%s)\n", e.use.file, e.decl.file, e.obj.Name())
		}
	}
	// renames
	need := map[types.Object]bool{}
	for _, e := range edges {
		if assign[e.use] != assign[e.decl] && !e.obj.Exported() {
			need[e.obj] = true
		}
	}
	for n, af := range byName {
		ast.Inspect(af, func(nd ast.Node) bool {
			cl, ok := nd.(*ast.CompositeLit)
			if !ok || len(cl.Elts) == 0 {
				return true
			}
			if _, keyed := cl.Elts[0].(*ast.KeyValueExpr); keyed {
				return true
			}
			tv, ok := info.Types[cl]
			if !ok {
				return true
			}
			nt, ok := tv.Type.(*types.Named)
			if !ok || nt.Obj().Pkg() != pkg {
				return true
			}
			st, ok := nt.Underlying().(*types.Struct)
			u := unitOf(cl.Pos())
			_ = n
			if !ok || u == nil || assign[unitOf(nt.Obj().Pos())] == assign[u] {
				return true
			}
			for i := 0; i < st.NumFields(); i++ {
				if !st.Field(i).Exported() {
					need[st.Field(i)] = true
				}
			}
			return true
		})
	}
	allIdents := map[*ast.Ident]types.Object{}
	for id, o := range info.Defs {
		if o != nil {
			allIdents[id] = o
		}
	}
	for id, o := range info.Uses {
		allIdents[id] = o
	}
	methodNames := map[string]bool{}
	for o := range need {
		if fn, ok := o.(*types.Func); ok && fn.Type().(*types.Signature).Recv() != nil {
			methodNames[o.Name()] = true
		}
	}
	for _, o := range allIdents {
		if fn, ok := o.(*types.Func); ok && fn.Pkg() == pkg && fn.Type().(*types.Signature).Recv() != nil && methodNames[fn.Name()] {
			need[o] = true
		}
	}
	for _, o := range allIdents {
		if v, ok := o.(*types.Var); ok && v.Embedded() && v.Pkg() == pkg {
			t := v.Type()
			if p, ok := t.(*types.Pointer); ok {
				t = p.Elem()
			}
			if nt, ok := t.(*types.Named); ok && need[nt.Obj()] {
				need[o] = true
			}
		}
	}
	taken := map[string]bool{}
	for _, n := range pkg.Scope().Names() {
		taken[n] = true
	}
	for _, o := range allIdents {
		if o == nil || o.Pkg() != pkg {
			continue
		}
		if v, ok := o.(*types.Var); ok && v.IsField() {
			taken["m."+o.Name()] = true
		}
		if fn, ok := o.(*types.Func); ok && fn.Type().(*types.Signature).Recv() != nil {
			taken["m."+o.Name()] = true
		}
	}
	exportOf := func(n string) string {
		r := []rune(n)
		if r[0] == '_' {
			return "X" + n
		}
		r[0] = unicode.ToUpper(r[0])
		return string(r)
	}
	var objs []types.Object
	for o := range need {
		objs = append(objs, o)
	}
	sort.Slice(objs, func(i, j int) bool { return objs[i].Pos() < objs[j].Pos() })
	newName := map[types.Object]string{}
	byMethod := map[string]string{}
	collisions := 0
	for _, o := range objs {
		if fn, ok := o.(*types.Func); ok && fn.Type().(*types.Signature).Recv() != nil {
			if nn, ok := byMethod[o.Name()]; ok {
				newName[o] = nn
				continue
			}
			nn := exportOf(o.Name())
			for taken["m."+nn] {
				nn = "X" + nn
				collisions++
			}
			byMethod[o.Name()] = nn
			taken["m."+nn] = true
			newName[o] = nn
			continue
		}
		nn := exportOf(o.Name())
		if v, ok := o.(*types.Var); ok && v.IsField() && !v.Embedded() {
			for taken["m."+nn] {
				nn = "X" + nn
				collisions++
			}
			taken["m."+nn] = true
		}
		if o.Parent() == pkg.Scope() {
			for taken[nn] {
				nn = "X" + nn
				collisions++
			}
			taken[nn] = true
		}
		newName[o] = nn
	}
	type ed struct {
		off, l int
		s      string
	}
	edits := map[string][]ed{}
	for id, o := range allIdents {
		if nn, ok := newName[o]; ok {
			n := filepath.Base(fset.Position(id.Pos()).Filename)
			edits[n] = append(edits[n], ed{fset.Position(id.Pos()).Offset, len(id.Name), nn})
		}
	}
	// per unit: imported packages used (PkgName), and our packages used unqualified
	pkgUses := map[*unit]map[string]string{} // path -> local name ("" = default)
	ourUses := map[*unit]map[string]bool{}
	for id, o := range info.Uses {
		u := unitOf(id.Pos())
		if u == nil {
			continue
		}
		if pn, ok := o.(*types.PkgName); ok {
			if pkgUses[u] == nil {
				pkgUses[u] = map[string]string{}
			}
			name := ""
			if pn.Name() != pn.Imported().Name() {
				name = pn.Name()
			}
			pkgUses[u][pn.Imported().Path()] = name
			continue
		}
		if o == nil || o.Pkg() != pkg || o.Parent() != pkg.Scope() {
			continue
		}
		d := unitOf(o.Pos())
		if d != nil && assign[d] != assign[u] {
			if ourUses[u] == nil {
				ourUses[u] = map[string]bool{}
			}
			ourUses[u][assign[d]] = true
		}
	}
	// blank / dot imports in original files that must be kept (side effects)
	counts := map[string]int{}
	fileCount := map[string]map[string]bool{}
	for u, p := range assign {
		counts[p]++
		if fileCount[p] == nil {
			fileCount[p] = map[string]bool{}
		}
		fileCount[p][u.file] = true
	}
	var ks []string
	for k := range counts {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	fmt.Fprintf(os.Stderr, "renamed objects: %d (collisions: %d), files with rename edits: %d, unresolved edges: %d\n", len(newName), collisions, len(edits), bad)
	var srcBytes = map[string]int{}
	for u, p := range assign {
		if u.idx == -1 {
			srcBytes[p] += len(src[u.file])
		} else {
			srcBytes[p] += int(u.decl.End() - u.decl.Pos())
		}
	}
	for _, k := range ks {
		fmt.Fprintf(os.Stderr, "  %-12s units=%5d files=%5d bytes=%.1fMB\n", k, counts[k], len(fileCount[k]), float64(srcBytes[k])/1e6)
	}
	if mode != "apply" {
		return
	}
	for n, es := range edits {
		sort.Slice(es, func(i, j int) bool { return es[i].off > es[j].off })
		b := append([]byte(nil), src[n]...)
		last := -1
		for _, e := range es {
			if e.off == last {
				continue
			}
			last = e.off
			b = append(b[:e.off:e.off], append([]byte(e.s), b[e.off+e.l:]...)...)
		}
		edits[n] = es
		src[n] = b
	}
	// offsets shift after renames: map an original offset to the new one
	shift := func(n string, off int) int {
		d := 0
		for _, e := range edits[n] { // descending offsets
			if e.off < off {
				d += len(e.s) - e.l
			}
		}
		return off + d
	}
	pkgDir := func(p string) (dirRel, pkgName string) {
		switch {
		case p == core:
			return ".", "effects"
		case p == tst:
			return "effectstest", "effectstest"
		case isShard(p):
			return "cards/" + strings.TrimPrefix(p, "cards"), p
		default:
			return "tests/" + strings.TrimPrefix(p, "tests"), p
		}
	}
	importPath := func(p string) string {
		d, _ := pkgDir(p)
		if d == "." {
			return modPath
		}
		return modPath + "/" + d
	}
	shardSet := map[string]bool{}
	for _, p := range assign {
		if isShard(p) {
			shardSet[p] = true
		}
	}
	// write non-test files
	for _, n := range names {
		if isTest(n) {
			continue
		}
		u := unitAt[n][0]
		p := assign[u]
		b := src[n]
		if p != core {
			_, pk := pkgDir(p)
			var imps []string
			for q := range ourUses[u] {
				imps = append(imps, `. "`+importPath(q)+`"`)
			}
			sort.Strings(imps)
			b = replacePkg(b, pk, imps)
			d, _ := pkgDir(p)
			os.MkdirAll(filepath.Join(dir, d), 0o755)
			os.WriteFile(filepath.Join(dir, d, n), b, 0o644)
			os.Remove(filepath.Join(dir, n))
		} else {
			os.WriteFile(filepath.Join(dir, n), b, 0o644)
		}
	}
	// test files: split by unit package
	for _, n := range names {
		if !isTest(n) {
			continue
		}
		af := byName[n]
		groups := map[string][]*unit{}
		for _, u := range unitAt[n] {
			groups[assign[u]] = append(groups[assign[u]], u)
		}
		os.Remove(filepath.Join(dir, n))
		for p, us := range groups {
			d, pk := pkgDir(p)
			used := map[string]string{}
			ours := map[string]bool{}
			for _, u := range us {
				for path, nm := range pkgUses[u] {
					used[path] = nm
				}
				for q := range ourUses[u] {
					ours[q] = true
				}
			}
			// keep blank/dot imports of the original file
			for _, is := range af.Imports {
				if is.Name != nil && (is.Name.Name == "_" || is.Name.Name == ".") {
					path, _ := strconv.Unquote(is.Path.Value)
					used[path] = is.Name.Name
				}
			}
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "package %s\n\nimport (\n", pk)
			var paths []string
			for path := range used {
				paths = append(paths, path)
			}
			sort.Strings(paths)
			for _, path := range paths {
				if used[path] != "" {
					fmt.Fprintf(&buf, "\t%s %q\n", used[path], path)
				} else {
					fmt.Fprintf(&buf, "\t%q\n", path)
				}
			}
			var oq []string
			for q := range ours {
				oq = append(oq, q)
			}
			sort.Strings(oq)
			for _, q := range oq {
				fmt.Fprintf(&buf, "\t. %q\n", importPath(q))
			}
			if isTestPkg(p) {
				fmt.Fprintf(&buf, "\t_ %q\n", modPath+"/cards")
			}
			buf.WriteString(")\n")
			for _, u := range us {
				start := u.decl.Pos()
				var doc *ast.CommentGroup
				switch x := u.decl.(type) {
				case *ast.FuncDecl:
					doc = x.Doc
				case *ast.GenDecl:
					doc = x.Doc
				}
				if doc != nil {
					start = doc.Pos()
				}
				s := shift(n, fset.Position(start).Offset)
				e := shift(n, fset.Position(u.decl.End()).Offset)
				buf.WriteString("\n")
				buf.Write(src[n][s:e])
				buf.WriteString("\n")
			}
			out := n
			if p == tst {
				out = strings.TrimSuffix(n, "_test.go") + "_h.go"
			}
			os.MkdirAll(filepath.Join(dir, d), 0o755)
			os.WriteFile(filepath.Join(dir, d, out), buf.Bytes(), 0o644)
		}
	}
	// aggregator
	var sl []string
	for s := range shardSet {
		sl = append(sl, s)
	}
	sort.Strings(sl)
	var buf bytes.Buffer
	buf.WriteString("// Package cards links every catalog shard into the binary (prototype, #5906).\npackage cards\n\nimport (\n")
	for _, s := range sl {
		fmt.Fprintf(&buf, "\t_ %q\n", importPath(s))
	}
	buf.WriteString(")\n")
	os.MkdirAll(filepath.Join(dir, "cards"), 0o755)
	os.WriteFile(filepath.Join(dir, "cards", "cards.go"), buf.Bytes(), 0o644)
}

func replacePkg(b []byte, pk string, imps []string) []byte {
	s := string(b)
	i := strings.Index(s, "\npackage effects\n")
	j := i + len("\npackage effects\n")
	if strings.HasPrefix(s, "package effects\n") {
		i, j = -1, len("package effects\n")
	}
	hdr := "package " + pk + "\n"
	for _, im := range imps {
		hdr += "\nimport " + im + "\n"
	}
	return []byte(s[:i+1] + hdr + s[j:])
}
PROTO_GO_EOF
