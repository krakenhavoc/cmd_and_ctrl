package game

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// snapshot_lock_guard_test.go is #877's lint, and it is
// life_continuation_guard_test.go's sibling: a source scan for a shape
// that compiles, passes every test run on an idle table, and wedges
// the process the moment a writer shows up.
//
// The shape is a LOCK-TAKING accessor called from inside a snapshot
// body:
//
//	g.ReadSnapshot(func() {
//	        if p := g.PlayerByID(seat); p != nil { … }   // ← deadlock
//	})
//
// ReadSnapshot holds g.mu in read mode for the whole callback and
// PlayerByID takes it again. sync.RWMutex is not reentrant in either
// mode: a second RLock from the same goroutine blocks as soon as a
// writer is queued between the two, because Go blocks a new reader
// behind a waiting Lock so writers cannot starve. A bot runner
// dispatching a move is exactly that writer, which is why #848's
// aiseat flake read as a slow machine rather than as a deadlock, and
// why #876 converted six sites in the aiseat tests to
// PlayerByIDForEffect. WithWriteLock is worse and simpler: g.mu is
// already held for writing, so ANY second acquisition hangs
// unconditionally.
//
// The fix at a call site is always one of two things: the lock-free
// *ForEffect accessor, or a plain read off the fields the body can
// already see.
//
// ONE SCANNER, THE ACCESSORS AS DATA. The dangerous set is not a hand
// list that can rot — it is DERIVED from this package's own sources:
// every exported method on *Game whose body takes g.mu. Add a locking
// accessor to game.go tomorrow and it is covered today, which is the
// property the hand list would have lost on its first rename.
//
// FALSE POSITIVES ARE EXPECTED AND CHEAP. The scan has no type
// information, so it matches on the receiver's spelling (`g`, `game`,
// anything ending in `.Game`) and on the method name. A same-named
// method on some other value read through a field called Game would be
// flagged; add the site to the allowlist with the reason.

// snapshotBodies are the methods whose func-literal argument runs with
// g.mu already held. Both modes are listed because both deadlock: the
// read one only once a writer is queued, the write one always.
var snapshotBodies = map[string]bool{
	"ReadSnapshot":  true,
	"WithWriteLock": true,
}

// snapshotLockAllowlist exempts one "<file>:<line>" call, keyed by the
// path as the walk prints it, with the reason it is not the bug. An
// unexplained entry is how a lint stops meaning anything.
var snapshotLockAllowlist = map[string]string{}

func TestNoLockTakingAccessorInsideASnapshotBody(t *testing.T) {
	accessors := lockTakingGameAccessors(t)
	if len(accessors) < 20 {
		t.Fatalf("derived only %d locking accessors from this package — the derivation is broken, "+
			"and a lint that knows no accessors passes everything", len(accessors))
	}
	findings, scanned := scanTreeForLocksInSnapshots(t, "../..", accessors)
	if scanned == 0 {
		t.Fatal("scanned no files — the walk is wrong, and a lint that reads nothing passes everything")
	}
	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Errorf("a lock-taking accessor is called inside a snapshot body, which deadlocks the moment a "+
		"writer is queued between the two acquisitions (#877).\nUse the lock-free *ForEffect accessor, "+
		"or read the fields the body can already see; if this call is fine, add it to "+
		"snapshotLockAllowlist with the reason.\n  %s", strings.Join(findings, "\n  "))
}

// lockTakingGameAccessors parses this package and returns the exported
// methods on *Game whose body acquires g.mu in either mode.
func lockTakingGameAccessors(t *testing.T) map[string]bool {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	out := map[string]bool{}
	fset := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !gameMethod(fn) || !ast.IsExported(fn.Name.Name) {
				continue
			}
			if takesTheGameLock(fn.Body) {
				out[fn.Name.Name] = true
			}
		}
	}
	return out
}

// gameMethod reports whether fn is a method on *Game.
func gameMethod(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return false
	}
	star, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	ident, ok := star.X.(*ast.Ident)
	return ok && ident.Name == "Game"
}

// takesTheGameLock reports whether body calls g.mu.Lock / g.mu.RLock
// directly. A method that reaches the lock through a helper is not
// counted: the helper is the one with the lock, and it is the one the
// derivation finds.
func takesTheGameLock(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || (sel.Sel.Name != "Lock" && sel.Sel.Name != "RLock") {
			return true
		}
		if inner, ok := sel.X.(*ast.SelectorExpr); ok && inner.Sel.Name == "mu" {
			found = true
		}
		return true
	})
	return found
}

// scanTreeForLocksInSnapshots walks root for Go sources — tests
// included, since both known sites were tests — and reports every
// accessor call that sits inside a snapshot body, plus how many files
// it read.
func scanTreeForLocksInSnapshots(t *testing.T, root string, accessors map[string]bool) ([]string, int) {
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
			case ".git", "testdata", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			t.Fatalf("parse %s: %v", path, perr)
		}
		scanned++
		findings = append(findings, locksInSnapshotBodies(fset, path, file, accessors)...)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return findings, scanned
}

// locksInSnapshotBodies reports every accessor call inside a func
// literal handed to one of the snapshot bodies.
func locksInSnapshotBodies(fset *token.FileSet, file string, f *ast.File, accessors map[string]bool) []string {
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !snapshotBodies[sel.Sel.Name] {
			return true
		}
		for _, arg := range call.Args {
			lit, ok := arg.(*ast.FuncLit)
			if !ok {
				continue
			}
			out = append(out, accessorCallsIn(fset, file, lit.Body, accessors)...)
		}
		return true
	})
	return out
}

// accessorCallsIn reports every `<game>.Accessor(…)` call in body.
func accessorCallsIn(fset *token.FileSet, file string, body *ast.BlockStmt, accessors map[string]bool) []string {
	var out []string
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !accessors[sel.Sel.Name] || !looksLikeAGame(sel.X) {
			return true
		}
		where := filepath.ToSlash(file) + ":" + strconv.Itoa(fset.Position(sel.Pos()).Line)
		if _, ok := snapshotLockAllowlist[where]; ok {
			return true
		}
		out = append(out, where+" calls "+sel.Sel.Name+" under a held g.mu")
		return true
	})
	return out
}

// TestSnapshotLockGuardCatchesTheShape proves the lint above is a lint
// and not a function that returns nil. A tree with no findings is the
// expected state, which means the scan could rot into a no-op (a wrong
// walk, a derivation that finds nothing, a receiver rule that matches
// nothing) and nothing would say so.
//
// The fixtures are the two ws helpers as they were written before this
// change, and their fixed forms.
func TestSnapshotLockGuardCatchesTheShape(t *testing.T) {
	accessors := map[string]bool{"PlayerByID": true, "ReadSnapshot": true, "DrawCard": true}
	const before = `package ws

func lifeIn(g *game.Game, seat uuid.UUID) int {
	var life int
	g.ReadSnapshot(func() {
		if p := g.PlayerByID(seat); p != nil {
			life = p.Life
		}
	})
	return life
}
`
	const after = `package ws

func lifeIn(g *game.Game, seat uuid.UUID) int {
	var life int
	g.ReadSnapshot(func() {
		if p := g.PlayerByIDForEffect(seat); p != nil {
			life = p.Life
		}
	})
	return life
}
`
	const outsideTheBody = `package ws

func lifeIn(g *game.Game, seat uuid.UUID) int {
	p := g.PlayerByID(seat)
	var life int
	g.ReadSnapshot(func() { life = p.Life })
	return life
}
`
	const writerInAWriteLock = `package ws

func drawFor(g *game.Game, seat uuid.UUID) {
	g.WithWriteLock(func() { _, _ = g.DrawCard(seat) })
}
`
	const nestedSnapshot = `package ws

func nested(g *game.Game) {
	g.ReadSnapshot(func() { g.ReadSnapshot(func() {}) })
}
`
	for _, tc := range []struct {
		name string
		src  string
		want int
	}{
		{"the recursive RLock #877 is about", before, 1},
		{"the lock-free accessor", after, 0},
		{"a locking accessor outside the body", outsideTheBody, 0},
		{"a writer called under the write lock", writerInAWriteLock, 1},
		{"a snapshot inside a snapshot", nestedSnapshot, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "fixture.go", tc.src, 0)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			got := len(locksInSnapshotBodies(fset, "fixture.go", file, accessors))
			if got != tc.want {
				t.Errorf("findings = %d, want %d — the lint has stopped seeing the shape it exists for", got, tc.want)
			}
		})
	}
}

// TestSnapshotLockGuardDerivesTheRealAccessors pins the other half a
// source scan can silently lose: the derivation. PlayerByID is the
// accessor #877 was filed about, so if it ever stops being found the
// lint is reading the wrong package.
func TestSnapshotLockGuardDerivesTheRealAccessors(t *testing.T) {
	accessors := lockTakingGameAccessors(t)
	for _, want := range []string{"PlayerByID", "ReadSnapshot", "WithWriteLock", "ChangePlayerLife"} {
		if !accessors[want] {
			t.Errorf("%s is not in the derived set of lock-taking accessors", want)
		}
	}
	if accessors["PlayerByIDForEffect"] {
		t.Error("PlayerByIDForEffect takes no lock and must not be flagged — it is the fix")
	}
}

// looksLikeAGame reports whether expr is spelled like a *Game: the
// bare receiver the engine uses everywhere, or a field reached as
// something.Game / something.game. Spelling is all a scan without type
// information has; the allowlist is the escape hatch.
func looksLikeAGame(expr ast.Expr) bool {
	var buf strings.Builder
	if err := printer.Fprint(&buf, token.NewFileSet(), expr); err != nil {
		return false
	}
	text := buf.String()
	switch text {
	case "g", "game", "gm":
		return true
	}
	return strings.HasSuffix(text, ".Game") || strings.HasSuffix(text, ".game")
}
