package effects

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// x_matters_guard_test.go is #810's lint, in the style of #806's
// (life_continuation_guard_test.go): a source scan of the catalog for
// a declaration that has to be there and that nothing else can
// derive.
//
// `Spec.XMatters` says "everything this card does scales with the
// announced X, so X=0 does nothing at all". One rule reads it — the
// legal-move enumerator's X rule (internal/legal/x.go) — and it uses
// it to stop offering a bot a move that cannot change the game, which
// is what let a table of bots activate Soothsaying for X=0 until the
// wall clock ran out.
//
// The flag is therefore only as good as its coverage, and coverage is
// exactly the thing a reviewer does not notice is missing: a new
// {X} card whose author forgets the line compiles, plays correctly
// for a human, and re-opens the bug for a bot. So the catalog's own
// source is the check. A Spec whose effects read `ctx.X()` (directly,
// or through a package helper that does) must declare `XMatters:
// true`, and a Spec that declares it must read X.
//
// THE EXEMPTION IS THE POINT OF THE ALLOWLIST. A card with an {X}
// cost and a fixed RIDER — something that happens whatever X is —
// does not scale wholly with X, and X=0 is a real move for it: The
// Goose Mother at X=0 is a 2/2 flier for two mana. Such a card leaves
// XMatters unset and takes an allowlist entry saying what the rider
// is. Eight of the catalog's forty-two X cards are that shape.

// xMattersAllowlist exempts one "<file> <card name>" Register(Spec{…})
// from the scan, with the reason. Keyed by name rather than by line so
// an edit further up the file cannot silently retarget an exemption.
// Keep every entry justified: an unexplained entry is how a lint stops
// meaning anything.
//
// The line between "scales with X" and "has a rider" is mechanical, so
// that an entry is a reading of the card rather than a judgement about
// how good X=0 is: does resolving the spell at X=0 leave ANYTHING
// behind? An instant or sorcery whose every clause counts X leaves
// nothing; a 0/0 creature that enters with X counters leaves nothing
// either, because CR 704.5f puts it straight in the graveyard. A
// permanent that survives at X=0 — a printed toughness of 1 or more,
// or a noncreature permanent — is a rider, and so is a fixed clause
// that happens whatever X is.
var xMattersAllowlist = map[string]string{
	"benevolent_hydra.go Benevolent Hydra":     "1/1 body: the Hydra survives at X=0 with its counter-boosting replacement and its {T} ability",
	"farmer_cotton.go Farmer Cotton":           "1/1 legendary body: the creature stays even when it brings no Halflings and no Food",
	"fated_firepower.go Fated Firepower":       "an enchantment with flash: the permanent stays on the battlefield at X=0, amplifying by the 0 fire counters it entered with",
	"pull_from_tomorrow.go Pull from Tomorrow": "\"then discard a card\" is fixed: X=0 draws nothing and still discards, so it is a bad play rather than a no-op",
	"spiteful_banditry.go Spiteful Banditry":   "the Treasure-on-death trigger never reads X: at X=0 the enchantment deals no damage and is still an engine",
	"springleaf_parade.go Springleaf Parade":   "the \"creature tokens you control have {T}: Add one mana of any color\" static never reads X",
	"the_goose_mother.go The Goose Mother":     "2/2 flying body and an attack trigger that never reads X",
	"voracious_hydra.go Voracious Hydra":       "0/1 body survives at X=0, and the enters-the-battlefield fight mode is X-independent",
}

// xMattersFinding is one disagreement between a Spec's source and its
// declaration.
type xMattersFinding struct {
	where    string
	name     string
	readsX   bool
	declared bool
}

func (f xMattersFinding) String() string {
	if f.readsX {
		return f.where + " " + f.name + " reads the announced X but does not declare XMatters: true"
	}
	return f.where + " " + f.name + " declares XMatters: true but nothing in it reads the announced X"
}

func TestEveryXReadingSpecDeclaresXMatters(t *testing.T) {
	fset := token.NewFileSet()
	files := parseCatalogSources(t, fset)
	if len(files) == 0 {
		t.Fatal("scanned no catalog files — the glob is wrong, and a lint that reads nothing passes everything")
	}

	xReaders := functionsThatReadX(files)
	var findings []string
	specs := 0
	for name, file := range files {
		for _, lit := range registeredSpecs(file) {
			specs++
			card := specName(lit)
			if _, ok := xMattersAllowlist[name+" "+card]; ok {
				continue
			}
			f := xMattersFinding{
				where:    name + ":" + strconv.Itoa(fset.Position(lit.Pos()).Line),
				name:     card,
				readsX:   subtreeReadsX(lit, xReaders),
				declared: specDeclaresXMatters(lit),
			}
			if f.readsX != f.declared {
				findings = append(findings, f.String())
			}
		}
	}
	if specs == 0 {
		t.Fatal("found no Register(Spec{…}) literals — the scan has rotted")
	}
	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Errorf("XMatters and the source disagree for %d of %d specs (#810).\n"+
		"A Spec whose effects read ctx.X() must declare XMatters: true, so the bot's legal "+
		"enumerator knows X=0 does nothing (internal/legal/x.go, CR 732.2a); a card with a fixed "+
		"rider that happens whatever X is should stay undeclared and take an xMattersAllowlist "+
		"entry naming the rider.\n  %s",
		len(findings), specs, strings.Join(findings, "\n  "))
}

// parseCatalogSources parses every non-test file in the catalog
// package, keyed by file name.
func parseCatalogSources(t *testing.T, fset *token.FileSet) map[string]*ast.File {
	t.Helper()
	names, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	out := make(map[string]*ast.File, len(names))
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		file, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		out[name] = file
	}
	return out
}

// functionsThatReadX names every package-level function whose body
// reaches the announced X, directly or by calling another one that
// does. The closure matters because the batch helper files are where
// half the catalog's X reads actually live: a card file that only
// says `b35DreadSummons(ctx, item)` reads X just as surely as one
// that writes `ctx.X()` itself.
//
// Methods are not indexed — `Context.X` is the accessor the direct
// pattern already recognises, and indexing methods by their bare name
// would let any same-named method stand in for a package function.
func functionsThatReadX(files map[string]*ast.File) map[string]bool {
	bodies := map[string]*ast.FuncDecl{}
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || fn.Recv != nil {
				continue
			}
			bodies[fn.Name.Name] = fn
		}
	}
	readers := map[string]bool{}
	for name, fn := range bodies {
		if readsXDirectly(fn.Body) {
			readers[name] = true
		}
	}
	// Fixed point: a caller of a reader is a reader. The catalog is a
	// few hundred functions, so the naive loop is instant and there is
	// no ordering to get wrong.
	for changed := true; changed; {
		changed = false
		for name, fn := range bodies {
			if readers[name] {
				continue
			}
			if callsAny(fn.Body, readers) {
				readers[name] = true
				changed = true
			}
		}
	}
	return readers
}

// readsXDirectly reports whether the subtree reads the announced X in
// either spelling the catalog uses: `ctx.X()` on an effect Context, or
// `item.XValue` off the stack item.
func readsXDirectly(n ast.Node) bool {
	found := false
	ast.Inspect(n, func(n ast.Node) bool {
		if found {
			return false
		}
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == "XValue" {
			found = true
			return false
		}
		// ctx.X() — the selector alone is not enough, because a field
		// called X on a point-like struct would match too.
		if sel.Sel.Name == "X" {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "ctx" {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

// callsAny reports whether the subtree calls one of `names` by its
// bare package-level name.
func callsAny(n ast.Node, names map[string]bool) bool {
	found := false
	ast.Inspect(n, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ident, ok := call.Fun.(*ast.Ident); ok && names[ident.Name] {
			found = true
			return false
		}
		return true
	})
	return found
}

// subtreeReadsX is the per-Spec question: does anything this literal
// hangs off the card reach the announced X?
func subtreeReadsX(lit *ast.CompositeLit, xReaders map[string]bool) bool {
	return readsXDirectly(lit) || callsAny(lit, xReaders)
}

// registeredSpecs returns every `Register(Spec{…})` literal in a file.
// A cycle that registers in a loop still builds its Spec as a literal,
// so the scan sees each one once however the call is wrapped.
func registeredSpecs(file *ast.File) []*ast.CompositeLit {
	var out []*ast.CompositeLit
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if ident, ok := lit.Type.(*ast.Ident); ok && ident.Name == "Spec" {
			out = append(out, lit)
		}
		return true
	})
	return out
}

// specDeclaresXMatters reports whether the literal sets XMatters to a
// literal true.
func specDeclaresXMatters(lit *ast.CompositeLit) bool {
	for _, el := range lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "XMatters" {
			continue
		}
		val, ok := kv.Value.(*ast.Ident)
		return ok && val.Name == "true"
	}
	return false
}

// specName pulls the Name field out of a Spec literal for the failure
// message; the file and line alone would make a reader go looking.
func specName(lit *ast.CompositeLit) string {
	for _, el := range lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if key, ok := kv.Key.(*ast.Ident); !ok || key.Name != "Name" {
			continue
		}
		if s, ok := kv.Value.(*ast.BasicLit); ok && s.Kind == token.STRING {
			if unquoted, err := strconv.Unquote(s.Value); err == nil {
				return unquoted
			}
		}
	}
	return "(unnamed Spec)"
}

// TestXMattersGuardCatchesTheShape proves the lint above is a lint and
// not a function that returns nil. The catalog passing it is the
// expected state, which means the scan could rot into a no-op — a
// wrong glob, a renamed accessor — and nothing would say so.
func TestXMattersGuardCatchesTheShape(t *testing.T) {
	const src = `package effects

func helper(ctx *Context) int { return ctx.X() }

func init() {
	Register(Spec{Name: "Undeclared Direct", OnResolve: func(item *game.StackItem, ctx *Context) error {
		return DrawCards{N: ctx.X()}.Apply(ctx)
	}})
	Register(Spec{Name: "Undeclared Through A Helper", OnResolve: func(item *game.StackItem, ctx *Context) error {
		return DrawCards{N: helper(ctx)}.Apply(ctx)
	}})
	Register(Spec{Name: "Declared And True", XMatters: true, OnResolve: func(item *game.StackItem, ctx *Context) error {
		return DrawCards{N: ctx.X()}.Apply(ctx)
	}})
	Register(Spec{Name: "Declared And False", XMatters: true})
	Register(Spec{Name: "Neither"})
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fake.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	files := map[string]*ast.File{"fake.go": file}
	xReaders := functionsThatReadX(files)
	if !xReaders["helper"] {
		t.Error("the call graph missed a direct ctx.X() reader")
	}
	want := map[string]bool{
		"Undeclared Direct":           true,
		"Undeclared Through A Helper": true,
		"Declared And False":          true,
	}
	got := map[string]bool{}
	for _, lit := range registeredSpecs(file) {
		if subtreeReadsX(lit, xReaders) != specDeclaresXMatters(lit) {
			got[specName(lit)] = true
		}
	}
	if len(got) != len(want) {
		t.Fatalf("findings = %v, want %v", got, want)
	}
	for name := range want {
		if !got[name] {
			t.Errorf("the guard missed %q", name)
		}
	}
}
