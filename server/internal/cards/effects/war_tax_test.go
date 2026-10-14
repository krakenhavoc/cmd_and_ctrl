package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// war_tax_test.go — War Tax and the any-defender attack tax (#3456).

const warTaxOracle = "fd5ff6ad-f3f1-470f-adef-fb4ce6536365"

func TestWarTaxIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(warTaxOracle)
	if !ok || spec.Name != "War Tax" {
		t.Fatalf("War Tax is not registered: %+v", spec)
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

func warTaxActivate(t *testing.T, g *game.Game, me *game.Player, tax uuid.UUID, x int) {
	t.Helper()
	fillPool(me, x)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	if err := g.ActivateCatalogAbility(me.ID, tax, 0, game.ActivateAbilityParams{XValue: x, Strict: true}); err != nil {
		t.Fatalf("activate at X=%d: %v", x, err)
	}
	passPriorityAroundTable(t, g)
}

func warTaxPrice(g *game.Game, attacker, target uuid.UUID) game.AttackTaxPrice {
	return g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: attacker, Target: target}})
}

// warTaxTable is the active seat with War Tax, and an attacker under
// the next seat.
func warTaxTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	tax := pushCatalogPermanent(g, me.ID, "War Tax", "Enchantment", warTaxOracle, false)
	raider := tdCreature(g, opp.ID, "Raider", 2, 2)
	return g, me, opp, tax, raider
}

func TestWarTaxAtXTwoTaxesEveryAttackerAndAnyTarget(t *testing.T) {
	g, me, _, tax, raider := warTaxTable(t)
	warTaxActivate(t, g, me, tax, 2)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	for name, target := range map[string]uuid.UUID{"its controller": me.ID, "another player": third.ID} {
		if p := warTaxPrice(g, raider, target); p.Total.Generic != 2 {
			t.Errorf("attack on %s: price %q, want {2}", name, p.Cost)
		}
	}
	walker := pushCatalogPermanent(g, third.ID, "Test Walker", "Legendary Planeswalker — Test", "test-walker", false)
	if p := warTaxPrice(g, raider, walker); p.Total.Generic != 2 {
		t.Errorf("attack on a planeswalker: price %q, want {2}", p.Cost)
	}
}

func TestWarTaxAtXZeroTaxesNothing(t *testing.T) {
	g, me, _, tax, raider := warTaxTable(t)
	warTaxActivate(t, g, me, tax, 0)
	if p := warTaxPrice(g, raider, me.ID); !p.IsFree() {
		t.Errorf("X=0 priced %q, want free", p.Cost)
	}
}

func TestWarTaxActivationsStack(t *testing.T) {
	g, me, _, tax, raider := warTaxTable(t)
	warTaxActivate(t, g, me, tax, 1)
	warTaxActivate(t, g, me, tax, 2)
	if p := warTaxPrice(g, raider, me.ID); p.Total.Generic != 3 || len(p.Lines) != 2 {
		t.Errorf("price %q with %d lines, want 3 from two", p.Cost, len(p.Lines))
	}
}

func TestWarTaxEndsAtEndOfTurn(t *testing.T) {
	g, me, _, tax, raider := warTaxTable(t)
	seat := g.Turn.ActiveSeat
	warTaxActivate(t, g, me, tax, 2)
	if p := warTaxPrice(g, raider, me.ID); p.Total.Generic != 2 {
		t.Fatalf("mid-turn price %q, want {2}", p.Cost)
	}
	for g.Turn.ActiveSeat == seat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if p := warTaxPrice(g, raider, me.ID); !p.IsFree() {
		t.Errorf("still priced %q after end of turn", p.Cost)
	}
}
