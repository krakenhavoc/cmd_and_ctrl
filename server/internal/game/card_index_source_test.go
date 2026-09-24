package game

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// card_index_source_test.go holds card_index.go's assumptions against
// the source (#1479, ADR 0094), the way event_log_generation_test.go
// holds the public-log cache's (#1401).
//
// The card index never has to be invalidated, because every answer it
// gives is checked against the live zone first. What it does rely on
// is set out at the top of card_index.go. This file pins what a
// program can see of it: hints hold no pointers, only the checked
// lookup reads them, and every write to a Cards slice outside the Zone
// primitives is listed with the reason it keeps one instance ID in one
// zone. The dynamic net under that last one is the cross-check mode
// (SetCardIndexCrossCheck), on for this package's whole run.

// TestCardIndexHintsHoldNoPointers: a hint must name its zone by slot
// and seat index, never by pointer. RestoreFrom swaps every zone
// pointer on the live game, so a hint holding a *Zone would outlive an
// undo still pointing at the discarded zone — and that zone still holds
// the card, so the check would pass and the answer would be the wrong
// object.
func TestCardIndexHintsHoldNoPointers(t *testing.T) {
	var loc cardLoc
	typ := reflect.TypeOf(loc)
	for i := 0; i < typ.NumField(); i++ {
		switch f := typ.Field(i); f.Type.Kind() {
		case reflect.Pointer, reflect.UnsafePointer, reflect.Map, reflect.Slice,
			reflect.Interface, reflect.Func, reflect.Chan, reflect.String:
			t.Errorf("cardLoc.%s is a %s; a hint must be plain values resolved through the live Game", f.Name, f.Type.Kind())
		}
	}
	if got := reflect.TypeOf(cardIndexTable{}.locs).Elem(); got != typ {
		t.Errorf("the table's values are %s, want cardLoc", got)
	}
}

// TestCardIndexIsReadOnlyThroughItsCheckedLookup: nothing outside
// card_index.go may touch the index, and inside it the table's
// entries are read in exactly one place — locateIndexedLocked, which
// checks every entry against the live zone before using it. A second
// reader could return a hint unchecked.
func TestCardIndexIsReadOnlyThroughItsCheckedLookup(t *testing.T) {
	for _, pf := range parseGamePackage(t) {
		for _, decl := range pf.file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				switch sel.Sel.Name {
				case "cardIndex":
					if pf.name != "card_index.go" {
						t.Errorf("%s: %s touches Game.cardIndex; go through locateCardLocked (card_index.go)",
							pf.fset.Position(sel.Pos()), fn.Name.Name)
					}
				case "locs":
					if fn.Name.Name != "locateIndexedLocked" && fn.Name.Name != "rebuildCardIndexLocked" {
						t.Errorf("%s: %s reads the card index's table; only locateIndexedLocked may, because it checks each entry against the live zone",
							pf.fset.Position(sel.Pos()), fn.Name.Name)
					}
				}
				return true
			})
		}
	}
}

// TestZoneCardsAreWrittenOnlyAtKnownPoints lists every place in the
// package that writes a Cards slice, or a whole element of one, or a
// card's InstanceID, outside Zone's own methods and MoveCard.
//
// The index is safe against any such write — a stale hint fails its
// check and falls back to the walk. What it is not safe against is a
// write that leaves one instance ID in TWO zones, because then the
// index and the walk can each find a different one. So a new writer
// fails here until it is added with a reason that says why it keeps
// each instance ID in one zone. (Syntactic: some of the listed writes
// are to other types' Cards fields, and say so.)
func TestZoneCardsAreWrittenOnlyAtKnownPoints(t *testing.T) {
	// The Zone primitives: every add is paired with the caller's
	// remove, or is a remove, or reorders in place.
	primitives := map[string]bool{
		"PushTop": true, "PushBottom": true, "InsertFromTop": true,
		"PopTop": true, "Remove": true, "Shuffle": true, "MoveCard": true,
	}
	known := map[string]string{
		// Filling a zone that was just made — no live zone to
		// duplicate into.
		"cloneZone":   "fills the zone it has just made, from the zone it is copying",
		"restoreZone": "fills the zone it has just made, from a snapshot",
		// Moves that do their own remove-then-add, in one function.
		"phaseOutLocked": "battlefield → PhasedOut: spliced out of one slice, then appended to the other",
		"phaseInLocked":  "PhasedOut → battlefield: spliced out of one slice, then appended to the other",
		// Writes that only take cards out.
		"ReplaceDeck":                "lobby only: empties the seat's library and command zone before the new deck goes in",
		"Mulligan":                   "empties the hand whose cards it has just pushed onto the library, one by one",
		"removeObjectsOwnedByLocked": "CR 800.4a: filters a leaving player's cards out of each zone",
	}
	// Other types in the package have a Cards field too. The check is
	// syntactic, so they are listed here by function, with the type.
	notZones := map[string]string{
		"snapshotZone":                        "zoneSnapshot.Cards",
		"QueueReselectAttackForEffect":        "an option's Cards []uuid.UUID",
		"GrantCastPermissionToCardsForEffect": "CastPermission.Cards",
		"cloneCastPermissions":                "CastPermission.Cards",
		"ScheduleDelayedTriggerForEffect":     "DelayedTrigger.Cards",
		"cloneDelayedTrigger":                 "DelayedTrigger.Cards",
		"snapshotDelayedTrigger":              "a delayed trigger snapshot's Cards",
		"restoreDelayedTrigger":               "DelayedTrigger.Cards",
		"cloneChoiceOptions":                  "a choice option's Cards",
		"reassignChoiceLocked":                "a pick option's Cards",
		"withoutPicked":                       "a pick option's Cards",
		"retargetAlternatives":                "a retarget option's Cards",
		"specMatchesLocked":                   "a target match's Cards",
	}
	for name, why := range notZones {
		known[name] = "not a Zone: " + why
	}
	var found []string
	seen := map[string]bool{}
	for _, pf := range parseGamePackage(t) {
		for _, decl := range pf.file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || primitives[fn.Name.Name] {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				as, ok := n.(*ast.AssignStmt)
				if !ok {
					return true
				}
				for _, l := range as.Lhs {
					if !writesZoneCards(l) {
						continue
					}
					seen[fn.Name.Name] = true
					if _, ok := known[fn.Name.Name]; !ok {
						found = append(found, pf.fset.Position(l.Pos()).String()+" "+fn.Name.Name)
					}
				}
				return true
			})
		}
	}
	sort.Strings(found)
	for _, f := range found {
		t.Errorf("%s writes a Cards slice (or an element, or an InstanceID) outside the zone primitives. "+
			"If it keeps every instance ID in one zone, add it to `known` with the reason; "+
			"if the Cards it writes is not a Zone's, add it to `notZones`. See ADR 0094.", f)
	}
	for name := range known {
		if !seen[name] {
			t.Errorf("known writer %s no longer writes a Cards slice; drop it from the list", name)
		}
	}
}

// writesZoneCards reports whether an assignment target is X.Cards,
// X.Cards[i] (a whole element) or X.Cards[i].InstanceID. A field of an
// element other than the ID (X.Cards[i].Tapped) changes no card's
// zone and no card's identity, and is not reported.
func writesZoneCards(e ast.Expr) bool {
	if sel, ok := e.(*ast.SelectorExpr); ok {
		if sel.Sel.Name == "Cards" {
			return true
		}
		if sel.Sel.Name == "InstanceID" {
			if ix, ok := sel.X.(*ast.IndexExpr); ok {
				if inner, ok := ix.X.(*ast.SelectorExpr); ok && inner.Sel.Name == "Cards" {
					return true
				}
			}
		}
		return false
	}
	if ix, ok := e.(*ast.IndexExpr); ok {
		if sel, ok := ix.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "Cards" {
			return true
		}
	}
	return false
}

type parsedGameFile struct {
	name string
	fset *token.FileSet
	file *ast.File
}

func parseGamePackage(t *testing.T) []parsedGameFile {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var out []parsedGameFile
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
		out = append(out, parsedGameFile{name: name, fset: fset, file: f})
	}
	return out
}
