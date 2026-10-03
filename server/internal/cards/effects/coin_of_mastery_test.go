package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const coinOfMasteryOracle = "d78518ee-df79-48d1-b9d5-4f968b441899"

// {T} makes a Treasure. The headline — counters per artifact mana — is
// pinned in granted_mana_spent_readers_test.go (ADR 0109 §11); the one
// caveat left is the strict-mana one every reader of spent mana carries.
func TestCoinOfMasteryTapsForATreasureAndDeclaresItsGap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	coin := b12Push(g, me.ID, "Coin of Mastery", "Artifact", coinOfMasteryOracle, 0, 0)
	advanceToMain(t, g)

	b16Activate(t, g, me.ID, coin, 0, game.ActivateAbilityParams{})
	treasure := findBattlefieldByName(g, "Treasure")
	if treasure == uuid.Nil {
		t.Fatal("the tap ability creates a Treasure token")
	}
	if controllerOf(t, g, treasure) != me.ID {
		t.Error("the Treasure is the activator's")
	}
	if err := g.ActivateCatalogAbility(me.ID, coin, 0, game.ActivateAbilityParams{}); err != game.ErrAlreadyTapped {
		t.Errorf("a tapped Coin: got %v, want ErrAlreadyTapped", err)
	}

	spec, ok := Lookup(coinOfMasteryOracle)
	if !ok || spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Fatalf("Coin of Mastery ships with exactly one caveat: %v / %v", spec.Completeness, spec.Caveats)
	}
	// The counters are a replacement on the Coin over another
	// creature's entry, and nothing else.
	if spec.EntersWithCountersFromCast != nil || len(spec.Replacements) != 1 || spec.Static != nil {
		t.Error("the +1/+1 counter rider is one replacement on the Coin")
	}
}
