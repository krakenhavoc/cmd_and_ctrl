package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// x_enters_trigger_cast_x_test.go — #1357. Before #1312
// (CastProvenance.X / Card.CastX(), CR 107.3m), an ETB trigger's Build
// had nowhere to read the X a spell was announced with — the
// resolving spell's StackItem is already gone by the time a trigger's
// Build runs — so The Goose Mother, Farmer Cotton, Springleaf Parade
// and Spiteful Banditry all moved their printed enters-TRIGGER effect
// into OnResolve instead, a beat before the permanent existed, with a
// declared caveat that the effect arrived without a trigger on the
// stack to respond to.
//
// Each test below proves the real shape now that #1312 has closed the
// seam: nothing happens as the spell resolves, the trigger reaches
// StackMeta as its own CR 603 object (waitForXTriggerThenSettle
// blocks there before letting it resolve), and only THEN does the
// printed effect happen, reading the X off Card.CastX().
// TestFlickeredFarmerCottonEntersTriggerReadsZeroCastX covers CR
// 107.3m's other half: a permanent that enters WITHOUT being cast —
// flickered, in this case — was never cast at all, so its trigger's
// CastX() is zero and its enters ability does nothing.

// waitForXTrigger passes priority until `source`'s ETB trigger
// reaches StackMeta and returns it — proving it is a real,
// independently respondable stack object rather than a side effect
// folded into the spell's own resolution. Fails the test if it never
// arrives (which is exactly what happens run against the pre-#1357
// shape: the token/damage effect lived in OnResolve, so no such
// trigger is ever queued at all).
func waitForXTrigger(t *testing.T, g *game.Game, source uuid.UUID) *game.StackItem {
	t.Helper()
	for i := 0; i < 8; i++ {
		if item := triggerOnStack(g, source); item != nil {
			return item
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority waiting for the enters trigger: %v", err)
		}
	}
	t.Fatal("the enters trigger never reached the stack")
	return nil
}

// waitForXTriggerThenSettle is waitForXTrigger followed by passing
// the rest of the way to let the trigger resolve.
func waitForXTriggerThenSettle(t *testing.T, g *game.Game, source uuid.UUID) {
	t.Helper()
	waitForXTrigger(t, g, source)
	passPriorityAroundTable(t, g)
}

func TestGooseMotherEntersTriggerReadsCastX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	goose := castXSpell(t, g, "The Goose Mother", "Legendary Creature — Bird Hydra",
		b27TheGooseMotherOracle, "{X}{G}{U}", 3, nil)

	// The permanent has already landed (its own trigger is sourced
	// from it) but the Food does not exist until the trigger — a
	// separate CR 603 object — resolves in its own right.
	waitForXTrigger(t, g, goose)
	if got := b27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("Food while the trigger is still on the stack = %d, want 0", got)
	}
	passPriorityAroundTable(t, g)

	if got := counterCount(g, goose, game.CounterPlusOne); got != 3 {
		t.Errorf("entry counters = %d, want 3 (X, CR 614.1c — unaffected by #1357)", got)
	}
	if got := b27CountNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("half of X=3 rounded up: %d Food, want 2", got)
	}
}

func TestFarmerCottonEntersTriggerReadsCastX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cotton := castXSpell(t, g, "Farmer Cotton", "Legendary Creature — Halfling Peasant",
		b27FarmerCottonOracle, "{X}{G}{W}", 2, nil)

	waitForXTrigger(t, g, cotton)
	if got := b27CountNamed(g, me.ID, "Halfling") + b27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("tokens while the trigger is still on the stack = %d, want 0", got)
	}
	passPriorityAroundTable(t, g)

	if got := b27CountNamed(g, me.ID, "Halfling"); got != 2 {
		t.Errorf("Halflings = %d, want X=2", got)
	}
	if got := b27CountNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("Food = %d, want X=2", got)
	}
	if spec, _ := Lookup(b27FarmerCottonOracle); spec.Completeness != CompletenessFull {
		t.Error("nothing is deferred any more — CompletenessFull")
	}
}

func TestSpringleafParadeEntersTriggerReadsCastX(t *testing.T) {
	g := newCatalogGame(t)
	parade := b12PlayFromHand(t, g, "Springleaf Parade", "Enchantment", b18SpringleafParadeOracle,
		game.CastSpellParams{XValue: 3})

	waitForXTrigger(t, g, parade)
	if got := b16CountNamed(g, "Shapeshifter"); got != 0 {
		t.Fatalf("Shapeshifters while the trigger is still on the stack = %d, want 0", got)
	}
	passPriorityAroundTable(t, g)

	if got := b16CountNamed(g, "Shapeshifter"); got != 3 {
		t.Errorf("Shapeshifters = %d, want X=3", got)
	}
	// ADR 0093 made the mana-ability grant real, so nothing is
	// declared any more.
	if spec, _ := Lookup(b18SpringleafParadeOracle); spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("Springleaf Parade has no simplification left — CompletenessFull")
	}
}

func TestSpitefulBanditryEntersTriggerReadsCastX(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	big := b16Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 5, 5, "G")
	banditry := castXSpell(t, g, "Spiteful Banditry", "Enchantment", b21SpitefulBanditryOracle, "{X}{R}{R}", 2, nil)

	waitForXTrigger(t, g, banditry)
	if damageMarkedOn(g, small) != 0 || damageMarkedOn(g, big) != 0 {
		t.Fatal("no damage while the trigger is still on the stack")
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(small) {
		t.Error("X=2 kills a 1/1")
	}
	if !g.Battlefield.Contains(big) || damageMarkedOn(g, big) != 2 {
		t.Error("the 5/5 takes 2 and lives")
	}
	if n := b16CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("an opponent's creature died: 1 Treasure, got %d", n)
	}
	if spec, _ := Lookup(b21SpitefulBanditryOracle); spec.Completeness != CompletenessFull {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	cotton := castXSpell(t, g, "Farmer Cotton", "Legendary Creature — Halfling Peasant",
		b27FarmerCottonOracle, "{X}{G}{W}", 3, nil)
	waitForXTriggerThenSettle(t, g, cotton)
	if got := b27CountNamed(g, me.ID, "Halfling"); got != 3 {
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

	halflingsBefore := b27CountNamed(g, me.ID, "Halfling")
	foodBefore := b27CountNamed(g, me.ID, "Food")
	waitForXTriggerThenSettle(t, g, newCotton)

	if got := b27CountNamed(g, me.ID, "Halfling"); got != halflingsBefore {
		t.Errorf("Halflings after the flicker's ETB = %d, want unchanged %d (CastX()==0, not cast)", got, halflingsBefore)
	}
	if got := b27CountNamed(g, me.ID, "Food"); got != foodBefore {
		t.Errorf("Food after the flicker's ETB = %d, want unchanged %d (CastX()==0, not cast)", got, foodBefore)
	}
}
