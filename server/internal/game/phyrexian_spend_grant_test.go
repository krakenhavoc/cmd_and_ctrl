package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// phyrexian_spend_grant_test.go — #1589, ADR 0066's 2026-09-28
// amendment. Under "spend mana as though it were mana of any color"
// (Breeches) or "mana of any type" (Hostage Taker, Gonti) a Phyrexian
// symbol keeps its "or 2 life" half (CR 107.4f). The fold in
// spendAsThoughAny used to turn it into generic mana, so a stolen
// Dismember cost three mana and a life claim was refused as an
// over-claim. It now stays a Phyrexian requirement that any mana pays
// (ColorRequirement.AnyMana).

// spendGrants is the two clauses every test here runs under.
var spendGrants = []struct {
	name string
	perm CastPermission
}{
	{"any color", CastPermission{AnyColor: true}},
	{"any type", CastPermission{AnyType: true}},
}

// stolenPhyrexianFixture exiles a card with `cost` off the opponent's
// library under `perm`, in the thief's main phase — Gonti / Hostage
// Taker shaped, through the impulse-exile primitive.
func stolenPhyrexianFixture(t *testing.T, perm CastPermission, cost string) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	opp.Library.Cards = nil
	seedTop(opp, "Stolen Dismember", "Sorcery", cost)
	return g, me, impulseExile(t, g, opp, me, perm)
}

// The fold itself: Phyrexian slots survive both clauses, widened and
// ordered behind the slots kept as printed; nothing else changes.
func TestSpendGrantsKeepPhyrexianSymbols(t *testing.T) {
	cost, err := ParseCost("{1}{B/P}{C}{W/U/P}{R}")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		perm     *CastPermission
		generic  int
		str      string
		widened  int
		colorles int
	}{
		// Every coloured slot widens; {C} stays unwidened (CR 106.1b),
		// and the widened slots follow it.
		{"any color", &CastPermission{AnyColor: true}, 1, "{1}{C}{B/P}{W/U/P}{R}", 3, 1},
		// Every slot widens, {C} included.
		{"any type", &CastPermission{AnyType: true}, 1, "{1}{B/P}{C}{W/U/P}{R}", 4, 1},
		// No grant: the parse, untouched.
		{"no grant", nil, 1, "{1}{B/P}{C}{W/U/P}{R}", 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := spendAsThoughAny(tc.perm, cost)
			if got.Generic != tc.generic || got.String() != tc.str {
				t.Errorf("folded = %s (generic %d), want %s (generic %d)", got.String(), got.Generic, tc.str, tc.generic)
			}
			if n := got.PhyrexianSymbols(); n != 2 {
				t.Errorf("PhyrexianSymbols = %d, want 2 — the life half survives the grant", n)
			}
			widened, colorless := 0, 0
			for i, r := range got.Required {
				if r.AnyMana {
					widened++
				} else if widened > 0 {
					t.Errorf("slot %d (%s) follows a widened slot; widened slots go last", i, r)
				}
				if requiresColorless(r) {
					colorless++
				}
			}
			if widened != tc.widened || colorless != tc.colorles {
				t.Errorf("widened %d, {C} %d; want %d and %d", widened, colorless, tc.widened, tc.colorles)
			}
		})
	}
}

// Paying 2 life for each Phyrexian symbol, under both clauses: one
// Mountain's {R} pays the {1}, and the two {B/P} cost 4 life. Before
// #1589 the fold left no Phyrexian symbol to claim and the cast was
// refused with ErrInvalidParam.
func TestSpendGrantPhyrexianPaysWithLife(t *testing.T) {
	for _, tc := range spendGrants {
		t.Run(tc.name, func(t *testing.T) {
			g, me, loot := stolenPhyrexianFixture(t, tc.perm, "{1}{B/P}{B/P}")
			life := me.Life
			me.ManaPool.AddMana(ManaToken{Color: "R"})
			if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true, PhyrexianLife: 2}); err != nil {
				t.Fatalf("cast paying both {B/P} with life: %v", err)
			}
			if !g.Stack.Contains(loot) {
				t.Fatal("the spell is not on the stack")
			}
			if me.Life != life-2*PhyrexianLifePerSymbol {
				t.Errorf("life = %d, want %d", me.Life, life-2*PhyrexianLifePerSymbol)
			}
			if len(me.ManaPool) != 0 {
				t.Errorf("pool = %v, want empty", me.ManaPool)
			}
		})
	}
}

// Paying with off-colour mana, under both clauses: the widened slot's
// mana half takes red for a {B/P}, and a mix of life and red works too.
func TestSpendGrantPhyrexianPaysWithOffColorMana(t *testing.T) {
	for _, tc := range spendGrants {
		t.Run(tc.name+"/all mana", func(t *testing.T) {
			g, me, loot := stolenPhyrexianFixture(t, tc.perm, "{1}{B/P}{B/P}")
			life := me.Life
			me.ManaPool.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "R"}, ManaToken{Color: "R"})
			if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
				t.Fatalf("cast paying {1}{B/P}{B/P} with three red: %v", err)
			}
			if me.Life != life || len(me.ManaPool) != 0 {
				t.Errorf("life %d (want %d), pool %v (want empty)", me.Life, life, me.ManaPool)
			}
		})
		t.Run(tc.name+"/one of each", func(t *testing.T) {
			g, me, loot := stolenPhyrexianFixture(t, tc.perm, "{1}{B/P}{B/P}")
			life := me.Life
			me.ManaPool.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "G"})
			if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true, PhyrexianLife: 1}); err != nil {
				t.Fatalf("cast paying one {B/P} with life and one with green: %v", err)
			}
			if me.Life != life-PhyrexianLifePerSymbol || len(me.ManaPool) != 0 {
				t.Errorf("life %d (want %d), pool %v (want empty)", me.Life, life-PhyrexianLifePerSymbol, me.ManaPool)
			}
		})
	}
}

// A hybrid Phyrexian symbol (CR 107.4f) follows the same rule: red
// pays a {W/U/P} under either grant, and so does 2 life.
func TestSpendGrantHybridPhyrexianKeepsBothHalves(t *testing.T) {
	for _, tc := range spendGrants {
		t.Run(tc.name+"/mana", func(t *testing.T) {
			g, me, loot := stolenPhyrexianFixture(t, tc.perm, "{W/U/P}")
			me.ManaPool.AddMana(ManaToken{Color: "R"})
			if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
				t.Fatalf("red for {W/U/P}: %v", err)
			}
		})
		t.Run(tc.name+"/life", func(t *testing.T) {
			g, me, loot := stolenPhyrexianFixture(t, tc.perm, "{W/U/P}")
			life := me.Life
			if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true, PhyrexianLife: 1}); err != nil {
				t.Fatalf("2 life for {W/U/P}: %v", err)
			}
			if me.Life != life-PhyrexianLifePerSymbol {
				t.Errorf("life = %d, want %d", me.Life, life-PhyrexianLifePerSymbol)
			}
		})
	}
}

// Through the auto-tapper, under both clauses: three Mountains pay
// {1}{B/P}{B/P} with no life claimed, and one Mountain pays it with
// both symbols claimed — tapping exactly one land.
func TestSpendGrantPhyrexianThroughTheAutoTapper(t *testing.T) {
	for _, tc := range spendGrants {
		t.Run(tc.name+"/mana", func(t *testing.T) {
			g, me, loot := stolenPhyrexianFixture(t, tc.perm, "{1}{B/P}{B/P}")
			pushUntappedLands(t, g, me, 3, "Basic Land — Mountain")
			if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true, AutoTap: true}); err != nil {
				t.Fatalf("auto-tapped cast off three Mountains: %v", err)
			}
			if got := tappedLands(g, me.ID); got != 3 {
				t.Errorf("tapped %d lands, want 3", got)
			}
		})
		t.Run(tc.name+"/life", func(t *testing.T) {
			g, me, loot := stolenPhyrexianFixture(t, tc.perm, "{1}{B/P}{B/P}")
			pushUntappedLands(t, g, me, 1, "Basic Land — Mountain")
			life := me.Life
			if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true, AutoTap: true, PhyrexianLife: 2}); err != nil {
				t.Fatalf("auto-tapped cast off one Mountain and 4 life: %v", err)
			}
			if got := tappedLands(g, me.ID); got != 1 {
				t.Errorf("tapped %d lands, want 1", got)
			}
			if me.Life != life-2*PhyrexianLifePerSymbol {
				t.Errorf("life = %d, want %d", me.Life, life-2*PhyrexianLifePerSymbol)
			}
		})
	}
}

// The {C} exception still holds under the any-colour grant: a
// Phyrexian slot beside it widens, the {C} does not, so two red are
// short and a colorless plus a red is enough. Any type pays both off
// red.
func TestSpendGrantColorlessStillNeedsColorlessUnderAnyColor(t *testing.T) {
	for _, tc := range []struct {
		name string
		perm CastPermission
		pool []ManaToken
		ok   bool
	}{
		{"any color/two red", CastPermission{AnyColor: true}, []ManaToken{{Color: "R"}, {Color: "R"}}, false},
		{"any color/colorless and red", CastPermission{AnyColor: true}, []ManaToken{{Color: "R"}, {Color: "C"}}, true},
		{"any type/two red", CastPermission{AnyType: true}, []ManaToken{{Color: "R"}, {Color: "R"}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, loot := stolenPhyrexianFixture(t, tc.perm, "{B/P}{C}")
			me.ManaPool.AddMana(tc.pool...)
			err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true})
			if tc.ok && err != nil {
				t.Fatalf("cast: %v", err)
			}
			if !tc.ok {
				var short *InsufficientManaError
				if !errors.As(err, &short) {
					t.Fatalf("cast: %v, want insufficient mana — {C} needs colorless under any colour", err)
				}
			}
		})
	}
}

// With no grant nothing changes: red does not pay {B/P}, and the life
// half still does.
func TestNoSpendGrantPhyrexianUnchanged(t *testing.T) {
	g, me, loot := stolenPhyrexianFixture(t, CastPermission{}, "{1}{B/P}{B/P}")
	me.ManaPool.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "R"}, ManaToken{Color: "R"})
	var short *InsufficientManaError
	if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true}); !errors.As(err, &short) {
		t.Fatalf("three red for {1}{B/P}{B/P} with no grant: %v, want insufficient mana", err)
	}
	life := me.Life
	if err := g.CastSpell(me.ID, loot, CastSpellParams{FromZone: "exile", Strict: true, PhyrexianLife: 2}); err != nil {
		t.Fatalf("red and 4 life with no grant: %v", err)
	}
	if me.Life != life-2*PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d", me.Life, life-2*PhyrexianLifePerSymbol)
	}
}

// The pool solver pays a widened slot LAST, whatever order Required
// arrives in — a cost modifier appends its coloured symbols after the
// fold, and a first-match walk that let the wildcard go first would
// spend the only Plains on a {B/P} the Swamp could have paid.
func TestWidenedPhyrexianSlotIsPaidLast(t *testing.T) {
	cost := ParsedCost{Required: []ColorRequirement{
		{Options: []string{"B"}, Phyrexian: true, AnyMana: true},
		{Options: []string{"W"}},
	}}
	pool := ManaPool{{Color: "W"}, {Color: "B"}}
	if !pool.CanPay(cost, 0) {
		t.Fatal("a Plains and a Swamp should pay {B/P}(any) + {W}")
	}
	if missing := pool.Missing(cost, 0); missing != nil {
		t.Errorf("Missing = %v, want nil", missing)
	}
}

// tappedLands counts the tapped permanents `owner` controls.
func tappedLands(g *Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Tapped {
			n++
		}
	}
	return n
}
