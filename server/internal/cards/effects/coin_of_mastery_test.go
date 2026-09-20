package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const coinOfMasteryOracle = "d78518ee-df79-48d1-b9d5-4f968b441899"

// The half that ships: {T} makes a Treasure. The headline does not,
// and the caveat is what says so on the catalogue page.
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
	// The counters rider is not declared anywhere: no entry-counter
	// clause, no replacement, no static.
	if spec.EntersWithCountersFromCast != nil || spec.Replacements != nil || spec.Static != nil {
		t.Error("the +1/+1 counter rider is deferred, not half-declared")
	}
}
