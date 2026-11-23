package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// waitForXTriggerThenSettle is waitForXTrigger followed by passing
// the rest of the way to let the trigger resolve.
func waitForXTriggerThenSettle(t *testing.T, g *game.Game, source uuid.UUID) {
	t.Helper()
	effectstest.WaitForXTrigger(t, g, source)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGooseMotherEntersTriggerReadsCastX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goose := effectstest.CastXSpell(t, g, "The Goose Mother", "Legendary Creature — Bird Hydra",
		effectstest.B27TheGooseMotherOracle, "{X}{G}{U}", 3, nil)

	// The permanent has already landed (its own trigger is sourced
	// from it) but the Food does not exist until the trigger — a
	// separate CR 603 object — resolves in its own right.
	effectstest.WaitForXTrigger(t, g, goose)
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("Food while the trigger is still on the stack = %d, want 0", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CounterCount(g, goose, game.CounterPlusOne); got != 3 {
		t.Errorf("entry counters = %d, want 3 (X, CR 614.1c — unaffected by #1357)", got)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("half of X=3 rounded up: %d Food, want 2", got)
	}
}

func TestFarmerCottonEntersTriggerReadsCastX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cotton := effectstest.CastXSpell(t, g, "Farmer Cotton", "Legendary Creature — Halfling Peasant",
		effectstest.B27FarmerCottonOracle, "{X}{G}{W}", 2, nil)

	effectstest.WaitForXTrigger(t, g, cotton)
	if got := effectstest.B27CountNamed(g, me.ID, "Halfling") + effectstest.B27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("tokens while the trigger is still on the stack = %d, want 0", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B27CountNamed(g, me.ID, "Halfling"); got != 2 {
		t.Errorf("Halflings = %d, want X=2", got)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("Food = %d, want X=2", got)
	}
	if spec, _ := effects.Lookup(effectstest.B27FarmerCottonOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("nothing is deferred any more — CompletenessFull")
	}
}

func TestSpringleafParadeEntersTriggerReadsCastX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	parade := effectstest.B12PlayFromHand(t, g, "Springleaf Parade", "Enchantment", effectstest.B18SpringleafParadeOracle,
		game.CastSpellParams{XValue: 3})

	effectstest.WaitForXTrigger(t, g, parade)
	if got := effectstest.B16CountNamed(g, "Shapeshifter"); got != 0 {
		t.Fatalf("Shapeshifters while the trigger is still on the stack = %d, want 0", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B16CountNamed(g, "Shapeshifter"); got != 3 {
		t.Errorf("Shapeshifters = %d, want X=3", got)
	}
	// ADR 0093 made the mana-ability grant real, so nothing is
	// declared any more.
	if spec, _ := effects.Lookup(effectstest.B18SpringleafParadeOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("Springleaf Parade has no simplification left — CompletenessFull")
	}
}

func TestSpitefulBanditryEntersTriggerReadsCastX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	small := effectstest.B16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	big := effectstest.B16Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 5, 5, "G")
	banditry := effectstest.CastXSpell(t, g, "Spiteful Banditry", "Enchantment", effectstest.B21SpitefulBanditryOracle, "{X}{R}{R}", 2, nil)

	effectstest.WaitForXTrigger(t, g, banditry)
	if effectstest.DamageMarkedOn(g, small) != 0 || effectstest.DamageMarkedOn(g, big) != 0 {
		t.Fatal("no damage while the trigger is still on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(small) {
		t.Error("X=2 kills a 1/1")
	}
	if !g.Battlefield.Contains(big) || effectstest.DamageMarkedOn(g, big) != 2 {
		t.Error("the 5/5 takes 2 and lives")
	}
	if n := effectstest.B16CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("an opponent's creature died: 1 Treasure, got %d", n)
	}
	if spec, _ := effects.Lookup(effectstest.B21SpitefulBanditryOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("nothing is deferred any more — CompletenessFull")
	}
}

// TestFlickeredFarmerCottonEntersTriggerReadsZeroCastX — CR 107.3m:
// CastProvenance is stamped only when a permanent comes from a
// resolving spell (executeEntryToBattlefieldLocked); a card that
// enters any other way — reanimated, searched for, or flickered, as
// here — was never cast, so Card.CastX() reads zero and Farmer
// Cotton's ETB trigger, though it still fires, creates nothing.
func TestFlickeredFarmerCottonEntersTriggerReadsZeroCastX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cotton := effectstest.CastXSpell(t, g, "Farmer Cotton", "Legendary Creature — Halfling Peasant",
		effectstest.B27FarmerCottonOracle, "{X}{G}{W}", 3, nil)
	waitForXTriggerThenSettle(t, g, cotton)
	if got := effectstest.B27CountNamed(g, me.ID, "Halfling"); got != 3 {
		t.Fatalf("setup: Halflings = %d, want X=3 before the flicker", got)
	}

	var newCotton uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(cotton); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
		var err error
		newCotton, err = g.ReturnFromExileToBattlefieldForEffect(cotton, uuid.Nil, false)
		if err != nil {
			t.Fatalf("ReturnFromExileToBattlefieldForEffect: %v", err)
		}
	})
	if newCotton == uuid.Nil || newCotton == cotton {
		t.Fatal("the flicker must mint a new object")
	}

	halflingsBefore := effectstest.B27CountNamed(g, me.ID, "Halfling")
	foodBefore := effectstest.B27CountNamed(g, me.ID, "Food")
	waitForXTriggerThenSettle(t, g, newCotton)

	if got := effectstest.B27CountNamed(g, me.ID, "Halfling"); got != halflingsBefore {
		t.Errorf("Halflings after the flicker's ETB = %d, want unchanged %d (CastX()==0, not cast)", got, halflingsBefore)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != foodBefore {
		t.Errorf("Food after the flicker's ETB = %d, want unchanged %d (CastX()==0, not cast)", got, foodBefore)
	}
}
