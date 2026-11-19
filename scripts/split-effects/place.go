package main

// place.go — --place: put files written for the one-package layout where the
// split layout wants them, and rewrite the names they use.
//
// A card or test file on a branch cut before the split refers to helpers by
// their old names (onBattlefield, newCatalogGame) and to the core unqualified
// (Register(Spec{...})). For each file --place:
//
//  1. picks its package: a _test.go file goes to its test package and a file
//     with an init() to its card shard (both from shards.txt); any other file
//     is a helper and goes to the core;
//  2. moves it there (git mv when git tracks it) and sets its package clause;
//  3. type-checks it with its new package, and rewrites every name that does
//     not resolve: through table.txt (the split's own rename table), else to
//     the core's or effectstest's exported name; an unexported core or
//     effectstest name it needs is exported in place (every use renamed);
//  4. writes keys into struct literals of another package's types, fixes the
//     imports and gofmts it.
//
// It never edits a string literal except a relative path in a test file that
// moved two directories down.

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type nameEntry struct{ dest, newName string }

type renameTable struct {
	name   map[string]nameEntry
	field  map[string]string // <pkgkey>.<Type>.<old> → new
	method map[string]string
}

func readTable(path string) (*renameTable, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &renameTable{name: map[string]nameEntry{}, field: map[string]string{}, method: map[string]string{}}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		l := sc.Text()
		if strings.HasPrefix(l, "#") || strings.TrimSpace(l) == "" {
			continue
		}
		fs := strings.Fields(l)
		switch {
		case fs[0] == "name" && len(fs) == 4:
			t.name[fs[1]] = nameEntry{fs[2], fs[3]}
		case fs[0] == "field" && len(fs) == 3:
			t.field[fs[1]] = fs[2]
		case fs[0] == "method" && len(fs) == 3:
			t.method[fs[1]] = fs[2]
		default:
			return nil, fmt.Errorf("%s: bad line %q", path, l)
		}
	}
	return t, sc.Err()
}

type placer struct {
	root, serverDir, effDir string
	tbl                     *rangeTable
	tab                     *renameTable
	ex                      *exportData
	coreDecls               map[string]map[string]bool // pkgkey → top-level names declared (core, effectstest)
	exportReq               map[string]map[string]bool // pkgkey → unexported names to export in place
}

func placeFiles(root string, args []string) error {
	serverDir := filepath.Join(root, "server")
	effDir := filepath.Join(serverDir, effectsRel)
	if !isSplit(effDir) {
		return fmt.Errorf("the effects tree is not split yet (no %s); --place only works after the split", coreCardsFile)
	}
	tbl, err := readRangeTable(filepath.Join(effDir, shardsFile))
	if err != nil {
		return err
	}
	tab, err := readTable(filepath.Join(toolDir(root), tableFile))
	if err != nil {
		return err
	}
	pl := &placer{root: root, serverDir: serverDir, effDir: effDir, tbl: tbl, tab: tab, ex: newExportData(serverDir),
		exportReq: map[string]map[string]bool{}}
	pl.coreDecls = map[string]map[string]bool{
		keyCore: topLevelNames(effDir, false),
		keyTst:  topLevelNames(filepath.Join(effDir, "effectstest"), false),
	}
	// 1-2: pick and move.
	byDest := map[string][]string{}
	moved := map[string]bool{}
	for _, a := range args {
		p, err := filepath.Abs(a)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(p, ".go") {
			return fmt.Errorf("%s: not a Go file", a)
		}
		if _, err := os.Stat(p); err != nil {
			return err
		}
		if !strings.HasPrefix(p, effDir+string(filepath.Separator)) {
			return fmt.Errorf("%s is not under %s", a, rel(root, effDir))
		}
		name := filepath.Base(p)
		var dest string
		switch {
		case isTestName(name):
			dest = tbl.testPkg(name)
		default:
			init, err := hasInit(p)
			if err != nil {
				return err
			}
			if init {
				dest = tbl.cardShard(name)
			} else {
				dest = keyCore
			}
		}
		to := filepath.Join(effDir, pkgDir(dest), name)
		if to != p {
			if _, err := os.Stat(to); err == nil {
				return fmt.Errorf("%s: %s already exists", a, rel(root, to))
			}
			if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
				return err
			}
			if _, err := run(root, "git", "ls-files", "--error-unmatch", p); err == nil {
				if _, err := run(root, "git", "mv", p, to); err != nil {
					return err
				}
			} else if err := os.Rename(p, to); err != nil {
				return err
			}
			// A test file that came from the core directory is now two levels down.
			moved[to] = filepath.Dir(p) == effDir && isTestKey(dest)
			logf("moved %s → %s", rel(root, p), rel(root, to))
		}
		byDest[dest] = append(byDest[dest], to)
	}
	// Core files that a rebase left with old names (a new helper file, an
	// edit to an existing one) are repaired too, whether named or not.
	if broken, err := pl.brokenCoreFiles(); err != nil {
		return err
	} else if len(broken) > 0 {
		have := map[string]bool{}
		for _, p := range byDest[keyCore] {
			have[p] = true
		}
		for _, p := range broken {
			if !have[p] {
				logf("repairing old names in %s", rel(root, p))
				byDest[keyCore] = append(byDest[keyCore], p)
			}
		}
	}
	var dests []string
	for d := range byDest {
		dests = append(dests, d)
	}
	sort.Strings(dests)
	// 3-4: rewrite. Core first, so helpers placed there resolve for the rest.
	sort.SliceStable(dests, func(i, j int) bool { return dests[i] == keyCore && dests[j] != keyCore })
	var problems []string
	for round := 0; round < 2; round++ {
		problems = nil
		for _, d := range dests {
			ps, err := pl.rewrite(d, byDest[d], moved)
			if err != nil {
				return err
			}
			problems = append(problems, ps...)
		}
		exported := false
		for _, k := range []string{keyCore, keyTst} {
			if len(pl.exportReq[k]) > 0 {
				if err := pl.exportInPlace(k); err != nil {
					return err
				}
				exported = true
			}
		}
		if !exported {
			break
		}
		// Once more, now that the names it exported exist.
		pl.exportReq = map[string]map[string]bool{}
		pl.coreDecls = map[string]map[string]bool{
			keyCore: topLevelNames(effDir, false),
			keyTst:  topLevelNames(filepath.Join(effDir, "effectstest"), false),
		}
		moved = map[string]bool{}
	}
	if len(problems) > 0 {
		return fmt.Errorf("placed the files, but these names need a hand fix:\n  %s", strings.Join(problems, "\n  "))
	}
	if _, err := run(root, "git", "add", "-A", "server/"+effectsRel); err != nil {
		logf("git add: %v", err)
	}
	logf("done: build and test the packages you placed into, then commit")
	return nil
}

// topLevelNames lists the package-level names declared in the non-test Go
// files of dir (syntactically).
func topLevelNames(dir string, tests bool) map[string]bool {
	out := map[string]bool{}
	fset := token.NewFileSet()
	for _, n := range goFiles(dir) {
		if isTestName(n) != tests {
			continue
		}
		f, err := parseFileSkip(fset, filepath.Join(dir, n))
		if err != nil {
			continue
		}
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					out[d.Name.Name] = true
				}
			case *ast.GenDecl:
				for _, sp := range d.Specs {
					switch sp := sp.(type) {
					case *ast.ValueSpec:
						for _, id := range sp.Names {
							out[id.Name] = true
						}
					case *ast.TypeSpec:
						out[sp.Name.Name] = true
					}
				}
			}
		}
	}
	return out
}

// checkPkg type-checks the package in dir (tests included for a test package),
// tolerating errors.
func (pl *placer) checkPkg(dest string) (*token.FileSet, []*srcFile, *types.Info, *types.Package, error) {
	dir := filepath.Join(pl.effDir, pkgDir(dest))
	var names []string
	for _, n := range goFiles(dir) {
		if isTestKey(dest) == isTestName(n) || (isTestKey(dest) && n == generatedMainTest) {
			names = append(names, n)
		}
	}
	fset := token.NewFileSet()
	files, err := parseDir(fset, dir, names)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	var asts []*ast.File
	for _, f := range files {
		asts = append(asts, f.ast)
	}
	info := &types.Info{
		Uses: map[*ast.Ident]types.Object{}, Defs: map[*ast.Ident]types.Object{},
		Types: map[ast.Expr]types.TypeAndValue{}, Implicits: map[ast.Node]types.Object{},
	}
	conf := types.Config{Importer: stubImporter{pl.ex.importer(fset)}, Error: func(error) {}}
	pkg, _ := conf.Check(importPath(dest), fset, asts, info)
	return fset, files, info, pkg, nil
}

func (pl *placer) rewrite(dest string, paths []string, moved map[string]bool) ([]string, error) {
	targets := map[string]bool{}
	for _, p := range paths {
		targets[p] = true
		// The package clause first.
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		fset := token.NewFileSet()
		f, err := parseSrc(fset, b)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", p, err)
		}
		tf := fset.File(f.Pos())
		es := []edit{{tf.Offset(f.Name.Pos()), len(f.Name.Name), pkgName(dest)}}
		if moved[p] {
			es = append(es, pathEdits(f, tf)...)
		}
		sortEdits(es)
		if err := os.WriteFile(p, applyEdits(b, 0, len(b), es), 0o644); err != nil {
			return nil, err
		}
	}
	// The dependencies other than the package itself.
	var problems []string
	for iter := 0; iter < 6; iter++ {
		fset, files, info, _, err := pl.checkPkg(dest)
		if err != nil {
			return nil, err
		}
		progress := false
		problems = nil
		for _, f := range files {
			if !targets[f.path] {
				continue
			}
			es, adds, ps := pl.fixFile(dest, f, info)
			problems = append(problems, ps...)
			if len(es) == 0 && len(adds) == 0 && iter > 0 {
				continue
			}
			sortEdits(es)
			out := applyEdits(f.src, 0, len(f.src), es)
			names := map[string]string{}
			for _, im := range f.ast.Imports {
				if pn, ok := info.Implicits[im].(*types.PkgName); ok {
					names[pn.Imported().Path()] = pn.Imported().Name()
				}
			}
			for _, a := range adds {
				names[a.path] = pkgName(keyOfImportPath(a.path))
			}
			b, err := fixImports(out, adds, names)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", f.path, err)
			}
			if !bytes.Equal(b, f.src) {
				progress = true
				if err := os.WriteFile(f.path, b, 0o644); err != nil {
					return nil, err
				}
			}
		}
		_ = fset
		if !progress {
			break
		}
	}
	return problems, nil
}

// pathEdits: relative paths in a test file that moved two levels down.
func pathEdits(f *ast.File, tf *token.File) []edit {
	var es []edit
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BasicLit:
			if n.Kind != token.STRING {
				return true
			}
			v, err := strconv.Unquote(n.Value)
			if err == nil && (v == "testdata" || strings.HasPrefix(v, "testdata/") || strings.HasPrefix(v, "../")) {
				es = append(es, edit{tf.Offset(n.Pos()), len(n.Value), strconv.Quote("../../" + v)})
			}
		case *ast.CallExpr:
			if se, ok := n.Fun.(*ast.SelectorExpr); ok && se.Sel.Name == "Join" {
				if x, ok := se.X.(*ast.Ident); ok && x.Name == "filepath" && len(n.Args) > 0 {
					if bl, ok := n.Args[0].(*ast.BasicLit); ok && bl.Value == `".."` {
						es = append(es, edit{tf.Offset(bl.Pos()), 0, `"..", "..", `})
					}
				}
			}
		}
		return true
	})
	return es
}

// fixFile rewrites the unresolved names of one file.
func (pl *placer) fixFile(dest string, f *srcFile, info *types.Info) ([]edit, []importSpec, []string) {
	var es []edit
	var problems []string
	need := map[string]bool{}
	alias := func(k string) string {
		need[k] = true
		return pkgName(k)
	}
	where := func(n ast.Node) string {
		p := f.tf.Position(n.Pos())
		return fmt.Sprintf("%s:%d", rel(pl.root, f.path), p.Line)
	}
	// Parent links.
	parent := map[ast.Node]ast.Node{}
	var stack []ast.Node
	ast.Inspect(f.ast, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		if len(stack) > 0 {
			parent[n] = stack[len(stack)-1]
		}
		stack = append(stack, n)
		return true
	})
	declared := map[string]bool{}
	for o := range info.Defs {
		declared[o.Name] = true
	}
	fieldFix := func(t types.Type, name string) (string, bool) {
		if p, ok := t.(*types.Pointer); ok {
			t = p.Elem()
		}
		nt, ok := t.(*types.Named)
		if !ok || nt.Obj().Pkg() == nil {
			return "", false
		}
		k := keyOfImportPath(nt.Obj().Pkg().Path())
		if nn, ok := pl.tab.field[k+"."+nt.Obj().Name()+"."+name]; ok {
			return nn, true
		}
		if nn, ok := pl.tab.method[name]; ok {
			if x, _, _ := types.LookupFieldOrMethod(t, true, nt.Obj().Pkg(), nn); x != nil {
				return nn, true
			}
		}
		nn := upperFirst(name)
		if x, _, _ := types.LookupFieldOrMethod(t, true, nt.Obj().Pkg(), nn); x != nil {
			return nn, true
		}
		return "", false
	}
	ast.Inspect(f.ast, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id.Name == "_" || id == f.ast.Name {
			return true
		}
		if info.Defs[id] != nil || info.Uses[id] != nil {
			return true
		}
		if _, isDef := info.Defs[id]; isDef {
			return true // a definition with no object (a package name, a label)
		}
		off := f.tf.Offset(id.Pos())
		switch p := parent[id].(type) {
		case *ast.SelectorExpr:
			if p.Sel != id {
				break
			}
			if x, ok := p.X.(*ast.Ident); ok {
				if pn, ok := info.Uses[x].(*types.PkgName); ok {
					// pkg.oldName: a qualified reference by an old name.
					k := keyOfImportPath(pn.Imported().Path())
					if e, ok := pl.tab.name[id.Name]; ok && e.dest == k {
						es = append(es, edit{off, len(id.Name), e.newName})
					} else if k != "" && pl.coreDecls[k] != nil && pl.coreDecls[k][upperFirst(id.Name)] {
						es = append(es, edit{off, len(id.Name), upperFirst(id.Name)})
					} else {
						problems = append(problems, fmt.Sprintf("%s: %s.%s is not defined", where(id), x.Name, id.Name))
					}
					return true
				}
			}
			tv, ok := info.Types[p.X]
			if !ok || !typed(tv.Type) {
				return true // the left side resolves on a later pass
			}
			if nn, ok := fieldFix(tv.Type, id.Name); ok {
				es = append(es, edit{off, len(id.Name), nn})
			} else {
				problems = append(problems, fmt.Sprintf("%s: %s has no field or method %s (export it where it is declared)", where(id), tv.Type, id.Name))
			}
			return true
		case *ast.KeyValueExpr:
			if p.Key != id {
				break
			}
			cl, ok := parent[p].(*ast.CompositeLit)
			if !ok {
				break
			}
			tv, ok := info.Types[cl]
			if !ok || !typed(tv.Type) {
				return true // the literal's type resolves on a later pass
			}
			if _, isStruct := tv.Type.Underlying().(*types.Struct); !isStruct {
				break // a map key: an ordinary name
			}
			if nn, ok := fieldFix(tv.Type, id.Name); ok {
				es = append(es, edit{off, len(id.Name), nn})
			} else {
				problems = append(problems, fmt.Sprintf("%s: %s has no field %s", where(id), tv.Type, id.Name))
			}
			return true
		}
		// A bare name.
		if e, ok := pl.tab.name[id.Name]; ok {
			switch {
			case e.dest == dest:
				if e.newName != id.Name {
					es = append(es, edit{off, len(id.Name), e.newName})
				}
			case isTestKey(e.dest):
				problems = append(problems, fmt.Sprintf("%s: %s lives in test package %s; move it to effectstest (exported) to share it", where(id), id.Name, pkgDir(e.dest)))
			case isShardKey(e.dest) && !isTestKey(dest) && dest != keyTst:
				problems = append(problems, fmt.Sprintf("%s: %s is declared in a card file in %s; move it into a core helper file to share it", where(id), id.Name, pkgDir(e.dest)))
			default:
				es = append(es, edit{off, len(id.Name), alias(e.dest) + "." + e.newName})
			}
			return true
		}
		for _, k := range []string{keyCore, keyTst} {
			if k == dest || (k == keyTst && !isTestKey(dest)) {
				continue
			}
			decls := pl.coreDecls[k]
			switch {
			case ast.IsExported(id.Name) && decls[id.Name]:
				es = append(es, edit{off, len(id.Name), alias(k) + "." + id.Name})
				return true
			case decls[upperFirst(id.Name)]:
				es = append(es, edit{off, len(id.Name), alias(k) + "." + upperFirst(id.Name)})
				return true
			case decls[id.Name]:
				// Unexported there: export it in place.
				if pl.exportReq[k] == nil {
					pl.exportReq[k] = map[string]bool{}
				}
				pl.exportReq[k][id.Name] = true
				es = append(es, edit{off, len(id.Name), alias(k) + "." + upperFirst(id.Name)})
				return true
			}
		}
		if !declared[id.Name] && types.Universe.Lookup(id.Name) == nil {
			problems = append(problems, fmt.Sprintf("%s: %s is not defined anywhere in the effects tree", where(id), id.Name))
		}
		return true
	})
	// Keys for unkeyed literals of another package's struct types.
	ast.Inspect(f.ast, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok || len(cl.Elts) == 0 {
			return true
		}
		if _, keyed := cl.Elts[0].(*ast.KeyValueExpr); keyed {
			return true
		}
		tv, ok := info.Types[cl]
		if !ok || tv.Type == nil {
			return true
		}
		nt, ok := tv.Type.(*types.Named)
		if !ok || nt.Obj().Pkg() == nil || keyOfImportPath(nt.Obj().Pkg().Path()) == "" || nt.Obj().Pkg().Path() == importPath(dest) {
			return true
		}
		st, ok := nt.Underlying().(*types.Struct)
		if !ok {
			return true
		}
		for i, el := range cl.Elts {
			if i < st.NumFields() {
				es = append(es, edit{f.tf.Offset(el.Pos()), 0, st.Field(i).Name() + ": "})
			}
		}
		return true
	})
	var adds []importSpec
	var ks []string
	for k := range need {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		spec := importSpec{path: importPath(k)}
		if !strings.HasSuffix(spec.path, "/"+pkgName(k)) {
			spec.name = pkgName(k)
		}
		adds = append(adds, spec)
	}
	return es, adds, problems
}

// exportInPlace exports unexported package-level names of the core or of
// effectstest that a placed file needs, renaming every use in that package.
func (pl *placer) exportInPlace(k string) error {
	_, files, info, pkg, err := pl.checkPkg(k)
	if err != nil {
		return err
	}
	edits := map[*srcFile][]edit{}
	byTF := map[*token.File]*srcFile{}
	for _, f := range files {
		byTF[f.tf] = f
	}
	var names []string
	for n := range pl.exportReq[k] {
		names = append(names, n)
		if pkg.Scope().Lookup(upperFirst(n)) != nil {
			return fmt.Errorf("cannot export %s in %s: %s already exists; rename it by hand", n, pkgDir(k), upperFirst(n))
		}
	}
	sort.Strings(names)
	add := func(m map[*ast.Ident]types.Object) {
		for id, o := range m {
			if o == nil || o.Parent() != pkg.Scope() || !pl.exportReq[k][o.Name()] {
				continue
			}
			for _, f := range files {
				if f.ast.Pos() <= id.Pos() && id.Pos() <= f.ast.End() {
					edits[f] = append(edits[f], edit{f.tf.Offset(id.Pos()), len(id.Name), upperFirst(id.Name)})
				}
			}
		}
	}
	add(info.Defs)
	add(info.Uses)
	for f, es := range edits {
		sortEdits(es)
		names := map[string]string{}
		for _, im := range f.ast.Imports {
			if pn, ok := info.Implicits[im].(*types.PkgName); ok {
				names[pn.Imported().Path()] = pn.Imported().Name()
			}
		}
		b, err := fixImports(applyEdits(f.src, 0, len(f.src), es), nil, names)
		if err != nil {
			return err
		}
		if err := os.WriteFile(f.path, b, 0o644); err != nil {
			return err
		}
	}
	logf("exported %s in %s: %s (other branches that use the old names: run --place on them too)", strings.Join(names, ", "), pkgDir(k), "every use renamed")
	return nil
}

func parseFileSkip(fset *token.FileSet, path string) (*ast.File, error) {
	return parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
}

// stubImporter answers for the aggregator with an empty package: a blank
// import needs nothing from it, and loading it would compile every shard.
type stubImporter struct{ next types.Importer }

func (s stubImporter) Import(path string) (*types.Package, error) {
	if path == effectsPath+"/cards" {
		p := types.NewPackage(path, "cards")
		p.MarkComplete()
		return p, nil
	}
	return s.next.Import(path)
}

// typed reports a type the checker could work out.
func typed(t types.Type) bool {
	if t == nil {
		return false
	}
	b, ok := t.(*types.Basic)
	return !ok || b.Kind() != types.Invalid
}

// brokenCoreFiles lists the core files with a name that does not resolve.
func (pl *placer) brokenCoreFiles() ([]string, error) {
	dir := filepath.Join(pl.effDir, pkgDir(keyCore))
	var names []string
	for _, n := range goFiles(dir) {
		if !isTestName(n) {
			names = append(names, n)
		}
	}
	fset := token.NewFileSet()
	files, err := parseDir(fset, dir, names)
	if err != nil {
		return nil, err
	}
	var asts []*ast.File
	for _, f := range files {
		asts = append(asts, f.ast)
	}
	bad := map[string]bool{}
	conf := types.Config{Importer: stubImporter{pl.ex.importer(fset)}, Error: func(err error) {
		if te, ok := err.(types.Error); ok && (strings.Contains(te.Msg, "undefined") || strings.Contains(te.Msg, "has no field or method")) {
			bad[fset.Position(te.Pos).Filename] = true
		}
	}}
	conf.Check(effectsPath, fset, asts, nil)
	var out []string
	for p := range bad {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}
