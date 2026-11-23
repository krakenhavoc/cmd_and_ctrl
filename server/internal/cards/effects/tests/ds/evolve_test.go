package testsds

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsgr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/gr"
	cardsp "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/p"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// evolveItemsFrom counts the evolve triggers from `source` on the
// stack or waiting to go there.
func evolveItemsFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source && item.Body == "evolve/grow" {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source && item.Body == "evolve/grow" {
			n++
		}
	}
	return n
}

func evolvedEvents(g *game.Game, id uuid.UUID) []int {
	var out []int
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventEvolved && ev.CardID == id {
				out = append(out, ev.Amount)
			}
		}
	})
	return out
}

// TestEvolveGrowsOnGreaterPowerOrToughness — CR 702.100a on a creature
// with no catalog entry: a creature with greater power, or greater
// toughness, puts ONE trigger on the stack, the counter waits for it to
// resolve, and an equal or smaller creature triggers nothing.
func TestEvolveGrowsOnGreaterPowerOrToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := effectstest.PushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1, "flying", game.KeywordEvolve)

	bear := effectstest.EnterCreature(t, g, me, "Grizzly Bears", 1, 1)
	if got := evolveItemsFrom(g, raptor); got != 1 {
		t.Fatalf("a 1/1 entering beside a 0/1 put %d evolve triggers on the stack, want 1", got)
	}
	if item := effectstest.TriggerOnStack(g, raptor); item == nil || !strings.Contains(item.Label, "Grizzly Bears") {
		t.Errorf("the evolve trigger's label does not name the creature that entered: %+v", item)
	}
	if n := effectstest.PlusOnes(g, raptor); n != 0 {
		t.Errorf("%d counters before the trigger resolved, want 0 (a trigger, not a static)", n)
	}
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, raptor); n != 1 {
		t.Fatalf("after greater power: %d +1/+1 counters, want 1", n)
	}
	if got := evolvedEvents(g, raptor); len(got) != 1 || got[0] != 1 {
		t.Errorf("EventEvolved amounts = %v, want [1]", got)
	}
	_ = bear

	// 1/2 now. A 1/1 is neither stronger nor tougher.
	effectstest.EnterCreature(t, g, me, "Squire", 1, 1)
	if got := evolveItemsFrom(g, raptor); got != 0 {
		t.Errorf("a 1/1 entering beside a 1/2 put %d evolve triggers on the stack, want 0", got)
	}
	effectstest.SettleProwess(t, g)

	// Greater toughness alone is enough.
	effectstest.EnterCreature(t, g, me, "Wall of Wood", 0, 3)
	if got := evolveItemsFrom(g, raptor); got != 1 {
		t.Fatalf("a 0/3 entering beside a 1/2 put %d evolve triggers on the stack, want 1", got)
	}
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, raptor); n != 2 {
		t.Errorf("after greater toughness: %d counters, want 2", n)
	}
}

// TestEvolveIgnoresOpponentsCreaturesAndNoncreatures — "a creature YOU
// CONTROL enters": an opponent's creature is not one, and a noncreature
// permanent can't be greater (CR 702.100c).
func TestEvolveIgnoresOpponentsCreaturesAndNoncreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	raptor := effectstest.PushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)
	theirs := effectstest.PushEvolveCreature(g, opp, "Their Raptor", 0, 1)

	effectstest.EnterCreature(t, g, opp, "Their Ogre", 3, 3)
	if got := evolveItemsFrom(g, raptor); got != 0 {
		t.Errorf("an opponent's creature triggered my evolve %d times", got)
	}
	if got := evolveItemsFrom(g, theirs); got != 1 {
		t.Errorf("the opponent's own evolve triggered %d times on their creature, want 1", got)
	}
	effectstest.SettleProwess(t, g)

	effectstest.EnterCard(t, g, me, game.Card{Name: "Ornithopter Statue", TypeLine: "Artifact", Power: 5, Toughness: 5})
	if got := evolveItemsFrom(g, raptor); got != 0 {
		t.Errorf("a noncreature artifact triggered evolve %d times", got)
	}
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, raptor); n != 0 {
		t.Errorf("my Raptor has %d counters, want 0", n)
	}
}

// TestEvolveRechecksTheComparisonOnResolution — CR 603.4: the condition
// is checked again as the trigger resolves, and false then means
// nothing happens, and nothing evolved.
func TestEvolveRechecksTheComparisonOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := effectstest.PushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)

	effectstest.EnterCreature(t, g, me, "Grizzly Bears", 2, 2)
	if got := evolveItemsFrom(g, raptor); got != 1 {
		t.Fatalf("%d evolve triggers, want 1", got)
	}
	// In response the Raptor grows to 2/3: no longer smaller in either.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(raptor, game.CounterPlusOne, 2) })
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, raptor); n != 2 {
		t.Errorf("%d counters after a failed re-check, want the 2 placed in response", n)
	}
	if got := evolvedEvents(g, raptor); len(got) != 0 {
		t.Errorf("EventEvolved %v after a failed re-check, want none", got)
	}
}

// TestEvolveReadsALeftCreatureAsItLastExisted — CR 608.2h: the creature
// that entered is compared by its last-known power and toughness once
// it has left.
func TestEvolveReadsALeftCreatureAsItLastExisted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := effectstest.PushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)

	ogre := effectstest.EnterCreature(t, g, me, "Hill Giant", 3, 3)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ogre) })
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, raptor); n != 1 {
		t.Errorf("%d counters after the entering creature died in response, want 1", n)
	}
}

// TestEvolveDoesNothingWhenTheEvolvingCreatureLeft — a counter can't be
// put on last-known information, and a creature that left and came back
// is a new object (CR 400.7).
func TestEvolveDoesNothingWhenTheEvolvingCreatureLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := effectstest.PushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)

	effectstest.EnterCreature(t, g, me, "Grizzly Bears", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(raptor) })
	effectstest.SettleProwess(t, g)
	if got := evolvedEvents(g, raptor); len(got) != 0 {
		t.Errorf("a creature that left evolved: %v", got)
	}
}

// TestEvolveSeesCreaturesEnteringTogether — CR 603.6a: two creatures
// entering at once are two triggers, and each re-checks on its own, so
// the second one sees the counter the first one put on.
func TestEvolveSeesCreaturesEnteringTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := effectstest.PushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)
	p := effectstest.PlayerByIDForTest(g, me)
	a, b := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{a, b} {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Squire", TypeLine: "Creature — Human Soldier",
			Power: 1, Toughness: 1, Owner: me, Controller: me})
	}
	g.WithWriteLock(func() {
		if err := g.PutOntoBattlefieldTogetherKeyedForEffect(
			[]game.BatchEntry{{CardID: a, From: game.ZoneHand}, {CardID: b, From: game.ZoneHand}},
			game.ZoneEntryOptions{}, game.KeyedTail{}); err != nil {
			t.Fatalf("enter together: %v", err)
		}
	})
	if got := evolveItemsFrom(g, raptor); got != 2 {
		t.Fatalf("two 1/1s entering together put %d evolve triggers, want 2", got)
	}
	effectstest.SettleProwess(t, g)
	// The first makes the Raptor 1/2; the second 1/1 is then neither
	// stronger nor tougher.
	if n := effectstest.PlusOnes(g, raptor); n != 1 {
		t.Errorf("%d counters, want 1 (the second trigger's re-check fails)", n)
	}
}

// TestMultipleInstancesOfEvolveTriggerSeparately — CR 702.100d, through
// Tyranid Prime's grant: a Raptor that prints evolve has two instances
// under the Prime, both trigger, and each re-checks against the Raptor
// as the other left it.
func TestMultipleInstancesOfEvolveTriggerSeparately(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := effectstest.PushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)
	prime := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tyranid Prime", OracleID: effectstest.OracleTyranidPrime,
		TypeLine: "Creature — Tyranid", Power: 0, Toughness: 4, Owner: me, Controller: me,
	})
	var raptorN, primeN int
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case raptor:
				raptorN = game.EvolveCount(&g.Battlefield.Cards[i])
			case prime:
				primeN = game.EvolveCount(&g.Battlefield.Cards[i])
			}
		}
	})
	if raptorN != 2 || primeN != 1 {
		t.Fatalf("instances: Raptor %d (want 2, printed + granted), Prime %d (want 1, \"other\")", raptorN, primeN)
	}

	effectstest.EnterCreature(t, g, me, "Grizzly Bears", 2, 2)
	if got := evolveItemsFrom(g, raptor); got != 2 {
		t.Fatalf("the Raptor's two instances put %d triggers on the stack, want 2", got)
	}
	effectstest.SettleProwess(t, g)
	// 0/1 → 1/2 (power 2 > 0) → 2/3 (power 2 > 1).
	if n := effectstest.PlusOnes(g, raptor); n != 2 {
		t.Errorf("Raptor: %d counters, want 2", n)
	}
	if n := effectstest.PlusOnes(g, prime); n != 1 {
		t.Errorf("Tyranid Prime: %d counters, want 1", n)
	}
}

// TestPropagatorDroneGivesCreatureTokensEvolve — a granted evolve on a
// token, which also sees the Drone itself enter (its reminder text).
func TestPropagatorDroneGivesCreatureTokensEvolve(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	spawn := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Eldrazi Spawn", TypeLine: "Token Creature — Eldrazi Spawn",
		Power: 0, Toughness: 1, Owner: me, Controller: me,
	})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 0, Toughness: 1, Owner: me, Controller: me,
	})
	drone := effectstest.EnterCard(t, g, me, game.Card{Name: "Propagator Drone", OracleID: effectstest.OraclePropagatorDrone,
		TypeLine: "Creature — Eldrazi Drone", Power: 2, Toughness: 2})
	if got := evolveItemsFrom(g, spawn); got != 1 {
		t.Fatalf("the Spawn saw the Drone enter %d times, want 1", got)
	}
	if got := evolveItemsFrom(g, bear) + evolveItemsFrom(g, drone); got != 0 {
		t.Errorf("a nontoken creature got evolve from the Drone (%d triggers)", got)
	}
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, spawn); n != 1 {
		t.Errorf("Spawn: %d counters, want 1", n)
	}
}

// TestHardenedScalesMakesEvolveTwoCountersAndOneEvolution — the counter
// goes through the CR 614 window, and EventEvolved carries what landed.
func TestHardenedScalesMakesEvolveTwoCountersAndOneEvolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me)
	raptor := effectstest.PushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)

	effectstest.EnterCreature(t, g, me, "Grizzly Bears", 2, 2)
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, raptor); n != 2 {
		t.Errorf("%d counters under Hardened Scales, want 2", n)
	}
	if got := evolvedEvents(g, raptor); len(got) != 1 || got[0] != 2 {
		t.Errorf("EventEvolved amounts = %v, want [2]", got)
	}
}

// TestRenegadeKrasisSpreadsCountersWhenItEvolves — CR 702.100b: an
// evolution, and only an evolution, triggers it; "each other" creature
// with a +1/+1 counter gets one.
func TestRenegadeKrasisSpreadsCountersWhenItEvolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	krasis := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Renegade Krasis", OracleID: effectstest.OracleRenegadeKrasis,
		TypeLine: "Creature — Beast Mutant", Power: 3, Toughness: 2, Owner: me, Controller: me,
	})
	grown := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grown", TypeLine: "Creature — Elf", Power: 9, Toughness: 9,
		Owner: me, Controller: me,
	})
	plain := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain", TypeLine: "Creature — Elf", Power: 9, Toughness: 9,
		Owner: me, Controller: me,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(grown, game.CounterPlusOne, 1) })
	// A counter from something else is not an evolution.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(krasis, game.CounterPlusOne, 1) })
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, grown); n != 1 {
		t.Fatalf("a plain counter on the Krasis triggered it: Grown has %d", n)
	}

	// 4/3 now; a 5/1 is stronger.
	effectstest.EnterCreature(t, g, me, "Hasty Ogre", 5, 1)
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, krasis); n != 2 {
		t.Fatalf("Krasis: %d counters, want 2", n)
	}
	if n := effectstest.PlusOnes(g, grown); n != 2 {
		t.Errorf("the other creature with a counter: %d, want 2", n)
	}
	if n := effectstest.PlusOnes(g, plain); n != 0 {
		t.Errorf("a creature with no counter got %d", n)
	}
}

// TestWatchfulRadstagCopiesItselfWhenItEvolves.
func TestWatchfulRadstagCopiesItselfWhenItEvolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Watchful Radstag", OracleID: effectstest.OracleWatchfulRadstag,
		TypeLine: "Creature — Elk Mutant", Power: 2, Toughness: 2, Owner: me, Controller: me,
	})
	effectstest.EnterCreature(t, g, me, "Hill Giant", 3, 3)
	effectstest.SettleProwess(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Watchful Radstag"); n != 2 {
		t.Errorf("%d Watchful Radstags after it evolved, want 2 (it and its token copy)", n)
	}
}

// TestDinosaurEggEvolves — the named card (ADR 0106): its evolve now
// works, and the discover reads the grown toughness.
func TestDinosaurEggEvolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	egg := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dinosaur Egg", OracleID: effectstest.OracleDinosaurEgg,
		TypeLine: "Creature — Dinosaur Egg", Power: 0, Toughness: 3, Owner: me, Controller: me,
	})
	effectstest.EnterCreature(t, g, me, "Grizzly Bears", 1, 1)
	effectstest.SettleProwess(t, g)
	if n := effectstest.PlusOnes(g, egg); n != 1 {
		t.Errorf("Dinosaur Egg: %d counters after a 1/1 entered, want 1", n)
	}
}

// TestScurryOakMakesASquirrelWhenItGrows — "one or more +1/+1 counters
// are put on this creature", here by its own evolve.
func TestScurryOakMakesASquirrelWhenItGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Scurry Oak", OracleID: effectstest.OracleScurryOak,
		TypeLine: "Creature — Treefolk", Power: 1, Toughness: 2, Owner: me, Controller: me,
	})
	effectstest.EnterCreature(t, g, me, "Grizzly Bears", 2, 2)
	effectstest.SettleProwess(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("the Oak's counter did not ask about a Squirrel")
	}
	if err := g.ResolveConfirm(ask.ID, me, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.SettleProwess(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Squirrel"); n != 1 {
		t.Errorf("%d Squirrels after the Oak evolved, want 1", n)
	}
}

// TestLonisInvestigatesOncePerCounterPlaced — "that many times".
func TestLonisInvestigatesOncePerCounterPlaced(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lonis, Genetics Expert", OracleID: effectstest.OracleLonis,
		TypeLine: "Legendary Creature — Snake Elf Detective", Power: 1, Toughness: 2, Owner: me, Controller: me,
	})
	effectstest.EnterCreature(t, g, me, "Grizzly Bears", 2, 2)
	effectstest.SettleProwess(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Clue"); n != 2 {
		t.Errorf("%d Clues after Lonis evolved under Hardened Scales, want 2", n)
	}
}

// TestGyreSageTapsForAGreenPerCounter.
func TestGyreSageTapsForAGreenPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	sage := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gyre Sage", OracleID: effectstest.OracleGyreSage,
		TypeLine: "Creature — Elf Druid", Power: 1, Toughness: 2, Owner: me, Controller: me,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sage, game.CounterPlusOne, 3) })
	var produced game.ManaAmount
	g.WithWriteLock(func() { produced = cardsgr.GyreSageProduced(g, me, sage) })
	if produced.N != 3 || len(produced.Options) != 1 || produced.Options[0] != "G" {
		t.Errorf("Gyre Sage with three counters produces %+v, want three {G}", produced)
	}
}

// TestSimicFluxmageMovesACounter — CR 122.5: off the Fluxmage, onto the
// target; with no counter to move, nothing happens.
func TestSimicFluxmageMovesACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	mage := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Simic Fluxmage", OracleID: effectstest.OracleSimicFluxmage,
		TypeLine: "Creature — Merfolk Wizard", Power: 1, Toughness: 2, Owner: me, Controller: me,
	})
	other := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me, Controller: me,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mage, game.CounterPlusOne, 1) })
	item := &game.StackItem{
		Kind: game.StackItemActivated, Controller: me, Owner: me, SourceCardID: mage,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}},
	}
	g.WithWriteLock(func() {
		if err := effects.SimicFluxmageMove(g, item); err != nil {
			t.Fatalf("move: %v", err)
		}
	})
	if a, b := effectstest.PlusOnes(g, mage), effectstest.PlusOnes(g, other); a != 0 || b != 1 {
		t.Errorf("after the move: Fluxmage %d, target %d; want 0, 1", a, b)
	}
	g.WithWriteLock(func() { _ = effects.SimicFluxmageMove(g, item) })
	if a, b := effectstest.PlusOnes(g, mage), effectstest.PlusOnes(g, other); a != 0 || b != 1 {
		t.Errorf("a move with no counter to move changed something: Fluxmage %d, target %d", a, b)
	}
}

// TestPollywogProdigyDrawsOnOpponentsCheapNoncreatureSpells — mana value
// LESS than its power, an opponent's, and not a creature.
func TestPollywogProdigyDrawsOnOpponentsCheapNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	prodigy := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pollywog Prodigy", OracleID: effectstest.OraclePollywog,
		TypeLine: "Creature — Frog Wizard", Power: 1, Toughness: 3, Owner: opp.ID, Controller: opp.ID,
	})
	var src game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == prodigy {
				src = c
			}
		}
	})
	ok := func(name, typeLine, cost string) bool {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		var got bool
		g.WithWriteLock(func() {
			got = cardsp.PollywogProdigyApplies(game.Event{Kind: game.EventCast, Actor: me.ID, CardID: id}, &src, game.Characteristic{}, g)
		})
		return got
	}
	if !ok("Free Spell", "Instant", "{0}") {
		t.Error("a mana value 0 instant from an opponent did not trigger a power-1 Prodigy")
	}
	if ok("One Drop", "Instant", "{U}") {
		t.Error("mana value 1 is not less than power 1")
	}
	if ok("Ornithopter", "Artifact Creature — Thopter", "{0}") {
		t.Error("a creature spell triggered it")
	}
	var mine bool
	g.WithWriteLock(func() {
		mine = cardsp.PollywogProdigyApplies(game.Event{Kind: game.EventCast, Actor: opp.ID, CardID: uuid.New()}, &src, game.Characteristic{}, g)
	})
	if mine {
		t.Error("its controller's own spell triggered it")
	}
}
