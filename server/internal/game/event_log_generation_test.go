package game

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRestoreFromMovesTheEventLogGeneration: an undo replaces the log
// with a different history, and the next emit regrows it under the
// Seq values the undone events carried — so the generation is the only
// thing a cache can tell the two logs apart by (#1401).
func TestRestoreFromMovesTheEventLogGeneration(t *testing.T) {
	g := NewGame()
	g.EmitEvent(Event{Kind: EventChangeLife, Amount: 1})
	pre := g.Clone()
	g.EmitEvent(Event{Kind: EventChangeLife, Amount: 2})
	before := g.EventLogGeneration()

	g.RestoreFrom(pre)
	if g.EventLogGeneration() == before {
		t.Fatal("RestoreFrom left the event-log generation where it was")
	}
	after := g.EventLogGeneration()
	g.EmitEvent(Event{Kind: EventChangeLife, Amount: 3})
	if g.EventLogGeneration() != after {
		t.Fatal("an emit moved the generation; appends must not")
	}
	if pre.EventLogGeneration() != 0 || g.Clone().EventLogGeneration() != 0 {
		t.Fatal("a clone carried the generation; a new *Game starts its own")
	}
}

// TestEventLogIsReplacedOnlyWhereTheGenerationMoves holds
// projection_cache.go's two facts against the source. Every assignment
// to a Game's Events field, and every write into one of its elements,
// must be one of:
//
//   - EmitEvent's append (the log grows);
//   - RestoreFrom's replacement, which bumps eventLogGen;
//   - a function that fills a Game it has just built (cloneLocked,
//     restoreGame) or a snapshot struct (captureSnapshotLocked) — a
//     new object whose cache is empty.
//
// Anything else fails with the function's name. The fix is to bump the
// generation where the log is replaced (and add the function here), or
// to show the Game being written is fresh.
func TestEventLogIsReplacedOnlyWhereTheGenerationMoves(t *testing.T) {
	allowed := map[string]string{
		"EmitEvent":             "appends",
		"RestoreFrom":           "bumps eventLogGen",
		"cloneLocked":           "fills a new *Game",
		"restoreGame":           "fills a new *Game",
		"captureSnapshotLocked": "fills a GameSnapshot",
	}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			bumps := false
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				var lhs []ast.Expr
				switch s := n.(type) {
				case *ast.AssignStmt:
					lhs = s.Lhs
				case *ast.IncDecStmt:
					if sel, ok := s.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "eventLogGen" && s.Tok == token.INC {
						bumps = true
					}
					lhs = []ast.Expr{s.X}
				default:
					return true
				}
				for _, l := range lhs {
					if !writesEvents(l) {
						continue
					}
					if _, ok := allowed[fn.Name.Name]; !ok {
						t.Errorf("%s: %s writes .Events; a replaced event log must bump eventLogGen (see projection_cache.go), or this function must be shown to fill a fresh Game",
							fset.Position(l.Pos()), fn.Name.Name)
					}
				}
				return true
			})
			if allowed[fn.Name.Name] == "bumps eventLogGen" && !bumps {
				t.Errorf("%s: %s replaces the event log and no longer bumps eventLogGen", fset.Position(fn.Pos()), fn.Name.Name)
			}
		}
	}
}

// writesEvents reports whether an assignment target is X.Events, or an
// element or field of one (X.Events[i], X.Events[i].Seq).
func writesEvents(e ast.Expr) bool {
	for {
		switch x := e.(type) {
		case *ast.SelectorExpr:
			if x.Sel.Name == "Events" {
				return true
			}
			e = x.X
		case *ast.IndexExpr:
			e = x.X
		case *ast.ParenExpr:
			e = x.X
		case *ast.StarExpr:
			e = x.X
		default:
			return false
		}
	}
}
