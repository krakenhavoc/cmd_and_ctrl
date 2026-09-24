package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const woeStriderOracle = "3adbd963-e85d-4569-963a-4472594f06f9"

// TestWoeStriderEntersWithAGoatToken pins the ETB half — the escape
// declaration and the escaped-counters half are pinned against the
// shared roster in escape_test.go.
func TestWoeStriderEntersWithAGoatToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := b43TokensNamed(g, me.ID, "Goat")

	castCatalogSpell(t, g, "Woe Strider", "Creature — Horror", woeStriderOracle, nil)
	passPriorityAroundTable(t, g)

	if got := b43TokensNamed(g, me.ID, "Goat"); got != before+1 {
		t.Fatalf("Goat tokens %d -> %d, want +1", before, got)
	}
}

// The free sacrifice outlet: any other creature, including the token
// it just made, pays for a scry.
func TestWoeStriderSacrificesAnotherCreatureToScry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	strider := pushCatalogPermanent(g, me.ID, "Woe Strider", "Creature — Horror", woeStriderOracle, false)
	fodder := pushVanillaCreature(g, me.ID, "Fodder Bear", 1, 1)
	seedLibrary(me, "Top Card")

	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("Sacrifice another creature: Scry 1: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry prompt after the sacrifice")
	}
}

// Sandbox simplification, declared: "another" is enforced by name, so
// a SECOND Woe Strider on the battlefield can't feed its own ability.
func TestWoeStriderCannotSacrificeASecondCopyOfItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	strider := pushCatalogPermanent(g, me.ID, "Woe Strider", "Creature — Horror", woeStriderOracle, false)
	second := pushCatalogPermanent(g, me.ID, "Woe Strider", "Creature — Horror", woeStriderOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{second},
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("sacrificing a second Woe Strider: err = %v, want ErrIllegalTarget", err)
	}
}
