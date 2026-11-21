package testsms

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func nghdPithingNeedle(t *testing.T, g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := effectstest.B21Push(g, owner, "Pithing Needle", "Artifact", effectstest.ActPithingNeedleOracle, 0, 0)
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(owner, id, "Pithing Needle — choose a card name")
	})
	effectstest.ActNameFor(t, g, owner, name)
	return id
}

func nghdNamers() map[string]effectstest.NghdNamer {
	return map[string]effectstest.NghdNamer{"Phyrexian Revoker": nghdRevoker, "Pithing Needle": nghdPithingNeedle}
}

// Naming a planeswalker stops its catalogued loyalty abilities, and the
// loyalty is not paid.
func TestNghdNamedPlaneswalkerLoyaltyIsRefusedOnTheCatalogPath(t *testing.T) {
	for name, namer := range nghdNamers() {
		t.Run(name, func(t *testing.T) {
			g, me, sorin := effectstest.NghdNamedWalkerTable(t, namer, "Sorin, Lord of Innistrad")
			err := g.ActivateCatalogAbility(me.ID, sorin, 0, game.ActivateAbilityParams{})
			if !errors.Is(err, game.ErrCantActivate) {
				t.Fatalf("a named walker's +1: err = %v, want ErrCantActivate", err)
			}
			if got := effectstest.LoyaltyOf(g, sorin); got != 3 {
				t.Errorf("loyalty = %d after a refused activation, want 3", got)
			}
		})
	}
}

// The sandbox verb asks the same gate.
func TestNghdNamedPlaneswalkerLoyaltyIsRefusedOnTheManualPath(t *testing.T) {
	for name, namer := range nghdNamers() {
		t.Run(name, func(t *testing.T) {
			g, me, sorin := effectstest.NghdNamedWalkerTable(t, namer, "Sorin, Lord of Innistrad")
			err := g.ActivateLoyalty(me.ID, sorin, "+1", 1)
			if !errors.Is(err, game.ErrCantActivate) {
				t.Fatalf("a named walker's manual +1: err = %v, want ErrCantActivate", err)
			}
			if got := effectstest.LoyaltyOf(g, sorin); got != 3 {
				t.Errorf("loyalty = %d after a refused activation, want 3", got)
			}
		})
	}
}

// A walker the name does not match activates on both paths.
func TestNghdUnnamedPlaneswalkerStillActivates(t *testing.T) {
	g, me, sorin := effectstest.NghdNamedWalkerTable(t, nghdRevoker, "Some Other Card")
	if err := g.ActivateCatalogAbility(me.ID, sorin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Errorf("an unnamed walker's catalogued +1: %v", err)
	}
	g2, me2, sorin2 := effectstest.NghdNamedWalkerTable(t, nghdRevoker, "Some Other Card")
	if err := g2.ActivateLoyalty(me2.ID, sorin2, "+1", 1); err != nil {
		t.Errorf("an unnamed walker's manual +1: %v", err)
	}
}
