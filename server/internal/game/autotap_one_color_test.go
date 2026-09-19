package game

import (
	"testing"

	"github.com/google/uuid"
)

// autotap_one_color_test.go — #779: the auto-tapper plans a source
// that adds "N mana of any one color" (Gilded Lotus, Lotus Field, Nyx
// Lotus, White Lotus Tile).
//
// Before this the planner skipped such a source outright (#742), so a
// board that could pay reported that it could not: strict-mode casts
// refused, the lobby's cast preview said `ok: false`, and no bot was
// ever offered the cast. That is the #273 failure shape — the engine
// saying "can't pay" about a board that can — one card family over.
//
// The rule the planner has to keep is the one the activation enforces:
// ONE pick for all N tokens. A Gilded Lotus may fund {U}{U}{U}; it may
// never fund {W}{U} on its own.

// pushOneColorSource seeds a permanent whose only mana ability is a
// one-pick-N-tokens "any one color" tap.
func pushOneColorSource(g *Game, owner *Player, name, produced string) uuid.UUID {
	return pushIntrinsicPermanent(g, owner, name, "Artifact", oneColorShape(produced), nil)
}

// pushIslandFor seeds a basic Island under `owner` (the synthetic
// basic-land mana ability comes from the type line).
func pushIslandFor(g *Game, owner *Player) uuid.UUID {
	return pushBattlefieldForTest(g, owner.ID, "Island", "Basic Land — Island", "")
}

// poolColors tallies the controller's pool by colour.
func poolColors(p *Player) map[string]int {
	out := map[string]int{}
	for _, tok := range p.ManaPool {
		out[tok.Color]++
	}
	return out
}

// Gilded Lotus plus two Islands pays {3}{U}{U}: the plan names all
// three permanents, and every token the Lotus minted is the same
// colour.
func TestAutoTapGildedLotusFundsAColouredCostWithTwoIslands(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lotus := pushOneColorSource(g, me, "Gilded Lotus", "{W3|U3|B3|R3|G3}")
	pushIslandFor(g, me)
	pushIslandFor(g, me)

	cost, err := ParseCost("{3}{U}{U}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok {
		t.Fatalf("no plan for {3}{U}{U} off Gilded Lotus + 2 Islands")
	}
	if len(plan) != 3 {
		t.Fatalf("plan = %v, want three permanents", plan)
	}
	found := false
	for _, id := range plan {
		if id == lotus {
			found = true
		}
	}
	if !found {
		t.Fatalf("plan = %v, want it to name the Gilded Lotus", plan)
	}

	// Execute it and check the pool really pays: five mana, of which
	// the Lotus's three share one colour.
	g.WithWriteLock(func() {
		full, ok := g.autoTapLocked(me.ID, cost, 0, nil)
		if !ok {
			t.Fatalf("autoTapLocked disagreed with AutoTapForCost")
		}
		g.materializePlanLocked(me, full, cost)
	})
	if got := len(me.ManaPool); got != 5 {
		t.Fatalf("pool = %v, want five mana", me.ManaPool)
	}
	fromLotus := map[string]int{}
	for _, tok := range me.ManaPool {
		if tok.Source == lotus {
			fromLotus[tok.Color]++
		}
	}
	if len(fromLotus) != 1 {
		t.Errorf("the Lotus minted %v, want three mana of ONE colour", fromLotus)
	}
	if !me.ManaPool.CanPayFor(cost, 0, ManaSpendContext{}) {
		t.Errorf("pool %v cannot pay {3}{U}{U} after the plan ran", me.ManaPool)
	}
}

// One pick cannot be two colours: a lone Gilded Lotus is never planned
// to cover {W}{U}.
func TestAutoTapGildedLotusNeverCoversTwoColours(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushOneColorSource(g, me, "Gilded Lotus", "{W3|U3|B3|R3|G3}")

	cost, err := ParseCost("{W}{U}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Errorf("plan = %v for {W}{U} off one Gilded Lotus — the pick is one colour", plan)
	}
	// It pays {W}{W} from the same board, because that IS one pick.
	same, err := ParseCost("{W}{W}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if _, ok := g.AutoTapForCost(me.ID, same, 0); !ok {
		t.Errorf("no plan for {W}{W} off one Gilded Lotus")
	}
}

// With a Plains on the board the {W}{U} cost is still unpayable — the
// Plains covers the {W} and the Lotus cannot make a second colour
// beside it... but {W}{U}{U} IS payable, Plains for the white and the
// Lotus booked blue.
func TestAutoTapGildedLotusPairsWithAnotherSourceForTheSecondColour(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lotus := pushOneColorSource(g, me, "Gilded Lotus", "{W3|U3|B3|R3|G3}")
	pushBattlefieldForTest(g, me.ID, "Plains", "Basic Land — Plains", "")

	cost, err := ParseCost("{W}{U}{U}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	g.WithWriteLock(func() {
		plan, ok := g.autoTapLocked(me.ID, cost, 0, nil)
		if !ok {
			t.Fatalf("no plan for {W}{U}{U} off Gilded Lotus + Plains")
		}
		g.materializePlanLocked(me, plan, cost)
	})
	got := poolColors(me)
	if got["U"] < 2 || got["W"] < 1 {
		t.Fatalf("pool = %v, want at least {W} and two {U}", got)
	}
	for _, tok := range me.ManaPool {
		if tok.Source == lotus && tok.Color != "U" {
			t.Errorf("the Lotus minted %q as well as blue — one pick, one colour", tok.Color)
		}
	}
}

// Nyx Lotus's amounts differ per colour: with devotion G4 / U1 it pays
// {G}{G}{G}{G} and cannot pay {U}{U}.
func TestAutoTapNyxLotusPicksTheColourWithEnoughDevotion(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lotus := pushOneColorSource(g, me, "Nyx Lotus", "{G4|U1}")

	green, err := ParseCost("{G}{G}{G}{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	g.WithWriteLock(func() {
		plan, ok := g.autoTapLocked(me.ID, green, 0, nil)
		if !ok {
			t.Fatalf("no plan for {G}{G}{G}{G} off a G4/U1 Nyx Lotus")
		}
		g.materializePlanLocked(me, plan, green)
	})
	if got := poolColors(me); got["G"] != 4 || len(got) != 1 {
		t.Errorf("pool = %v, want four green", got)
	}
	_ = lotus

	blue := newActiveGame(t)
	them := blue.Seats[0]
	pushOneColorSource(blue, them, "Nyx Lotus", "{G4|U1}")
	twoBlue, err := ParseCost("{U}{U}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if plan, ok := blue.AutoTapForCost(them.ID, twoBlue, 0); ok {
		t.Errorf("plan = %v for {U}{U} off a Nyx Lotus with one blue devotion", plan)
	}
}

// The plan's colour is the one the executor mints. Booked for the
// green in {2}{G}, the Lotus must not re-derive a different colour
// off the leftover generic.
func TestMaterializeUsesThePlannedOneColourPick(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lotus := pushOneColorSource(g, me, "Gilded Lotus", "{W3|U3|B3|R3|G3}")

	cost, err := ParseCost("{2}{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	g.WithWriteLock(func() {
		plan, ok := g.autoTapLocked(me.ID, cost, 0, nil)
		if !ok {
			t.Fatalf("no plan for {2}{G} off one Gilded Lotus")
		}
		if len(plan) != 1 || plan[0].CardID != lotus {
			t.Fatalf("plan = %+v, want the Lotus alone", plan)
		}
		if plan[0].OneColor != "G" {
			t.Fatalf("plan booked %q, want the colour the cost demands", plan[0].OneColor)
		}
		g.materializePlanLocked(me, plan, cost)
	})
	if got := poolColors(me); got["G"] != 3 {
		t.Errorf("pool = %v, want three green", got)
	}
}

// A stale plan naming a colour the source no longer offers drops the
// source BEFORE it is tapped, the way the CR 903.4f narrowing and the
// counter-cost checks do — tapping a permanent for no mana is worse
// than not tapping it.
func TestMaterializeDropsAStaleOneColourPickWithoutTapping(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lotus := pushOneColorSource(g, me, "Nyx Lotus", "{G4|U1}")

	cost, _ := ParseCost("{2}")
	g.WithWriteLock(func() {
		// "R" was never on offer: the shape of a plan built before the
		// devotion moved.
		g.materializePlanLocked(me, tapPlan{{CardID: lotus, OneColor: "R"}}, cost)
	})
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want nothing minted", me.ManaPool)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == lotus && c.Tapped {
				t.Error("the Lotus was tapped for a colour it does not offer")
			}
		}
	})
}

// The legal-move enumerator and the lobby preview both ask the
// auto-tapper, so the planner's answer is the one a bot and the cast
// preview see. This pins the shared contract at the engine boundary:
// AutoTapForCostForEffect (the enumerator's entry point) agrees with
// AutoTapForCost (the preview's).
func TestOneColourSourceIsVisibleToBothAutoTapSurfaces(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushOneColorSource(g, me, "Gilded Lotus", "{W3|U3|B3|R3|G3}")
	pushIslandFor(g, me)
	pushIslandFor(g, me)

	cost, _ := ParseCost("{3}{U}{U}")
	if _, ok := g.AutoTapForCost(me.ID, cost, 0); !ok {
		t.Error("the lobby preview surface says the cast is unpayable")
	}
	var ok bool
	g.ReadSnapshot(func() { _, ok = g.AutoTapForCostForEffect(me.ID, cost, 0) })
	if !ok {
		t.Error("the enumerator surface says the cast is unpayable")
	}
}
