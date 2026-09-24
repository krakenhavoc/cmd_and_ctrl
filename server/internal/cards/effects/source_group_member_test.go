package effects

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// source_group_member_test.go — #1463, the "each" half of #1432. A
// primitive applied to the members of a set the text reads off the
// board ("each creature you control") goes through ctx.asGroupMember(),
// so the source's new object — a member of that set like any other
// (CR 400.7, CR 611.2c) — is acted on. ADR 0018 amendment 2026-09-24,
// Decision 20.

// --- the primitives, reached as "this" and as a group member ---------

// Every #1432 primitive, aimed at its own source four ways: live or
// flickered in response, reached as "this" or as a member of a group
// (ctx.asGroupMember()). Only "this" + flickered does nothing — the new
// object is a stranger to "this" but a member of "each creature you
// control" like any other. A zero Target (crew's BecomeCreatureUntilEOT{})
// means "this" by construction, so the group mark does not switch its
// question off.
func TestAPrimitiveReachedAsAGroupMemberActsOnTheSourcesNewObject(t *testing.T) {
	for _, group := range []bool{false, true} {
		for _, flicker := range []bool{false, true} {
			for _, tc := range primitiveCases(uuid.Nil) {
				name := tc.name + "/this"
				if group {
					name = tc.name + "/group"
				}
				if flicker {
					name += "/flickered"
				} else {
					name += "/live"
				}
				t.Run(name, func(t *testing.T) {
					g := newCatalogGame(t)
					me, opp := g.Seats[0], g.Seats[1]
					tc := tc
					for _, c := range primitiveCases(opp.ID) {
						if c.name == tc.name {
							tc = c
						}
					}
					typeLine := tc.typeLine
					if typeLine == "" {
						typeLine = "Creature — Bear"
					}
					src := pushDiesCreatureForTest(g, me.ID, "Relic Bear", "", typeLine, 2, 2)
					item := &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered,
						Controller: me.ID, Owner: me.ID, SourceCardID: src}
					g.WithWriteLock(func() { item.SourceObject, _ = g.PermanentRefForEffect(src) })
					if flicker {
						flickerInResponse(t, g, src)
					}
					if tc.setup != nil {
						tc.setup(t, g, src)
					}
					g.WithWriteLock(func() {
						ctx := NewContext(g, item)
						if group {
							ctx = ctx.asGroupMember()
						}
						if err := tc.apply(ctx, src); err != nil {
							t.Fatalf("Apply: %v", err)
						}
					})
					thisByConstruction := strings.Contains(tc.name, "zero target")
					want := !flicker || (group && !thisByConstruction)
					if got := tc.acted(t, g, src); got != want {
						t.Errorf("%s (group=%v, flicker=%v): acted = %v, want %v", tc.name, group, flicker, got, want)
					}
				})
			}
		}
	}
}

// The list primitives keep the source when the list is a group, and
// the mark lives on the copy only.
func TestWithoutNewSourceObjectKeepsAGroupMember(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushDiesCreatureForTest(g, me.ID, "Relic Bear", "", "Creature — Bear", 2, 2)
	other := pushDiesCreatureForTest(g, me.ID, "Other Bear", "", "Creature — Bear", 2, 2)
	item := &game.StackItem{ID: uuid.New(), Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID, SourceCardID: src}
	g.WithWriteLock(func() { item.SourceObject, _ = g.PermanentRefForEffect(src) })
	flickerInResponse(t, g, src)
	g.WithWriteLock(func() {
		ctx := NewContext(g, item)
		if got := ctx.asGroupMember().withoutNewSourceObject([]uuid.UUID{src, other}); len(got) != 2 {
			t.Errorf("group: got %v, want both bears", got)
		}
		if got := ctx.withoutNewSourceObject([]uuid.UUID{src, other}); len(got) != 1 || got[0] != other {
			t.Errorf("this: got %v, want only the other bear", got)
		}
		if !ctx.isNewSourceObject(src) {
			t.Error("asGroupMember marked the Context it was called on, not a copy")
		}
		// A site that names the source by construction still asks.
		if !ctx.asGroupMember().isNewSourceObjectAsThis(src) {
			t.Error("isNewSourceObjectAsThis was switched off by the group mark")
		}
		// "For as long as ~ remains on the battlefield" names the
		// source whatever Context builds it: a new object never begins
		// the duration (CR 611.2b), group mark or not.
		if _, ok := DurationWhileSourceRemains(ctx.asGroupMember(), src); ok {
			t.Error("a duration keyed on a source that came back began under a group-marked Context")
		}
	})
}

// --- proof cards -------------------------------------------------------

// Steel Overseer — "{T}: Put a +1/+1 counter on each artifact creature
// you control." The Overseer counts itself, and the Overseer that came
// back in response is still an artifact creature you control.
func TestSteelOverseerFlickeredInResponseStillCountsItself(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		overseer := pushCatalogPermanent(g, me.ID, "Steel Overseer", "Artifact Creature — Construct", b07SteelOverseerOracle, false)
		myr := pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Myr", TypeLine: "Artifact Creature — Myr",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
		if err := g.ActivateCatalogAbility(me.ID, overseer, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if flicker {
			flickerInResponse(t, g, overseer)
		}
		passPriorityAroundTable(t, g)
		for _, id := range []uuid.UUID{overseer, myr} {
			if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
				t.Errorf("flicker=%v: %v has %d +1/+1 counters, want 1", flicker, id, got)
			}
		}
	}
}

// Mazirek, Kraul Death Priest — a TRIGGERED "each creature you
// control". Mazirek bounced and replayed in response still gets his
// counter.
func TestMazirekFlickeredInResponseStillCountsHimself(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		mazirek := pushCatalogPermanent(g, me.ID, "Mazirek, Kraul Death Priest",
			"Legendary Creature — Insect Shaman", mazirekOracle, false)
		mine := pushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
		fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
		g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
		if flicker {
			flickerInResponse(t, g, mazirek)
		}
		passPriorityAroundTable(t, g)
		for _, id := range []uuid.UUID{mazirek, mine} {
			if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
				t.Errorf("flicker=%v: %v has %d +1/+1 counters, want 1", flicker, id, got)
			}
		}
	}
}

// Kalonian Hydra — "Whenever this creature attacks, double the number
// of +1/+1 counters on each creature you control." The Hydra that came
// back in response is no longer attacking, but it is a creature you
// control, so its counters double.
func TestKalonianHydraFlickeredInResponseStillDoublesItself(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		hydra := pushCatalogPermanent(g, me.ID, "Kalonian Hydra", "Creature — Hydra", b10KalonianHydraOracle, false)
		b10Awake(g, hydra)
		declareAttack(t, g, opp.ID, hydra)
		if flicker {
			flickerInResponse(t, g, hydra)
		}
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(hydra, game.CounterPlusOne, 2); err != nil {
				t.Fatal(err)
			}
		})
		passPriorityAroundTable(t, g)
		if got := counterCount(g, hydra, game.CounterPlusOne); got != 4 {
			t.Errorf("flicker=%v: Hydra has %d +1/+1 counters, want 4 (two, doubled)", flicker, got)
		}
	}
}

// Whitemane Lion — "When this creature enters, return a creature you
// control to its owner's hand." The creature is CHOSEN at resolution
// from the ones you control, so a Lion that left and came back in
// response is on offer like any other, and picking it returns it.
// Without the flicker the Lion returns itself as it always did.
func TestWhitemaneLionFlickeredInResponseCanStillReturnItself(t *testing.T) {
	for _, flicker := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		lion := castCatalogSpell(t, g, "Whitemane Lion", "Creature — Cat", b36WhitemaneLionOracle, nil)
		passUntilOnBattlefield(t, g, lion)
		if flicker {
			flickerInResponse(t, g, lion)
		}
		passPriorityAroundTable(t, g)
		p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
		if p == nil {
			t.Fatalf("flicker=%v: Whitemane Lion queued no return pick", flicker)
		}
		if !hasID(p.ChooseCards, lion) {
			t.Fatalf("flicker=%v: the Lion is not offered", flicker)
		}
		if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{lion}); err != nil {
			t.Fatalf("ResolveOwnPermanents: %v", err)
		}
		if g.Battlefield.Contains(lion) || !me.Hand.Contains(lion) {
			t.Errorf("flicker=%v: the Lion that was picked did not return to hand", flicker)
		}
	}
}

// --- the regression: no "each" loop may forget to say so ------------

// Every function that reads a set off the board and applies a
// per-permanent primitive (one that asks the #1432 question) inside a
// loop says asGroupMember. Best effort by construction, like its
// #1432 siblings: "reads a set off the board" is a call to
// BattlefieldCardsForEffect, a range over Battlefield.Cards, or a call
// to a package function that returns []uuid.UUID and does either. A
// loop that forgets is weaker than printed, never stronger, which is
// why the default stays guarded and this test exists.
func TestEveryEachLoopReachesItsMembersAsAGroup(t *testing.T) {
	// Functions that match the shape but whose loop is NOT over the
	// set the board read produced, keyed "file.go:function" (a FuncLit
	// is keyed by its enclosing FuncDecl).
	exempt := map[string]string{
		"rakdos_charm.go:init": "the loop exiles a graveyard (mode 1); the board read is mode 3's damage, which DealDamage does not ask about",
	}

	ps := scanPrimitives(t)
	perPermanent := map[string]bool{}
	for name := range ps.guarded {
		if st, ok := ps.structs[name]; ok && actsOnACard(st) {
			perPermanent[name] = true
		}
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	parsed := map[string]*ast.File{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		parsed[f] = file
	}

	// A board read, directly or through a package set reader.
	setReaders := map[string]bool{}
	readsBoard := func(body *ast.BlockStmt) bool {
		found := false
		inspectShallow(body, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				switch fn := x.Fun.(type) {
				case *ast.SelectorExpr:
					if fn.Sel.Name == "BattlefieldCardsForEffect" {
						found = true
					}
				case *ast.Ident:
					if setReaders[fn.Name] {
						found = true
					}
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "Cards" && strings.HasSuffix(exprString(x.X), "Battlefield") {
					found = true
				}
			}
			return !found
		})
		return found
	}
	for changed := true; changed; {
		changed = false
		for _, file := range parsed {
			for _, decl := range file.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok || fd.Recv != nil || fd.Body == nil || setReaders[fd.Name.Name] || fd.Type.Results == nil {
					continue
				}
				returnsIDs := false
				for _, r := range fd.Type.Results.List {
					if exprString(r.Type) == "[]uuid.UUID" {
						returnsIDs = true
					}
				}
				if returnsIDs && readsBoard(fd.Body) {
					setReaders[fd.Name.Name] = true
					changed = true
				}
			}
		}
	}

	// appliesPerPermanentInALoop: a range whose body applies a
	// per-permanent primitive literal.
	appliesPerPermanentInALoop := func(body *ast.BlockStmt) bool {
		found := false
		inspectShallow(body, func(n ast.Node) bool {
			rs, ok := n.(*ast.RangeStmt)
			if !ok {
				return !found
			}
			inspectShallow(rs.Body, func(m ast.Node) bool {
				call, ok := m.(*ast.CallExpr)
				if !ok {
					return !found
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Apply" {
					return !found
				}
				x := sel.X
				if p, ok := x.(*ast.ParenExpr); ok {
					x = p.X
				}
				if cl, ok := x.(*ast.CompositeLit); ok {
					if id, ok := cl.Type.(*ast.Ident); ok && perPermanent[id.Name] {
						found = true
					}
				}
				return !found
			})
			return !found
		})
		return found
	}
	says := func(body *ast.BlockStmt) bool {
		found := false
		inspectShallow(body, func(n ast.Node) bool {
			if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "asGroupMember" {
				found = true
			}
			return !found
		})
		return found
	}

	var bad []string
	seen := map[string]bool{}
	for f, file := range parsed {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			key := f + ":" + fd.Name.Name
			// The innermost function that holds the loop decides: a
			// card's Effect closure, or the helper it calls.
			var check func(body *ast.BlockStmt)
			check = func(body *ast.BlockStmt) {
				ast.Inspect(body, func(n ast.Node) bool {
					if lit, ok := n.(*ast.FuncLit); ok && lit.Body != body {
						check(lit.Body)
						return false
					}
					return true
				})
				if !appliesPerPermanentInALoop(body) || !readsBoard(body) || says(body) {
					return
				}
				// A loop nested in a closure was already judged there.
				seen[key] = true
				if _, ok := exempt[key]; ok {
					return
				}
				bad = append(bad, key+" ("+fset.Position(body.Pos()).String()+")")
			}
			check(fd.Body)
		}
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		t.Errorf("these functions read a set off the board and apply a per-permanent primitive to its members without ctx.asGroupMember() (#1463): the source's new object, a member of the set, is skipped (weaker than printed):\n  %s\n"+
			"Apply the loop's primitive with ctx.asGroupMember(), or add the function to the exemption table with the reason its loop is not over that set.",
			strings.Join(bad, "\n  "))
	}
	for key := range exempt {
		if !seen[key] {
			t.Errorf("exempt function %s no longer matches the shape — drop it from the table", key)
		}
	}
}

// inspectShallow is ast.Inspect that does not descend into a nested
// function literal: a closure's body is judged on its own.
func inspectShallow(root ast.Node, fn func(ast.Node) bool) {
	ast.Inspect(root, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok && n != root {
			return false
		}
		return fn(n)
	})
}
