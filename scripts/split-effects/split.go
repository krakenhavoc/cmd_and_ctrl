package main

// split.go — the one type-checked pass that splits package effects.
//
// Placement rules (each a fixpoint over the use→declaration edges of the
// package's sources and tests):
//
//   - A non-test file with a top-level init() and no exported declaration is
//     a card file: it goes to the card shard its name picks in shards.txt.
//     Every other non-test file is core.
//   - A card file whose declarations another non-test file uses stays in core
//     (a shard file declares nothing another source file reads), and so does
//     one that declares a method of a core type, or a type with a core method.
//   - Each top-level declaration of a _test.go file is placed on its own. It
//     goes to the test package its file name picks, unless a declaration in
//     another package needs it: then it is a test helper and goes to
//     effectstest. Every init() and TestMain of a test file, and every
//     package-level var whose initializer registers something, also goes to
//     effectstest, so every test binary sees what the one test binary saw.
//   - Dependencies point one way: tests/* → effectstest → cards/* → core.
//
// Every unexported name that ends up used across a package boundary is
// exported (first letter upper-cased, or the name overrides.txt gives it),
// everywhere it is used. References to another package are qualified
// (effects.Register(effects.Spec{...})); there are no dot imports.

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type unit struct {
	f          *srcFile
	idx        int // index into f.ast.Decls; -1 for a whole (non-test) file
	decl       ast.Decl
	start, end int // byte range of the unit's text, comments before it included
	dest       string
	testFunc   bool
}

func (u *unit) String() string {
	if u.idx < 0 {
		return u.f.name
	}
	return fmt.Sprintf("%s#%d", u.f.name, u.idx)
}

type splitter struct {
	root, serverDir, effDir string
	tbl                     *rangeTable
	ov                      *overrides
	fset                    *token.FileSet
	files                   []*srcFile
	byTF                    map[*token.File]*srcFile
	pkg                     *types.Package
	info                    *types.Info
	units                   []*unit
	unitsOf                 map[*srcFile][]*unit
	fieldOwner              map[*types.Var]*types.TypeName
	fieldStruct             map[*types.Var]*types.Struct
	pkgNames                map[string]string // import path → package name
	pathRewrites            int
	litRenames              int
	lineRemapCount          int
}

func isTestName(n string) bool { return strings.HasSuffix(n, "_test.go") }

func (s *splitter) fileOf(p token.Pos) *srcFile {
	tf := s.fset.File(p)
	if tf == nil {
		return nil
	}
	return s.byTF[tf]
}

func (s *splitter) unitOf(p token.Pos) *unit {
	f := s.fileOf(p)
	if f == nil {
		return nil
	}
	us := s.unitsOf[f]
	if len(us) == 1 && us[0].idx < 0 {
		return us[0]
	}
	off := f.tf.Offset(p)
	i := sort.Search(len(us), func(i int) bool { return us[i].end > off })
	if i < len(us) && us[i].start <= off {
		return us[i]
	}
	return nil
}

func (s *splitter) offset(p token.Pos) int { return s.fileOf(p).tf.Offset(p) }

func isTestFuncName(n string) bool {
	for _, pre := range []string{"Test", "Benchmark", "Fuzz", "Example"} {
		if n == pre {
			return true
		}
		if strings.HasPrefix(n, pre) {
			r, _ := utf8.DecodeRuneInString(n[len(pre):])
			if !unicode.IsLower(r) {
				return n != "TestMain"
			}
		}
	}
	return false
}

func upperFirst(n string) string {
	r, sz := utf8.DecodeRuneInString(n)
	if r == '_' {
		return "U" + n
	}
	return string(unicode.ToUpper(r)) + n[sz:]
}

// load parses and type-checks the old package.
func (s *splitter) load() error {
	deps, err := testDepsOfEffects(s.serverDir)
	if err != nil {
		return err
	}
	ex := newExportData(s.serverDir)
	logf("listing export data for %d dependencies of the effects tests", len(deps))
	if err := ex.add(deps...); err != nil {
		return err
	}
	s.fset = token.NewFileSet()
	names := goFiles(s.effDir)
	logf("parsing %d files", len(names))
	s.files, err = parseDir(s.fset, s.effDir, names)
	if err != nil {
		return err
	}
	s.byTF = map[*token.File]*srcFile{}
	var asts []*ast.File
	for _, f := range s.files {
		if f.ast.Name.Name != "effects" {
			return fmt.Errorf("%s: package %s; the splitter reads package effects only (an external test package needs a hand move first)", f.name, f.ast.Name.Name)
		}
		s.byTF[f.tf] = f
		asts = append(asts, f.ast)
	}
	s.info = &types.Info{
		Uses:       map[*ast.Ident]types.Object{},
		Defs:       map[*ast.Ident]types.Object{},
		Types:      map[ast.Expr]types.TypeAndValue{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
		Implicits:  map[ast.Node]types.Object{},
	}
	var terrs []string
	conf := types.Config{Importer: ex.importer(s.fset), Error: func(err error) {
		if len(terrs) < 20 {
			terrs = append(terrs, err.Error())
		}
	}}
	logf("type-checking package effects (sources and tests)")
	s.pkg, _ = conf.Check(effectsPath, s.fset, asts, s.info)
	if len(terrs) > 0 {
		return fmt.Errorf("package effects does not type-check:\n  %s", strings.Join(terrs, "\n  "))
	}
	return nil
}

// buildUnits splits files into units and makes the first assignment.
func (s *splitter) buildUnits() {
	s.unitsOf = map[*srcFile][]*unit{}
	for _, f := range s.files {
		if !isTestName(f.name) {
			u := &unit{f: f, idx: -1, start: 0, end: len(f.src)}
			if s.isCardFile(f) {
				u.dest = s.tbl.cardShard(f.name)
			} else {
				u.dest = keyCore
			}
			s.units = append(s.units, u)
			s.unitsOf[f] = []*unit{u}
			continue
		}
		// The text before the first non-import declaration belongs to the
		// file (package clause, imports); each unit owns the gap before it.
		prevEnd := f.tf.Offset(f.ast.Name.End())
		var us []*unit
		for i, d := range f.ast.Decls {
			if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
				prevEnd = f.tf.Offset(g.End())
				continue
			}
			u := &unit{f: f, idx: i, decl: d, start: prevEnd, end: f.tf.Offset(d.End())}
			prevEnd = u.end
			u.dest = s.tbl.testPkg(f.name)
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil && (d.Name.Name == "init" || d.Name.Name == "TestMain") {
					u.dest = keyTst
				}
				u.testFunc = d.Recv == nil && isTestFuncName(d.Name.Name)
			case *ast.GenDecl:
				if d.Tok == token.VAR && registers(d) {
					u.dest = keyTst
				}
			}
			us = append(us, u)
		}
		if len(us) > 0 {
			us[len(us)-1].end = len(f.src)
		}
		s.units = append(s.units, us...)
		s.unitsOf[f] = us
	}
}

// registers reports whether a var declaration's initializer calls a function
// whose name starts with Register or register (a side effect at init time).
func registers(d *ast.GenDecl) bool {
	found := false
	ast.Inspect(d, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		name := ""
		switch f := c.Fun.(type) {
		case *ast.Ident:
			name = f.Name
		case *ast.SelectorExpr:
			name = f.Sel.Name
		}
		if strings.HasPrefix(name, "Register") || strings.HasPrefix(name, "register") {
			found = true
		}
		return !found
	})
	return found
}

func (s *splitter) isCardFile(f *srcFile) bool {
	init, exported := false, false
	for _, d := range f.ast.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil && d.Name.Name == "init" {
				init = true
			}
			if d.Recv == nil && ast.IsExported(d.Name.Name) {
				exported = true
			}
		case *ast.GenDecl:
			for _, sp := range d.Specs {
				switch sp := sp.(type) {
				case *ast.ValueSpec:
					for _, id := range sp.Names {
						if ast.IsExported(id.Name) {
							exported = true
						}
					}
				case *ast.TypeSpec:
					if ast.IsExported(sp.Name.Name) {
						exported = true
					}
				}
			}
		}
	}
	return init && !exported
}

// inScope: package-level objects, fields and methods of package effects.
func (s *splitter) inScope(o types.Object) bool {
	if o == nil || o.Pkg() != s.pkg || !o.Pos().IsValid() {
		return false
	}
	if o.Parent() == s.pkg.Scope() {
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

func isPkgLevel(o types.Object, pkg *types.Package) bool {
	return o != nil && o.Pkg() == pkg && o.Parent() == pkg.Scope()
}

// allowed: may a unit in package a use a declaration in package b?
func allowed(a, b string) bool {
	switch {
	case b == keyCore:
		return true
	case isShardKey(b):
		// A shard file is read by its own file and by tests only.
		return a == keyTst || isTestKey(a)
	case b == keyTst:
		return a == keyTst || isTestKey(a)
	case isTestKey(b):
		return a == b
	}
	return false
}

type edgeKey struct{ use, decl *unit }

func (s *splitter) place() error {
	edges := map[edgeKey]types.Object{}
	for id, o := range s.info.Uses {
		if !s.inScope(o) {
			continue
		}
		u, d := s.unitOf(id.Pos()), s.unitOf(o.Pos())
		if u == nil || d == nil || u == d {
			continue
		}
		k := edgeKey{u, d}
		if prev, ok := edges[k]; !ok || prev.Pos() > o.Pos() {
			edges[k] = o
		}
	}
	var ekeys []edgeKey
	for k := range edges {
		ekeys = append(ekeys, k)
	}
	sort.Slice(ekeys, func(i, j int) bool {
		a, b := ekeys[i], ekeys[j]
		if a.use.f.name != b.use.f.name {
			return a.use.f.name < b.use.f.name
		}
		if a.use.idx != b.use.idx {
			return a.use.idx < b.use.idx
		}
		if a.decl.f.name != b.decl.f.name {
			return a.decl.f.name < b.decl.f.name
		}
		return a.decl.idx < b.decl.idx
	})
	// Methods live with their receiver's type.
	var coloc [][2]*unit
	for id, o := range s.info.Defs {
		fn, ok := o.(*types.Func)
		if !ok || fn.Pkg() != s.pkg {
			continue
		}
		recv := fn.Type().(*types.Signature).Recv()
		if recv == nil {
			continue
		}
		t := recv.Type()
		if p, ok := t.(*types.Pointer); ok {
			t = p.Elem()
		}
		nt, ok := t.(*types.Named)
		if !ok || nt.Obj().Pkg() != s.pkg {
			continue
		}
		a, b := s.unitOf(id.Pos()), s.unitOf(nt.Obj().Pos())
		if a != nil && b != nil && a != b {
			coloc = append(coloc, [2]*unit{a, b})
		}
	}
	sort.Slice(coloc, func(i, j int) bool {
		return coloc[i][0].String()+coloc[i][1].String() < coloc[j][0].String()+coloc[j][1].String()
	})
	for round := 0; ; round++ {
		changed := false
		set := func(u *unit, k string) {
			if u.dest != k {
				u.dest = k
				changed = true
			}
		}
		for _, e := range ekeys {
			pu, pd := e.use.dest, e.decl.dest
			if allowed(pu, pd) {
				continue
			}
			switch {
			case e.decl.testFunc:
				// A test function another test calls: keep them together.
				if isTestKey(pu) {
					set(e.use, pd)
				}
			case isTestName(e.decl.f.name):
				set(e.decl, keyTst)
			default:
				set(e.decl, keyCore)
			}
		}
		for _, c := range coloc {
			a, b := c[0], c[1]
			if a.dest == b.dest {
				continue
			}
			if isTestName(a.f.name) || isTestName(b.f.name) {
				if !isTestName(a.f.name) || !isTestName(b.f.name) {
					return fmt.Errorf("a method and its type are split between a source and a test file: %s, %s", a, b)
				}
				set(a, keyTst)
				set(b, keyTst)
			} else {
				set(a, keyCore)
				set(b, keyCore)
			}
		}
		if !changed {
			break
		}
		if round > 100 {
			return fmt.Errorf("placement did not converge")
		}
	}
	var bad []string
	for _, e := range ekeys {
		if !allowed(e.use.dest, e.decl.dest) {
			bad = append(bad, fmt.Sprintf("%s (%s) uses %s (%s): %s", e.use, e.use.dest, e.decl, e.decl.dest, edges[e].Name()))
		}
	}
	if len(bad) > 0 {
		return fmt.Errorf("unresolved placements:\n  %s", strings.Join(bad, "\n  "))
	}
	return nil
}

func (s *splitter) destOf(o types.Object) string {
	u := s.unitOf(o.Pos())
	if u == nil {
		return ""
	}
	return u.dest
}

// fmtDanger are method names with meaning to the standard library: exporting
// an unexported method to one of these would change behaviour.
var fmtDanger = map[string]bool{
	"String": true, "Error": true, "Format": true, "GoString": true, "Unwrap": true,
	"Is": true, "As": true, "MarshalJSON": true, "UnmarshalJSON": true, "MarshalText": true,
	"UnmarshalText": true, "Len": true, "Less": true, "Swap": true, "Read": true,
	"Write": true, "Close": true, "ServeHTTP": true, "Scan": true, "Value": true,
}

type renamer struct {
	s        *splitter
	need     map[types.Object]bool
	newName  map[types.Object]string
	keyLits  []*ast.CompositeLit // unkeyed literals that cross a package boundary
	auto     []string            // names chosen without an override after a collision
	conflict []string
}

func (s *splitter) structOf(t types.Type) (*types.Struct, *types.Named) {
	nt, _ := t.(*types.Named)
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
		nt, _ = t.(*types.Named)
	}
	st, ok := t.Underlying().(*types.Struct)
	if !ok {
		return nil, nil
	}
	return st, nt
}

func (s *splitter) computeRenames() (*renamer, error) {
	r := &renamer{s: s, need: map[types.Object]bool{}, newName: map[types.Object]string{}}
	// 1. Names used across a package boundary.
	for id, o := range s.info.Uses {
		if !s.inScope(o) || o.Exported() {
			continue
		}
		u := s.unitOf(id.Pos())
		if u != nil && u.dest != s.destOf(o) {
			r.need[o] = true
		}
	}
	// 2. Unkeyed composite literals of a struct type from another package: the
	//    tool writes the keys, so every field given a value must be exported.
	for _, f := range s.files {
		ast.Inspect(f.ast, func(n ast.Node) bool {
			cl, ok := n.(*ast.CompositeLit)
			if !ok || len(cl.Elts) == 0 {
				return true
			}
			if _, keyed := cl.Elts[0].(*ast.KeyValueExpr); keyed {
				return true
			}
			tv, ok := s.info.Types[cl]
			if !ok {
				return true
			}
			st, nt := s.structOf(tv.Type)
			if st == nil || nt == nil || nt.Obj().Pkg() != s.pkg {
				return true
			}
			u := s.unitOf(cl.Pos())
			if u == nil || u.dest == s.destOf(nt.Obj()) {
				return true
			}
			r.keyLits = append(r.keyLits, cl)
			for i := 0; i < st.NumFields() && i < len(cl.Elts); i++ {
				if !st.Field(i).Exported() {
					r.need[st.Field(i)] = true
				}
			}
			return true
		})
	}
	// 3. A type in one package that satisfies an interface of another, whose
	//    methods are unexported, needs those methods exported.
	scope := s.pkg.Scope()
	var ifaces []*types.TypeName
	var named []*types.TypeName
	for _, n := range scope.Names() {
		tn, ok := scope.Lookup(n).(*types.TypeName)
		if !ok || tn.IsAlias() {
			continue
		}
		if it, ok := tn.Type().Underlying().(*types.Interface); ok {
			hasUnexported := false
			for i := 0; i < it.NumMethods(); i++ {
				if !it.Method(i).Exported() {
					hasUnexported = true
				}
			}
			if hasUnexported {
				ifaces = append(ifaces, tn)
			}
			continue
		}
		named = append(named, tn)
	}
	for _, it := range ifaces {
		iface := it.Type().Underlying().(*types.Interface)
		for _, tn := range named {
			if s.destOf(tn) == s.destOf(it) {
				continue
			}
			if _, ok := tn.Type().(*types.Named); ok && tn.Type().(*types.Named).TypeParams().Len() > 0 {
				continue
			}
			if types.Implements(tn.Type(), iface) || types.Implements(types.NewPointer(tn.Type()), iface) {
				for i := 0; i < iface.NumMethods(); i++ {
					if m := iface.Method(i); !m.Exported() {
						r.need[m] = true
					}
				}
			}
		}
	}
	// 4. A method name is renamed everywhere at once, so implementations and
	//    interfaces stay in step.
	methodNames := map[string]bool{}
	for o := range r.need {
		if isMethod(o) {
			methodNames[o.Name()] = true
		}
	}
	allObjs := map[types.Object]bool{}
	for _, o := range s.info.Defs {
		if o != nil && o.Pkg() == s.pkg {
			allObjs[o] = true
		}
	}
	for o := range allObjs {
		if isMethod(o) && methodNames[o.Name()] {
			r.need[o] = true
		}
	}
	// Interface methods declared in interface literals are Defs too.
	// 5. An embedded field takes its type's name.
	for o := range allObjs {
		v, ok := o.(*types.Var)
		if !ok || !v.Embedded() {
			continue
		}
		if tn := embeddedTypeName(v.Type()); tn != nil && tn.Pkg() == s.pkg && r.need[tn] {
			r.need[v] = true
		}
	}
	// 6. An unnamed struct type with unexported fields in the type of a name
	//    used across a package boundary: its fields, and those of every
	//    identical unnamed struct in the package, are exported, so the type is
	//    still identical on both sides (unexported field names are
	//    package-qualified).
	var unnamed []*types.Struct
	seenSt := map[*types.Struct]bool{}
	for _, f := range s.files {
		ast.Inspect(f.ast, func(n ast.Node) bool {
			if st, ok := n.(*ast.StructType); ok {
				if t, ok := s.info.Types[st].Type.(*types.Struct); ok && !seenSt[t] {
					seenSt[t] = true
					unnamed = append(unnamed, t)
				}
			}
			return true
		})
	}
	var seeds []*types.Struct
	crossObjs := map[types.Object]bool{}
	for id, o := range s.info.Uses {
		if isPkgLevel(o, s.pkg) {
			if u := s.unitOf(id.Pos()); u != nil && u.dest != s.destOf(o) {
				crossObjs[o] = true
			}
		}
	}
	for o := range crossObjs {
		walkUnnamedStructs(o.Type(), map[types.Type]bool{}, func(st *types.Struct) {
			for i := 0; i < st.NumFields(); i++ {
				if !st.Field(i).Exported() {
					seeds = append(seeds, st)
					return
				}
			}
		})
	}
	for _, seed := range seeds {
		for _, st := range unnamed {
			if !types.Identical(seed, st) {
				continue
			}
			for i := 0; i < st.NumFields(); i++ {
				if f := st.Field(i); !f.Exported() && f.Pkg() == s.pkg {
					r.need[f] = true
				}
			}
		}
	}
	if err := r.name(); err != nil {
		return nil, err
	}
	return r, nil
}

// walkUnnamedStructs calls fn for every unnamed struct type in t, through
// pointers, slices, arrays, maps, channels and signatures (not into named types).
func walkUnnamedStructs(t types.Type, seen map[types.Type]bool, fn func(*types.Struct)) {
	if seen[t] {
		return
	}
	seen[t] = true
	switch t := t.(type) {
	case *types.Struct:
		fn(t)
		for i := 0; i < t.NumFields(); i++ {
			walkUnnamedStructs(t.Field(i).Type(), seen, fn)
		}
	case *types.Pointer:
		walkUnnamedStructs(t.Elem(), seen, fn)
	case *types.Slice:
		walkUnnamedStructs(t.Elem(), seen, fn)
	case *types.Array:
		walkUnnamedStructs(t.Elem(), seen, fn)
	case *types.Map:
		walkUnnamedStructs(t.Key(), seen, fn)
		walkUnnamedStructs(t.Elem(), seen, fn)
	case *types.Chan:
		walkUnnamedStructs(t.Elem(), seen, fn)
	case *types.Signature:
		for _, tup := range []*types.Tuple{t.Params(), t.Results()} {
			for i := 0; i < tup.Len(); i++ {
				walkUnnamedStructs(tup.At(i).Type(), seen, fn)
			}
		}
	}
}

func isMethod(o types.Object) bool {
	fn, ok := o.(*types.Func)
	return ok && fn.Type().(*types.Signature).Recv() != nil
}

func embeddedTypeName(t types.Type) *types.TypeName {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	switch t := t.(type) {
	case *types.Named:
		return t.Obj()
	case *types.Alias:
		return t.Obj()
	}
	return nil
}

func (s *splitter) posKey(o types.Object) string {
	p := s.fset.Position(o.Pos())
	return fmt.Sprintf("%s:%09d", filepath.Base(p.Filename), p.Offset)
}

// fieldKey names a field for overrides.txt: <Type>.<field> with the type's
// old name, or ?<file>.<field> for a field of an unnamed struct.
func (s *splitter) fieldKey(v *types.Var) string {
	if tn := s.fieldOwner[v]; tn != nil {
		return tn.Name() + "." + v.Name()
	}
	return "?" + filepath.Base(s.fset.Position(v.Pos()).Filename) + "." + v.Name()
}

func (r *renamer) name() error {
	s := r.s
	s.indexFieldOwners()
	var objs []types.Object
	for o := range r.need {
		objs = append(objs, o)
	}
	sort.Slice(objs, func(i, j int) bool { return s.posKey(objs[i]) < s.posKey(objs[j]) })

	// Names taken in each destination package scope.
	taken := map[string]map[string]types.Object{}
	scope := s.pkg.Scope()
	for _, n := range scope.Names() {
		o := scope.Lookup(n)
		d := s.destOf(o)
		if taken[d] == nil {
			taken[d] = map[string]types.Object{}
		}
		taken[d][n] = o
	}
	uses := map[types.Object][]*ast.Ident{}
	for id, o := range s.info.Uses {
		if r.need[o] {
			uses[o] = append(uses[o], id)
		}
	}
	for _, ids := range uses {
		sort.Slice(ids, func(i, j int) bool { return ids[i].Pos() < ids[j].Pos() })
	}
	selsOf := map[types.Object][]*ast.SelectorExpr{}
	for se, sel := range s.info.Selections {
		if r.need[sel.Obj()] {
			selsOf[sel.Obj()] = append(selsOf[sel.Obj()], se)
		}
	}
	// conflicts reports why nn cannot be o's new name, or "".
	shadowed := func(o types.Object, nn string) string {
		d := s.destOf(o)
		for _, id := range uses[o] {
			u := s.unitOf(id.Pos())
			if u == nil || u.dest != d {
				continue // qualified there
			}
			inner := s.pkg.Scope().Innermost(id.Pos())
			if inner == nil {
				continue
			}
			if _, x := inner.LookupParent(nn, id.Pos()); x != nil && x.Parent() != s.pkg.Scope() && x.Parent() != types.Universe {
				return fmt.Sprintf("a local %s at %s", nn, s.fset.Position(id.Pos()))
			}
		}
		return ""
	}
	selConflict := func(o types.Object, nn string) string {
		for _, se := range selsOf[o] {
			sel := s.info.Selections[se]
			if x, _, _ := types.LookupFieldOrMethod(sel.Recv(), true, s.pkg, nn); x != nil {
				return fmt.Sprintf("%s already has %s (%s)", sel.Recv(), nn, s.fset.Position(se.Pos()))
			}
		}
		return ""
	}
	methodNew := map[string]string{}
	fieldNew := map[*types.Var]string{}
	// Package-level objects and methods first; embedded fields follow their types.
	for pass := 0; pass < 2; pass++ {
		for _, o := range objs {
			v, isVar := o.(*types.Var)
			embedded := isVar && v.Embedded()
			if (pass == 0) == embedded {
				continue
			}
			switch {
			case isPkgLevel(o, s.pkg):
				d := s.destOf(o)
				nn, hasOv := r.s.ov.name[o.Name()]
				if !hasOv {
					nn = upperFirst(o.Name())
				}
				why := ""
				if x, ok := taken[d][nn]; ok && x != o {
					why = "the package already has " + nn
				} else if w := shadowed(o, nn); w != "" {
					why = w
				}
				if why != "" {
					if hasOv {
						r.conflict = append(r.conflict, fmt.Sprintf("name %s: override %s does not work: %s", o.Name(), nn, why))
						continue
					}
					base := nn
					for i := 0; why != ""; i++ {
						nn = "X" + nn
						why = ""
						if x, ok := taken[d][nn]; ok && x != o {
							why = "taken"
						} else if w := shadowed(o, nn); w != "" {
							why = w
						}
						if i > 5 {
							return fmt.Errorf("no name for %s", o.Name())
						}
					}
					r.auto = append(r.auto, fmt.Sprintf("name %s %s\t# %s was taken (%s)", o.Name(), nn, base, s.fset.Position(o.Pos())))
				}
				taken[d][nn] = o
				r.newName[o] = nn
			case isMethod(o):
				if nn, ok := methodNew[o.Name()]; ok {
					r.newName[o] = nn
					continue
				}
				nn, hasOv := r.s.ov.method[o.Name()]
				if !hasOv {
					nn = upperFirst(o.Name())
				}
				// Every type with a method of this name, and every selector of one.
				why := ""
				if fmtDanger[nn] {
					why = nn + " means something to the standard library"
				}
				for _, o2 := range objs {
					if why != "" {
						break
					}
					if !isMethod(o2) || o2.Name() != o.Name() {
						continue
					}
					recv := o2.Type().(*types.Signature).Recv().Type()
					if x, _, _ := types.LookupFieldOrMethod(recv, true, s.pkg, nn); x != nil {
						why = fmt.Sprintf("%s already has %s", recv, nn)
					} else if w := selConflict(o2, nn); w != "" {
						why = w
					}
				}
				if why != "" {
					if hasOv {
						r.conflict = append(r.conflict, fmt.Sprintf("method %s: override %s does not work: %s", o.Name(), nn, why))
						continue
					}
					base := nn
					nn = "X" + nn
					r.auto = append(r.auto, fmt.Sprintf("method %s %s\t# %s: %s", o.Name(), nn, base, why))
				}
				methodNew[o.Name()] = nn
				r.newName[o] = nn
			case isVar && v.IsField():
				var nn string
				hasOv := false
				if embedded {
					tn := embeddedTypeName(v.Type())
					nn = r.newName[tn]
					if nn == "" {
						nn = tn.Name()
					}
				} else {
					nn, hasOv = r.s.ov.field[s.fieldKey(v)]
					if !hasOv {
						nn = upperFirst(o.Name())
					}
				}
				why := ""
				if owner := s.fieldOwner[v]; owner != nil {
					if x, _, _ := types.LookupFieldOrMethod(owner.Type(), true, s.pkg, nn); x != nil && x != o {
						why = fmt.Sprintf("%s already has %s", owner.Name(), nn)
					}
				}
				for v2, n2 := range fieldNew {
					if n2 == nn && s.fieldStruct[v2] == s.fieldStruct[v] && s.fieldStruct[v] != nil {
						why = "two fields would both be " + nn
					}
				}
				if why == "" {
					why = selConflict(o, nn)
				}
				if why != "" {
					if hasOv || embedded {
						r.conflict = append(r.conflict, fmt.Sprintf("field %s: %s does not work: %s", s.fieldKey(v), nn, why))
						continue
					}
					base := nn
					nn = "X" + nn
					r.auto = append(r.auto, fmt.Sprintf("field %s %s\t# %s: %s", s.fieldKey(v), nn, base, why))
				}
				fieldNew[v] = nn
				r.newName[o] = nn
			default:
				return fmt.Errorf("cannot rename %v", o)
			}
		}
	}
	if len(r.conflict) > 0 {
		return fmt.Errorf("rename conflicts:\n  %s", strings.Join(r.conflict, "\n  "))
	}
	return nil
}

// indexFieldOwners maps each field of a struct to the named type that
// declares it (nil for an unnamed struct) and to its struct.
func (s *splitter) indexFieldOwners() {
	s.fieldOwner = map[*types.Var]*types.TypeName{}
	s.fieldStruct = map[*types.Var]*types.Struct{}
	seen := map[*types.Struct]bool{}
	var walk func(st *types.Struct, owner *types.TypeName)
	walk = func(st *types.Struct, owner *types.TypeName) {
		if seen[st] {
			return
		}
		seen[st] = true
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			if f.Pkg() != s.pkg {
				continue
			}
			if s.fieldOwner[f] == nil {
				s.fieldOwner[f] = owner
			}
			s.fieldStruct[f] = st
		}
	}
	// Named types first, so a struct's owner is its named type.
	var tns []*types.TypeName
	for _, o := range s.info.Defs {
		if tn, ok := o.(*types.TypeName); ok && tn.Pkg() == s.pkg && !tn.IsAlias() {
			tns = append(tns, tn)
		}
	}
	sort.Slice(tns, func(i, j int) bool { return s.posKey(tns[i]) < s.posKey(tns[j]) })
	for _, tn := range tns {
		if st, ok := tn.Type().Underlying().(*types.Struct); ok {
			walk(st, tn)
		}
	}
	for _, tv := range s.info.Types {
		if st, ok := tv.Type.(*types.Struct); ok {
			walk(st, nil)
		}
	}
}

// edit replaces src[off:off+n] with s. Inserts (n == 0) at an offset go
// before a replacement at the same offset.
type edit struct {
	off, n int
	s      string
}

func applyEdits(src []byte, start, end int, es []edit) []byte {
	var out bytes.Buffer
	pos := start
	for _, e := range es {
		if e.off < start || e.off+e.n > end {
			continue
		}
		if e.off < pos {
			continue // overlapping: the first one wins
		}
		out.Write(src[pos:e.off])
		out.WriteString(e.s)
		pos = e.off + e.n
	}
	out.Write(src[pos:end])
	return out.Bytes()
}

func sortEdits(es []edit) {
	sort.SliceStable(es, func(i, j int) bool {
		if es[i].off != es[j].off {
			return es[i].off < es[j].off
		}
		return es[i].n < es[j].n
	})
}

// piece is one output file: a whole file, or the units of a test file that
// share a destination.
type piece struct {
	f        *srcFile
	dest     string
	units    []*unit
	whole    bool
	main     bool   // the piece that keeps the file's package clause and name
	outName  string // file name in the destination directory
	foreign  map[string][]token.Pos
	alias    map[string]string
	dropPkgs map[*ast.Ident]bool
}

func (s *splitter) pieces() []*piece {
	var out []*piece
	for _, f := range s.files {
		us := s.unitsOf[f]
		if !isTestName(f.name) {
			out = append(out, &piece{f: f, dest: us[0].dest, units: us, whole: true, main: true, outName: f.name})
			continue
		}
		byDest := map[string][]*unit{}
		var dests []string
		for _, u := range us {
			if _, ok := byDest[u.dest]; !ok {
				dests = append(dests, u.dest)
			}
			byDest[u.dest] = append(byDest[u.dest], u)
		}
		if len(dests) == 0 {
			// A test file with no declarations: it goes to its test package.
			dests = []string{s.tbl.testPkg(f.name)}
		}
		home := s.tbl.testPkg(f.name)
		if _, ok := byDest[home]; !ok {
			home = dests[0]
		}
		for _, d := range dests {
			p := &piece{f: f, dest: d, units: byDest[d], whole: len(dests) == 1, main: d == home}
			if d == keyTst {
				p.outName = helperFileName(f.name)
			} else {
				p.outName = f.name
			}
			out = append(out, p)
		}
	}
	return out
}

var buildSuffixes = map[string]bool{}

func init() {
	for _, s := range strings.Fields("aix android darwin dragonfly freebsd hurd illumos ios js linux nacl netbsd openbsd plan9 solaris wasip1 windows zos 386 amd64 amd64p32 arm armbe arm64 arm64be loong64 mips mipsle mips64 mips64le mips64p32 mips64p32le ppc ppc64 ppc64le riscv riscv64 s390 s390x sparc sparc64 wasm test") {
		buildSuffixes[s] = true
	}
}

// helperFileName is the effectstest name of a test file's helpers.
func helperFileName(n string) string {
	stem := strings.TrimSuffix(n, "_test.go")
	if i := strings.LastIndex(stem, "_"); i >= 0 && buildSuffixes[stem[i+1:]] {
		stem += "_helpers"
	}
	return stem + ".go"
}

func logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "split-effects: "+format+"\n", args...)
}

// write is the whole rewrite: edits, pieces, files.
func (s *splitter) write(r *renamer) (*splitResult, error) {
	res := &splitResult{}
	ps := s.pieces()
	pieceOf := map[*unit]*piece{}
	for _, p := range ps {
		p.foreign = map[string][]token.Pos{}
		p.alias = map[string]string{}
		for _, u := range p.units {
			pieceOf[u] = p
		}
	}
	pieceAt := func(pos token.Pos) *piece {
		u := s.unitOf(pos)
		if u == nil {
			return nil
		}
		return pieceOf[u]
	}
	edits := map[*srcFile][]edit{}
	// Qualified references: collect, so each piece can pick its aliases.
	type qual struct {
		id  *ast.Ident
		obj types.Object
		p   *piece
	}
	var quals []qual
	identDone := map[*ast.Ident]bool{}
	handle := func(id *ast.Ident, o types.Object, use bool) {
		if identDone[id] || o == nil || o.Pkg() != s.pkg {
			return
		}
		nn, renamed := r.newName[o]
		if use && isPkgLevel(o, s.pkg) {
			p := pieceAt(id.Pos())
			if p != nil && p.dest != s.destOf(o) {
				identDone[id] = true
				quals = append(quals, qual{id, o, p})
				p.foreign[s.destOf(o)] = append(p.foreign[s.destOf(o)], id.Pos())
				return
			}
		}
		if renamed {
			identDone[id] = true
			f := s.fileOf(id.Pos())
			edits[f] = append(edits[f], edit{f.tf.Offset(id.Pos()), len(id.Name), nn})
		}
	}
	// Uses first: an embedded field's ident is both a use (the type) and a def.
	var useIDs []*ast.Ident
	for id := range s.info.Uses {
		useIDs = append(useIDs, id)
	}
	sort.Slice(useIDs, func(i, j int) bool { return useIDs[i].Pos() < useIDs[j].Pos() })
	for _, id := range useIDs {
		handle(id, s.info.Uses[id], true)
	}
	var defIDs []*ast.Ident
	for id := range s.info.Defs {
		defIDs = append(defIDs, id)
	}
	sort.Slice(defIDs, func(i, j int) bool { return defIDs[i].Pos() < defIDs[j].Pos() })
	for _, id := range defIDs {
		handle(id, s.info.Defs[id], false)
	}
	// Keys for unkeyed literals that now cross a package boundary.
	for _, cl := range r.keyLits {
		st, _ := s.structOf(s.info.Types[cl].Type)
		f := s.fileOf(cl.Pos())
		for i, el := range cl.Elts {
			fv := st.Field(i)
			name := fv.Name()
			if nn, ok := r.newName[fv]; ok {
				name = nn
			}
			edits[f] = append(edits[f], edit{f.tf.Offset(el.Pos()), 0, name + ": "})
		}
		res.keyed++
	}
	// Test code that moved: paths, and the source scanners.
	tstRefs := s.testRewrites(ps, r, edits, pieceOf)
	for _, tr := range tstRefs {
		if tr.p.dest != keyTst {
			tr.p.foreign[keyTst] = append(tr.p.foreign[keyTst], tr.pos)
		}
	}
	// Aliases.
	for _, p := range ps {
		var ds []string
		for d := range p.foreign {
			ds = append(ds, d)
		}
		sort.Strings(ds)
		used := map[string]bool{}
		for _, d := range ds {
			alias := ""
			for _, cand := range []string{pkgName(d), pkgName(d) + "pkg", pkgName(d) + "x"} {
				if used[cand] || s.aliasConflict(p, cand, p.foreign[d]) {
					continue
				}
				alias = cand
				break
			}
			if alias == "" {
				return nil, fmt.Errorf("%s: no free import name for %s", p.f.name, importPath(d))
			}
			used[alias] = true
			p.alias[d] = alias
		}
	}
	for _, q := range quals {
		nn := q.obj.Name()
		if x, ok := r.newName[q.obj]; ok {
			nn = x
		}
		f := q.p.f
		edits[f] = append(edits[f], edit{f.tf.Offset(q.id.Pos()), len(q.id.Name), q.p.alias[s.destOf(q.obj)] + "." + nn})
	}
	for _, tr := range tstRefs {
		f := tr.p.f
		qual := ""
		if tr.p.dest != keyTst {
			qual = tr.p.alias[keyTst] + "."
		}
		edits[f] = append(edits[f], edit{tr.off, tr.n, qual + tr.name})
	}
	for f := range edits {
		sortEdits(edits[f])
	}
	// Write every piece: sources first, so the tests can learn how many
	// lines each source gained at its top.
	deltas := map[string]int{}
	for phase := 0; phase < 2; phase++ {
		if phase == 1 {
			s.lineRemaps(deltas, edits)
			for f := range edits {
				sortEdits(edits[f])
			}
		}
		for _, p := range ps {
			if isTestName(p.f.name) != (phase == 1) {
				continue
			}
			b, err := s.render(p, edits[p.f])
			if err != nil {
				return nil, err
			}
			if phase == 0 {
				deltas[p.f.name] = bytes.Count(b, []byte("\n")) - bytes.Count(p.f.src, []byte("\n"))
			}
			dir := filepath.Join(s.effDir, pkgDir(p.dest))
			out := filepath.Join(dir, p.outName)
			res.files = append(res.files, outFile{from: p.f.path, to: out, content: b, main: p.main})
		}
	}
	return res, nil
}

func (s *splitter) aliasConflict(p *piece, cand string, poss []token.Pos) bool {
	// A package-level name in the destination package.
	scope := s.pkg.Scope()
	if o := scope.Lookup(cand); o != nil && s.destOf(o) == p.dest {
		return true
	}
	// An import of the file with that name.
	for _, im := range p.f.ast.Imports {
		path, _ := strconv.Unquote(im.Path.Value)
		name := path[strings.LastIndex(path, "/")+1:]
		if im.Name != nil {
			name = im.Name.Name
		} else if pn, ok := s.info.Implicits[im].(*types.PkgName); ok {
			name = pn.Name()
		}
		if name == cand {
			return true
		}
	}
	// A local that would shadow it at a use.
	for _, pos := range poss {
		inner := scope.Innermost(pos)
		if inner == nil {
			continue
		}
		if _, x := inner.LookupParent(cand, pos); x != nil && x.Parent() != scope && x.Parent() != types.Universe {
			return true
		}
	}
	return false
}

type outFile struct {
	from, to string
	content  []byte
	main     bool
}

type splitResult struct {
	files []outFile
	keyed int
}

// render produces one output file.
func (s *splitter) render(p *piece, es []edit) ([]byte, error) {
	f := p.f
	var body []byte
	var hdr bytes.Buffer
	pkgStart := f.tf.Offset(f.ast.Package)
	nameStart := f.tf.Offset(f.ast.Name.Pos())
	nameEnd := f.tf.Offset(f.ast.Name.End())
	if p.whole {
		es2 := append([]edit(nil), es...)
		es2 = append(es2, edit{nameStart, nameEnd - nameStart, pkgName(p.dest)})
		sortEdits(es2)
		body = applyEdits(f.src, 0, len(f.src), es2)
	} else {
		// Header: for the main piece, everything up to the end of the imports
		// (comments, package clause, imports); for the others, a package clause.
		importsEnd := nameEnd
		for _, d := range f.ast.Decls {
			if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
				importsEnd = f.tf.Offset(g.End())
			}
		}
		if p.main {
			hdr.Write(f.src[:pkgStart])
			hdr.WriteString("package " + pkgName(p.dest))
			hdr.Write(f.src[nameEnd:importsEnd])
		} else {
			fmt.Fprintf(&hdr, "package %s\n", pkgName(p.dest))
			// Keep the imports; unused ones are pruned below.
			hdr.WriteString("\n")
			for _, d := range f.ast.Decls {
				if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
					hdr.Write(f.src[f.tf.Offset(g.Pos()):f.tf.Offset(g.End())])
					hdr.WriteString("\n")
				}
			}
		}
		for _, u := range p.units {
			hdr.Write(applyEdits(f.src, u.start, u.end, es))
		}
		hdr.WriteString("\n")
		body = hdr.Bytes()
	}
	var need []string
	for d := range p.foreign {
		need = append(need, d)
	}
	sort.Strings(need)
	var add []importSpec
	for _, d := range need {
		name := ""
		// goimports names an import whose package name is not its last path
		// element (cardsb ".../cards/b").
		if p.alias[d] != pkgName(d) || !strings.HasSuffix(importPath(d), "/"+pkgName(d)) {
			name = p.alias[d]
		}
		add = append(add, importSpec{name: name, path: importPath(d)})
	}
	out, err := fixImports(body, add, s.pkgNames)
	if err != nil {
		return nil, fmt.Errorf("%s → %s: %v", f.name, pkgDir(p.dest), err)
	}
	return out, nil
}

type importSpec struct{ name, path string }

// fixImports adds specs to a Go source's imports, drops imports nothing
// uses any more, and formats it. Usage is read syntactically: a selector
// whose left side is the import's name.
func fixImports(src []byte, add []importSpec, names map[string]string) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parseSrc(fset, src)
	if err != nil {
		return nil, fmt.Errorf("%v\n%s", err, src)
	}
	tf := fset.File(f.Pos())
	used := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if se, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := se.X.(*ast.Ident); ok {
				used[id.Name] = true
			}
		}
		return true
	})
	localName := func(im *ast.ImportSpec) string {
		path, _ := strconv.Unquote(im.Path.Value)
		if im.Name != nil {
			return im.Name.Name
		}
		if n, ok := names[path]; ok {
			return n
		}
		return path[strings.LastIndex(path, "/")+1:]
	}
	have := map[string]bool{}
	var es []edit
	var lastDecl *ast.GenDecl  // the last import decl that keeps a spec
	var anchor *ast.ImportSpec // the last kept non-standard-library spec in it
	isStd := func(path string) bool {
		first, _, _ := strings.Cut(path, "/")
		return !strings.Contains(first, ".")
	}
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.IMPORT {
			continue
		}
		var drop []*ast.ImportSpec
		var keptNonStd *ast.ImportSpec
		for _, sp := range g.Specs {
			im := sp.(*ast.ImportSpec)
			path, _ := strconv.Unquote(im.Path.Value)
			ln := localName(im)
			if ln == "_" || ln == "." || used[ln] {
				have[path] = true
				if !isStd(path) {
					keptNonStd = im
				}
				continue
			}
			drop = append(drop, im)
		}
		if len(drop) == len(g.Specs) {
			// Nothing left: remove the whole declaration and its line.
			start, end := tf.Offset(g.Pos()), tf.Offset(g.End())
			if end < len(src) && src[end] == '\n' {
				end++
			}
			es = append(es, edit{start, end - start, ""})
			continue
		}
		lastDecl, anchor = g, keptNonStd
		for _, im := range drop {
			start, end := tf.Offset(im.Pos()), tf.Offset(im.End())
			if im.Doc != nil {
				start = tf.Offset(im.Doc.Pos())
			}
			for start > 0 && (src[start-1] == '\t' || src[start-1] == ' ') {
				start--
			}
			if end < len(src) && src[end] == '\n' {
				end++
			}
			es = append(es, edit{start, end - start, ""})
		}
	}
	var lines []string
	for _, a := range add {
		if have[a.path] {
			continue
		}
		if a.name != "" {
			lines = append(lines, fmt.Sprintf("%s %q", a.name, a.path))
		} else {
			lines = append(lines, fmt.Sprintf("%q", a.path))
		}
	}
	if len(lines) > 0 {
		block := strings.Join(lines, "\n\t")
		switch {
		case lastDecl == nil:
			at := tf.Offset(f.Name.End())
			es = append(es, edit{at, 0, "\n\nimport (\n\t" + block + "\n)"})
		case lastDecl.Lparen.IsValid() && anchor != nil:
			// Into the group of the last module import (goimports keeps the
			// standard library in a group of its own).
			es = append(es, edit{tf.Offset(anchor.End()), 0, "\n\t" + block})
		case lastDecl.Lparen.IsValid():
			// Only the standard library so far: a new group after it.
			es = append(es, edit{tf.Offset(lastDecl.Rparen), 0, "\n\t" + block + "\n"})
		default:
			// import "x" → a block with both.
			sp := lastDecl.Specs[0].(*ast.ImportSpec)
			old := string(src[tf.Offset(sp.Pos()):tf.Offset(sp.End())])
			es = append(es, edit{tf.Offset(lastDecl.Pos()), tf.Offset(lastDecl.End()) - tf.Offset(lastDecl.Pos()),
				"import (\n\t" + old + "\n\n\t" + block + "\n)"})
		}
	}
	sortEdits(es)
	out := applyEdits(src, 0, len(src), es)
	b, err := format.Source(out)
	if err != nil {
		return nil, fmt.Errorf("gofmt: %v\n%s", err, out)
	}
	return b, nil
}

func parseSrc(fset *token.FileSet, src []byte) (*ast.File, error) {
	return parser.ParseFile(fset, "x.go", src, parser.ParseComments|parser.SkipObjectResolution)
}

// indexPkgNames records the package name of every import of the file set.
func (s *splitter) indexPkgNames() {
	s.pkgNames = map[string]string{}
	for _, f := range s.files {
		for _, im := range f.ast.Imports {
			if pn, ok := s.info.Implicits[im].(*types.PkgName); ok {
				s.pkgNames[pn.Imported().Path()] = pn.Imported().Name()
			} else if pn, ok := s.info.Defs[im.Name].(*types.PkgName); ok && im.Name != nil {
				s.pkgNames[pn.Imported().Path()] = pn.Imported().Name()
			}
		}
	}
}

// tstRef is a call in moved test code that becomes a call of an effectstest
// function (the flat view of the effects tree, flat_fs.go).
type tstRef struct {
	p      *piece
	off, n int
	name   string
	pos    token.Pos
}

// flatCalls are the file-system calls a source-scanning test makes on its
// own directory, and their effectstest replacements.
var flatCalls = map[string]map[string]string{
	"os":            {"ReadDir": "FlatReadDir", "ReadFile": "FlatReadFile", "Open": "FlatOpen", "Stat": "FlatStat"},
	"path/filepath": {"Glob": "FlatGlob"},
	"go/parser":     {"ParseFile": "FlatParseFile"},
}

// testRewrites edits test code for its new directory, two levels below the
// old one: relative paths gain "../../", and a test that scans the package's
// own sources (it imports go/ast or go/parser) reads them through the flat
// view in effectstest, with string literals that name a renamed identifier
// renamed to match.
func (s *splitter) testRewrites(ps []*piece, r *renamer, edits map[*srcFile][]edit, pieceOf map[*unit]*piece) []tstRef {
	lit := map[string]string{}
	ambiguous := map[string]bool{}
	for o, nn := range r.newName {
		if o.Name() == nn {
			continue
		}
		if prev, ok := lit[o.Name()]; ok && prev != nn {
			ambiguous[o.Name()] = true
		}
		lit[o.Name()] = nn
	}
	for n := range ambiguous {
		delete(lit, n)
	}
	var refs []tstRef
	for _, f := range s.files {
		if !isTestName(f.name) {
			continue
		}
		// astScanner: the file parses Go source. fsScanner: it also lists or
		// reads files by bare name, as a scan of its own directory does.
		astScanner, fsScanner := false, false
		for _, im := range f.ast.Imports {
			if p, _ := strconv.Unquote(im.Path.Value); p == "go/ast" || p == "go/parser" {
				astScanner = true
			}
		}
		ast.Inspect(f.ast, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok && len(c.Args) > 0 {
				if se, ok := c.Fun.(*ast.SelectorExpr); ok && (se.Sel.Name == "Glob" || se.Sel.Name == "ReadDir") {
					if bl, ok := c.Args[0].(*ast.BasicLit); ok && bl.Kind == token.STRING && !strings.Contains(bl.Value, "/") {
						fsScanner = true
					}
				}
			}
			return true
		})
		scanner := astScanner || fsScanner
		// Elements of filepath.Join after the first are not paths of their own.
		joinTail := map[*ast.BasicLit]bool{}
		// Literals compared with an AST node's name (x.Name, x.Sel.Name).
		nameCompared := map[*ast.BasicLit]bool{}
		isName := func(e ast.Expr) bool {
			se, ok := e.(*ast.SelectorExpr)
			if !ok || se.Sel.Name != "Name" {
				return false
			}
			// The Name of a *go/ast.Ident, not a card's name.
			return types.TypeString(s.info.Types[se.X].Type, nil) == "*go/ast.Ident"
		}
		ast.Inspect(f.ast, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.BinaryExpr:
				if n.Op == token.EQL || n.Op == token.NEQ {
					if bl, ok := n.Y.(*ast.BasicLit); ok && isName(n.X) {
						nameCompared[bl] = true
					}
					if bl, ok := n.X.(*ast.BasicLit); ok && isName(n.Y) {
						nameCompared[bl] = true
					}
				}
			case *ast.SwitchStmt:
				if n.Tag != nil && isName(n.Tag) {
					for _, c := range n.Body.List {
						for _, e := range c.(*ast.CaseClause).List {
							if bl, ok := e.(*ast.BasicLit); ok {
								nameCompared[bl] = true
							}
						}
					}
				}
			}
			return true
		})
		requote := func(old, v string) string {
			if strings.HasPrefix(old, "`") && !strings.Contains(v, "`") {
				return "`" + v + "`"
			}
			return strconv.Quote(v)
		}
		ast.Inspect(f.ast, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.BasicLit:
				if n.Kind != token.STRING {
					return true
				}
				v, err := strconv.Unquote(n.Value)
				if err != nil {
					return true
				}
				switch {
				case joinTail[n]:
				case v == "testdata" || strings.HasPrefix(v, "testdata/") || strings.HasPrefix(v, "../"):
					edits[f] = append(edits[f], edit{f.tf.Offset(n.Pos()), len(n.Value), requote(n.Value, "../../"+v)})
					s.pathRewrites++
				case scanner && lit[v] != "" && nameCompared[n]:
					// Compared with an AST name: fn.Name.Name != "footprint".
					edits[f] = append(edits[f], edit{f.tf.Offset(n.Pos()), len(n.Value), requote(n.Value, lit[v])})
					s.litRenames++
					logf("  %s: scanner literal %q → %q", f.name, v, lit[v])
				case scanner:
					// camelCase identifiers inside the literal ("drawThenDiscard",
					// a vocabulary regexp, a fixture's source) follow their rename;
					// a plain word ("flying") never does.
					nv := identToken.ReplaceAllStringFunc(v, func(w string) string {
						if nn := lit[w]; nn != "" && camelCase(w) {
							return nn
						}
						return w
					})
					if nv != v {
						edits[f] = append(edits[f], edit{f.tf.Offset(n.Pos()), len(n.Value), requote(n.Value, nv)})
						s.litRenames++
						logf("  %s: scanner literal %q → %q", f.name, v, nv)
					}
				}
			case *ast.CallExpr:
				se, ok := n.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				x, ok := se.X.(*ast.Ident)
				if !ok {
					return true
				}
				pn, ok := s.info.Uses[x].(*types.PkgName)
				if !ok {
					return true
				}
				path := pn.Imported().Path()
				if path == "path/filepath" && se.Sel.Name == "Join" {
					for _, a := range n.Args[min(1, len(n.Args)):] {
						if bl, ok := a.(*ast.BasicLit); ok {
							joinTail[bl] = true
						}
					}
				}
				if path == "path/filepath" && se.Sel.Name == "Join" && len(n.Args) > 0 {
					if bl, ok := n.Args[0].(*ast.BasicLit); ok && bl.Value == `".."` {
						edits[f] = append(edits[f], edit{f.tf.Offset(bl.Pos()), 0, `"..", "..", `})
						s.pathRewrites++
					}
				}
				if !scanner {
					return true
				}
				if repl, ok := flatCalls[path][se.Sel.Name]; ok {
					u := s.unitOf(n.Pos())
					if u == nil {
						return true
					}
					refs = append(refs, tstRef{p: pieceOf[u], off: f.tf.Offset(se.Pos()), n: f.tf.Offset(se.End()) - f.tf.Offset(se.Pos()), name: repl, pos: n.Pos()})
				}
			}
			return true
		})
	}
	return refs
}

// camelCase: an identifier with an upper-case letter after its first, so it
// cannot be an ordinary English word in a test's expected text.
func camelCase(v string) bool {
	if v == "" || !token.IsIdentifier(v) {
		return false
	}
	return strings.ToLower(v[1:]) != v[1:]
}

var identToken = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

var fileLine = regexp.MustCompile(`^([A-Za-z0-9_]+\.go):([0-9]+)$`)

// lineRemaps rewrites "<file>.go:<line>" literals in source-scanning tests
// (allowlists keyed by where a finding is) for the lines the split added to
// the top of that file (its imports).
func (s *splitter) lineRemaps(deltas map[string]int, edits map[*srcFile][]edit) {
	for _, f := range s.files {
		if !isTestName(f.name) {
			continue
		}
		ast.Inspect(f.ast, func(n ast.Node) bool {
			bl, ok := n.(*ast.BasicLit)
			if !ok || bl.Kind != token.STRING {
				return true
			}
			v, err := strconv.Unquote(bl.Value)
			if err != nil {
				return true
			}
			m := fileLine.FindStringSubmatch(v)
			if m == nil || deltas[m[1]] == 0 {
				return true
			}
			line, _ := strconv.Atoi(m[2])
			nv := fmt.Sprintf("%s:%d", m[1], line+deltas[m[1]])
			edits[f] = append(edits[f], edit{f.tf.Offset(bl.Pos()), len(bl.Value), strconv.Quote(nv)})
			s.lineRemapCount++
			return true
		})
	}
}
