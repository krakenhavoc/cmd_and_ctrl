package testsu

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestWoeStriderEntersWithAGoatToken pins the ETB half — the escape
// declaration and the escaped-counters half are pinned against the
// shared roster in escape_test.go.
func TestWoeStriderEntersWithAGoatToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.B43TokensNamed(g, me.ID, "Goat")

	effectstest.CastCatalogSpell(t, g, "Woe Strider", "Creature — Horror", effectstest.WoeStriderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B43TokensNamed(g, me.ID, "Goat"); got != before+1 {
		t.Fatalf("Goat tokens %d -> %d, want +1", before, got)
	}
}

// The free sacrifice outlet: any other creature, including the token
// it just made, pays for a scry.
func TestWoeStriderSacrificesAnotherCreatureToScry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	strider := effectstest.PushCatalogPermanent(g, me.ID, "Woe Strider", "Creature — Horror", effectstest.WoeStriderOracle, false)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder Bear", 1, 1)
	effectstest.SeedLibrary(me, "Top Card")

	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("Sacrifice another creature: Scry 1: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry prompt after the sacrifice")
	}
}

// "Another" is object identity (#1738): a SECOND Woe Strider can feed
// the first one's ability, and the first cannot feed itself.
func TestWoeStriderSacrificesASecondCopyButNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	strider := effectstest.PushCatalogPermanent(g, me.ID, "Woe Strider", "Creature — Horror", effectstest.WoeStriderOracle, false)
	second := effectstest.PushCatalogPermanent(g, me.ID, "Woe Strider", "Creature — Horror", effectstest.WoeStriderOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{strider},
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("sacrificing itself: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, strider, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{second},
	}); err != nil {
		t.Errorf("sacrificing a second Woe Strider: %v", err)
	}
}
