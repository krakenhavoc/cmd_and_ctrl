package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nghd_loyalty_test.go — a loyalty ability is an activated ability
// (CR 606.1), so naming a planeswalker with Phyrexian Revoker or Pithing
// Needle stops it, on the catalogued path and on the sandbox verb.

const nghdSorinLordOracle = "5c780dda-ee76-482a-a610-53c321c4fdec"

type nghdNamer func(*testing.T, *game.Game, uuid.UUID, string) uuid.UUID

func nghdNamedWalkerTable(t *testing.T, namer nghdNamer, named string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]
	namer(t, g, them.ID, named)
	sorin := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sorin, Lord of Innistrad", OracleID: nghdSorinLordOracle,
		TypeLine: "Legendary Planeswalker — Sorin", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	return g, me, sorin
}

func nghdPithingNeedle(t *testing.T, g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := b21Push(g, owner, "Pithing Needle", "Artifact", actPithingNeedleOracle, 0, 0)
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(owner, id, "Pithing Needle — choose a card name")
	})
	actNameFor(t, g, owner, name)
	return id
}

func nghdNamers() map[string]nghdNamer {
	return map[string]nghdNamer{"Phyrexian Revoker": nghdRevoker, "Pithing Needle": nghdPithingNeedle}
}

// Naming a planeswalker stops its catalogued loyalty abilities, and the
// loyalty is not paid.
func TestNghdNamedPlaneswalkerLoyaltyIsRefusedOnTheCatalogPath(t *testing.T) {
	for name, namer := range nghdNamers() {
		t.Run(name, func(t *testing.T) {
			g, me, sorin := nghdNamedWalkerTable(t, namer, "Sorin, Lord of Innistrad")
			err := g.ActivateCatalogAbility(me.ID, sorin, 0, game.ActivateAbilityParams{})
			if !errors.Is(err, game.ErrCantActivate) {
				t.Fatalf("a named walker's +1: err = %v, want ErrCantActivate", err)
			}
			if got := loyaltyOf(g, sorin); got != 3 {
				t.Errorf("loyalty = %d after a refused activation, want 3", got)
			}
		})
	}
}

// The sandbox verb asks the same gate.
func TestNghdNamedPlaneswalkerLoyaltyIsRefusedOnTheManualPath(t *testing.T) {
	for name, namer := range nghdNamers() {
		t.Run(name, func(t *testing.T) {
			g, me, sorin := nghdNamedWalkerTable(t, namer, "Sorin, Lord of Innistrad")
			err := g.ActivateLoyalty(me.ID, sorin, "+1", 1)
			if !errors.Is(err, game.ErrCantActivate) {
				t.Fatalf("a named walker's manual +1: err = %v, want ErrCantActivate", err)
			}
			if got := loyaltyOf(g, sorin); got != 3 {
				t.Errorf("loyalty = %d after a refused activation, want 3", got)
			}
		})
	}
}

// A walker the name does not match activates on both paths.
func TestNghdUnnamedPlaneswalkerStillActivates(t *testing.T) {
	g, me, sorin := nghdNamedWalkerTable(t, nghdRevoker, "Some Other Card")
	if err := g.ActivateCatalogAbility(me.ID, sorin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Errorf("an unnamed walker's catalogued +1: %v", err)
	}
	g2, me2, sorin2 := nghdNamedWalkerTable(t, nghdRevoker, "Some Other Card")
	if err := g2.ActivateLoyalty(me2.ID, sorin2, "+1", 1); err != nil {
		t.Errorf("an unnamed walker's manual +1: %v", err)
	}
}
