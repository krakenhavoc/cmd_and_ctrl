package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const coralhelmCommanderOracle = "0ca0a186-026d-4cec-9c23-15d94e6acca9"

// coralhelmTable is seat 0's main phase with a Coralhelm Commander, one
// of its controller's Merfolk, one of an opponent's and a non-Merfolk.
func coralhelmTable(t *testing.T) (g *game.Game, me *game.Player, commander, myFish, theirFish, bear uuid.UUID) {
	t.Helper()
	g, me, opp := spendTable(t)
	commander = b12Push(g, me.ID, "Coralhelm Commander", "Creature — Merfolk Soldier", coralhelmCommanderOracle, 2, 2)
	myFish = b12Creature(g, me.ID, "My Merfolk", "Creature — Merfolk Wizard", 1, 1)
	theirFish = b12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk Rogue", 1, 1)
	bear = b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	return g, me, commander, myFish, theirFish, bear
}

// coralhelmLevelUp pays {1} for level up and resolves it.
func coralhelmLevelUp(t *testing.T, g *game.Game, me *game.Player, commander uuid.UUID) {
	t.Helper()
	apaMana(me, "C")
	p7Activate(t, g, me, commander, 0, game.ActivateAbilityParams{Strict: true})
}

// coralhelmWant checks the Commander's level counters, power, toughness
// and flying, and the size of the controller's other Merfolk.
func coralhelmWant(t *testing.T, g *game.Game, commander, myFish uuid.UUID, level, pt int, flying bool, fishPT int) {
	t.Helper()
	if got := countersOn(g, commander, game.CounterLevel); got != level {
		t.Fatalf("%d level counters, want %d", got, level)
	}
	if p, tg := effectivePower(t, g, commander), effectiveToughness(t, g, commander); p != pt || tg != pt {
		t.Errorf("level %d: Commander is %d/%d, want %d/%d", level, p, tg, pt, pt)
	}
	if got := hasEffectiveKeyword(t, g, commander, "flying"); got != flying {
		t.Errorf("level %d: flying %v, want %v", level, got, flying)
	}
	if p, tg := effectivePower(t, g, myFish), effectiveToughness(t, g, myFish); p != fishPT || tg != fishPT {
		t.Errorf("level %d: the other Merfolk is %d/%d, want %d/%d", level, p, tg, fishPT, fishPT)
	}
}

func TestCoralhelmCommanderIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(coralhelmCommanderOracle)
	if !ok {
		t.Fatal("Coralhelm Commander is not registered")
	}
	if spec.Name != "Coralhelm Commander" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// Level up by level up, through both bands and past the last one
// (CR 711.2a, 711.2b, 711.4, 711.5).
func TestCoralhelmCommanderLevelsUpThroughItsBands(t *testing.T) {
	g, me, commander, myFish, theirFish, bear := coralhelmTable(t)
	coralhelmWant(t, g, commander, myFish, 0, 2, false, 1)
	for _, step := range []struct {
		level, pt int
		flying    bool
		fishPT    int
	}{
		{1, 2, false, 1}, // below LEVEL 2-3: the uppermost box (CR 711.5)
		{2, 3, true, 1},
		{3, 3, true, 1},
		{4, 4, true, 2},
		{5, 4, true, 2}, // level up still works at LEVEL 4+ (CR 711.4)
	} {
		coralhelmLevelUp(t, g, me, commander)
		coralhelmWant(t, g, commander, myFish, step.level, step.pt, step.flying, step.fishPT)
	}
	// "OTHER Merfolk creatures YOU CONTROL".
	if p := effectivePower(t, g, theirFish); p != 1 {
		t.Errorf("an opponent's Merfolk has power %d, want 1", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a Bear has power %d, want 2", p)
	}
}

// The bands count level counters whoever put them there, and stop
// applying when the count leaves them.
func TestCoralhelmCommanderBandsFollowTheCounters(t *testing.T) {
	g, _, commander, myFish, _, _ := coralhelmTable(t)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(commander, game.CounterLevel, 4); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	coralhelmWant(t, g, commander, myFish, 4, 4, true, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(commander, game.CounterLevel, -1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	coralhelmWant(t, g, commander, myFish, 3, 3, true, 1)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(commander, game.CounterLevel, -2); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	coralhelmWant(t, g, commander, myFish, 1, 2, false, 1)
}

// A band sets BASE power and toughness (CR 711.2a), so a +1/+1 counter
// counts on top of it, and +1/+1 counters are not level counters.
func TestCoralhelmCommanderCountersApplyOnTopOfTheBand(t *testing.T) {
	g, _, commander, myFish, _, _ := coralhelmTable(t)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(commander, game.CounterPlusOne, 4); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	coralhelmWant(t, g, commander, myFish, 0, 2, false, 1)
	if got := effectivePowerWithCounters(t, g, commander); got != 6 {
		t.Errorf("a 2/2 with four +1/+1 counters has power %d, want 6", got)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(commander, game.CounterLevel, 2); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	if got := effectivePowerWithCounters(t, g, commander); got != 7 {
		t.Errorf("a level 2 Commander with four +1/+1 counters has power %d, want 7", got)
	}
}

// "Level up only as a sorcery" (CR 702.87a), and it costs {1}.
func TestCoralhelmCommanderLevelUpIsSorcerySpeedAndCostsOne(t *testing.T) {
	g, me, commander, _, _, _ := coralhelmTable(t)
	if err := g.ActivateCatalogAbility(me.ID, commander, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("level up was activated with an empty mana pool")
	}
	advanceTo(t, g, game.StepBeginCombat)
	apaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, commander, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("level up was activated at the beginning of combat")
	}
	if got := countersOn(g, commander, game.CounterLevel); got != 0 {
		t.Errorf("%d level counters, want 0", got)
	}
}

// A level symbol is a keyword ability that represents a static ability
// (CR 711.2), so a Commander that loses its abilities loses its bands,
// level up included, whatever its counters say.
func TestCoralhelmCommanderLosesItsBandsWithItsAbilities(t *testing.T) {
	g, me, commander, myFish, _, _ := coralhelmTable(t)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(commander, game.CounterLevel, 4); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	coralhelmWant(t, g, commander, myFish, 4, 4, true, 2)
	enchant(t, g, "Darksteel Mutation", darksteelMutationOracle, commander)
	if hasEffectiveKeyword(t, g, commander, "flying") {
		t.Error("a Commander with no abilities still has flying")
	}
	if p, tg := effectivePower(t, g, myFish), effectiveToughness(t, g, myFish); p != 1 || tg != 1 {
		t.Errorf("the other Merfolk is %d/%d under a Commander with no abilities, want 1/1", p, tg)
	}
	if p, tg := effectivePower(t, g, commander), effectiveToughness(t, g, commander); p != 0 || tg != 1 {
		t.Errorf("the mutated Commander is %d/%d, want 0/1", p, tg)
	}
	apaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, commander, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a Commander with no abilities leveled up")
	}
}

// A copy has the same level symbols and counts its OWN level counters:
// counters are not copiable values (CR 707.2), so a copy of a level 4
// Commander is a 2/2 until it levels up itself.
func TestCoralhelmCommanderCopyReadsItsOwnLevelCounters(t *testing.T) {
	g, me, commander, myFish, _, _ := coralhelmTable(t)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(commander, game.CounterLevel, 4); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	clone := b12Creature(g, me.ID, "Clone", "Creature — Shapeshifter", 0, 0)
	g.WithWriteLock(func() {
		v, _ := g.CopiableValuesForEffect(commander)
		if !g.BecomeCopyForEffect(uuid.Nil, commander, []uuid.UUID{clone}, v, game.IndefiniteDuration(), "test copy") {
			t.Fatal("the copy did not happen")
		}
	})
	// The copy is a Merfolk under the original's anthem: 2/2 and +1/+1.
	if p, tg := effectivePower(t, g, clone), effectiveToughness(t, g, clone); p != 3 || tg != 3 {
		t.Errorf("a copy with no level counters is %d/%d, want 3/3 (a 2/2 under the original's anthem)", p, tg)
	}
	if hasEffectiveKeyword(t, g, clone, "flying") {
		t.Error("a copy with no level counters has flying")
	}
	coralhelmLevelUp(t, g, me, clone)
	coralhelmLevelUp(t, g, me, clone)
	if p, tg := effectivePower(t, g, clone), effectiveToughness(t, g, clone); p != 4 || tg != 4 {
		t.Errorf("a level 2 copy is %d/%d, want 4/4 (a 3/3 under the original's anthem)", p, tg)
	}
	if !hasEffectiveKeyword(t, g, clone, "flying") {
		t.Error("a level 2 copy has no flying")
	}
	// The original is still level 4 and pumps the other Merfolk once.
	coralhelmWant(t, g, commander, myFish, 4, 4, true, 2)
}
