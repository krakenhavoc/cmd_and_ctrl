package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const coinOfMasteryOracle = "d78518ee-df79-48d1-b9d5-4f968b441899"

// {T} makes a Treasure. The headline — counters per artifact mana — is
// pinned in granted_mana_spent_readers_test.go (ADR 0109 §11); the one
// caveat left is the strict-mana one every reader of spent mana carries.
func TestCoinOfMasteryTapsForATreasureAndDeclaresItsGap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	coin := effectstest.B12Push(g, me.ID, "Coin of Mastery", "Artifact", coinOfMasteryOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)

	effectstest.B16Activate(t, g, me.ID, coin, 0, game.ActivateAbilityParams{})
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	if treasure == uuid.Nil {
		t.Fatal("the tap ability creates a Treasure token")
	}
	if effectstest.ControllerOf(t, g, treasure) != me.ID {
		t.Error("the Treasure is the activator's")
	}
	if err := g.ActivateCatalogAbility(me.ID, coin, 0, game.ActivateAbilityParams{}); err != game.ErrAlreadyTapped {
		t.Errorf("a tapped Coin: got %v, want ErrAlreadyTapped", err)
	}

	spec, ok := effects.Lookup(coinOfMasteryOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Coin of Mastery is complete since a waived payment asks (#3855): %v / %v", spec.Completeness, spec.Caveats)
	}
	// The counters are a replacement on the Coin over another
	// creature's entry, and nothing else.
	if spec.EntersWithCountersFromCast != nil || len(spec.Replacements) != 1 || spec.Static != nil {
		t.Error("the +1/+1 counter rider is one replacement on the Coin")
	}
}
