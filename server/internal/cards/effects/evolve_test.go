package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// evolve_test.go — #1805, CR 702.100 (ADR 0106 §3). Evolve is a
// canonical keyword token the engine turns into one trigger per
// instance (game/evolve.go). These tests hold the rule's edges — what
// triggers it, the CR 603.4 re-check on resolution, last-known
// information, instances, grants and the CR 702.100b "evolves" event —
// plus one per card that lands with it.

const (
	oracleTyranidPrime    = "12a43da3-57aa-4394-8cce-e53ec7e3538e"
	oraclePropagatorDrone = "9485db26-7e03-4a66-beeb-627a3bc6f367"
	oracleRenegadeKrasis  = "7b7e622c-815a-44a0-88ff-62bbe8dd5582"
	oracleWatchfulRadstag = "572afa0f-9e82-4899-b622-a7554625b9f4"
	oracleDinosaurEgg     = "ac4d5a97-6177-4eac-b0f2-cf10531ad879"
	oracleScurryOak       = "eee6c02b-7d6a-4445-ad27-8b03176147d4"
	oracleLonis           = "52d8e492-b7b7-4ad4-8f1f-5c50993e55e0"
	oracleGyreSage        = "e3aa16cf-079b-4737-9ffd-7bfdffef0cb2"
	oracleSimicFluxmage   = "68af451d-82f9-4c78-8bb1-36503e3f2e34"
	oraclePollywog        = "99a3cea9-f4ab-4da3-a085-c09dc93fc8cb"
)

// pushEvolveCreature is a creature with no catalog entry whose printed
// evolve the deck importer stamped — Cloudfin Raptor's case, and the
// proof that a vanilla evolve creature needs no card file.
func pushEvolveCreature(g *game.Game, controller uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	if len(keywords) == 0 {
		keywords = []string{game.KeywordEvolve}
	}
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird Mutant",
		Power: power, Toughness: toughness, Keywords: keywords,
		Owner: controller, Controller: controller,
	})
}

// enterCreature puts a creature into owner's hand and moves it onto the
// battlefield through the real entry path, which harvests its ETB
// triggers onto the stack before it returns.
func enterCreature(t *testing.T, g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	t.Helper()
	return enterCard(t, g, owner, game.Card{Name: name, TypeLine: "Creature — Bear", Power: power, Toughness: toughness})
}

func enterCard(t *testing.T, g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	playerByIDForTest(g, owner).Hand.PushTop(c)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner},
		game.ZoneRef{Kind: game.ZoneBattlefield},
		c.InstanceID,
	); err != nil {
		t.Fatalf("enter %s: %v", c.Name, err)
	}
	return c.InstanceID
}

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

func plusOnes(g *game.Game, id uuid.UUID) int { return countersOn(g, id, game.CounterPlusOne) }

// TestEvolveGrowsOnGreaterPowerOrToughness — CR 702.100a on a creature
// with no catalog entry: a creature with greater power, or greater
// toughness, puts ONE trigger on the stack, the counter waits for it to
// resolve, and an equal or smaller creature triggers nothing.
func TestEvolveGrowsOnGreaterPowerOrToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := pushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1, "flying", game.KeywordEvolve)

	bear := enterCreature(t, g, me, "Grizzly Bears", 1, 1)
	if got := evolveItemsFrom(g, raptor); got != 1 {
		t.Fatalf("a 1/1 entering beside a 0/1 put %d evolve triggers on the stack, want 1", got)
	}
	if item := triggerOnStack(g, raptor); item == nil || !strings.Contains(item.Label, "Grizzly Bears") {
		t.Errorf("the evolve trigger's label does not name the creature that entered: %+v", item)
	}
	if n := plusOnes(g, raptor); n != 0 {
		t.Errorf("%d counters before the trigger resolved, want 0 (a trigger, not a static)", n)
	}
	settleProwess(t, g)
	if n := plusOnes(g, raptor); n != 1 {
		t.Fatalf("after greater power: %d +1/+1 counters, want 1", n)
	}
	if got := evolvedEvents(g, raptor); len(got) != 1 || got[0] != 1 {
		t.Errorf("EventEvolved amounts = %v, want [1]", got)
	}
	_ = bear

	// 1/2 now. A 1/1 is neither stronger nor tougher.
	enterCreature(t, g, me, "Squire", 1, 1)
	if got := evolveItemsFrom(g, raptor); got != 0 {
		t.Errorf("a 1/1 entering beside a 1/2 put %d evolve triggers on the stack, want 0", got)
	}
	settleProwess(t, g)

	// Greater toughness alone is enough.
	enterCreature(t, g, me, "Wall of Wood", 0, 3)
	if got := evolveItemsFrom(g, raptor); got != 1 {
		t.Fatalf("a 0/3 entering beside a 1/2 put %d evolve triggers on the stack, want 1", got)
	}
	settleProwess(t, g)
	if n := plusOnes(g, raptor); n != 2 {
		t.Errorf("after greater toughness: %d counters, want 2", n)
	}
}

// TestEvolveIgnoresOpponentsCreaturesAndNoncreatures — "a creature YOU
// CONTROL enters": an opponent's creature is not one, and a noncreature
// permanent can't be greater (CR 702.100c).
func TestEvolveIgnoresOpponentsCreaturesAndNoncreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	raptor := pushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)
	theirs := pushEvolveCreature(g, opp, "Their Raptor", 0, 1)

	enterCreature(t, g, opp, "Their Ogre", 3, 3)
	if got := evolveItemsFrom(g, raptor); got != 0 {
		t.Errorf("an opponent's creature triggered my evolve %d times", got)
	}
	if got := evolveItemsFrom(g, theirs); got != 1 {
		t.Errorf("the opponent's own evolve triggered %d times on their creature, want 1", got)
	}
	settleProwess(t, g)

	enterCard(t, g, me, game.Card{Name: "Ornithopter Statue", TypeLine: "Artifact", Power: 5, Toughness: 5})
	if got := evolveItemsFrom(g, raptor); got != 0 {
		t.Errorf("a noncreature artifact triggered evolve %d times", got)
	}
	settleProwess(t, g)
	if n := plusOnes(g, raptor); n != 0 {
		t.Errorf("my Raptor has %d counters, want 0", n)
	}
}

// TestEvolveRechecksTheComparisonOnResolution — CR 603.4: the condition
// is checked again as the trigger resolves, and false then means
// nothing happens, and nothing evolved.
func TestEvolveRechecksTheComparisonOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := pushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)

	enterCreature(t, g, me, "Grizzly Bears", 2, 2)
	if got := evolveItemsFrom(g, raptor); got != 1 {
		t.Fatalf("%d evolve triggers, want 1", got)
	}
	// In response the Raptor grows to 2/3: no longer smaller in either.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(raptor, game.CounterPlusOne, 2) })
	settleProwess(t, g)
	if n := plusOnes(g, raptor); n != 2 {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := pushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)

	ogre := enterCreature(t, g, me, "Hill Giant", 3, 3)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ogre) })
	settleProwess(t, g)
	if n := plusOnes(g, raptor); n != 1 {
		t.Errorf("%d counters after the entering creature died in response, want 1", n)
	}
}

// TestEvolveDoesNothingWhenTheEvolvingCreatureLeft — a counter can't be
// put on last-known information, and a creature that left and came back
// is a new object (CR 400.7).
func TestEvolveDoesNothingWhenTheEvolvingCreatureLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := pushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)

	enterCreature(t, g, me, "Grizzly Bears", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(raptor) })
	settleProwess(t, g)
	if got := evolvedEvents(g, raptor); len(got) != 0 {
		t.Errorf("a creature that left evolved: %v", got)
	}
}

// TestEvolveSeesCreaturesEnteringTogether — CR 603.6a: two creatures
// entering at once are two triggers, and each re-checks on its own, so
// the second one sees the counter the first one put on.
func TestEvolveSeesCreaturesEnteringTogether(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := pushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)
	p := playerByIDForTest(g, me)
	a, b := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{a, b} {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Squire", TypeLine: "Creature — Human Soldier",
			Power: 1, Toughness: 1, Owner: me, Controller: me})
	}
	g.WithWriteLock(func() {
		if err := g.PutOntoBattlefieldTogetherThenForEffect(
			[]game.BatchEntry{{CardID: a, From: game.ZoneHand}, {CardID: b, From: game.ZoneHand}},
			game.ZoneEntryOptions{}, nil); err != nil {
			t.Fatalf("enter together: %v", err)
		}
	})
	if got := evolveItemsFrom(g, raptor); got != 2 {
		t.Fatalf("two 1/1s entering together put %d evolve triggers, want 2", got)
	}
	settleProwess(t, g)
	// The first makes the Raptor 1/2; the second 1/1 is then neither
	// stronger nor tougher.
	if n := plusOnes(g, raptor); n != 1 {
		t.Errorf("%d counters, want 1 (the second trigger's re-check fails)", n)
	}
}

// TestMultipleInstancesOfEvolveTriggerSeparately — CR 702.100d, through
// Tyranid Prime's grant: a Raptor that prints evolve has two instances
// under the Prime, both trigger, and each re-checks against the Raptor
// as the other left it.
func TestMultipleInstancesOfEvolveTriggerSeparately(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := pushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)
	prime := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tyranid Prime", OracleID: oracleTyranidPrime,
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

	enterCreature(t, g, me, "Grizzly Bears", 2, 2)
	if got := evolveItemsFrom(g, raptor); got != 2 {
		t.Fatalf("the Raptor's two instances put %d triggers on the stack, want 2", got)
	}
	settleProwess(t, g)
	// 0/1 → 1/2 (power 2 > 0) → 2/3 (power 2 > 1).
	if n := plusOnes(g, raptor); n != 2 {
		t.Errorf("Raptor: %d counters, want 2", n)
	}
	if n := plusOnes(g, prime); n != 1 {
		t.Errorf("Tyranid Prime: %d counters, want 1", n)
	}
}

// TestPropagatorDroneGivesCreatureTokensEvolve — a granted evolve on a
// token, which also sees the Drone itself enter (its reminder text).
func TestPropagatorDroneGivesCreatureTokensEvolve(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	spawn := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Eldrazi Spawn", TypeLine: "Token Creature — Eldrazi Spawn",
		Power: 0, Toughness: 1, Owner: me, Controller: me,
	})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 0, Toughness: 1, Owner: me, Controller: me,
	})
	drone := enterCard(t, g, me, game.Card{Name: "Propagator Drone", OracleID: oraclePropagatorDrone,
		TypeLine: "Creature — Eldrazi Drone", Power: 2, Toughness: 2})
	if got := evolveItemsFrom(g, spawn); got != 1 {
		t.Fatalf("the Spawn saw the Drone enter %d times, want 1", got)
	}
	if got := evolveItemsFrom(g, bear) + evolveItemsFrom(g, drone); got != 0 {
		t.Errorf("a nontoken creature got evolve from the Drone (%d triggers)", got)
	}
	settleProwess(t, g)
	if n := plusOnes(g, spawn); n != 1 {
		t.Errorf("Spawn: %d counters, want 1", n)
	}
}

// TestHardenedScalesMakesEvolveTwoCountersAndOneEvolution — the counter
// goes through the CR 614 window, and EventEvolved carries what landed.
func TestHardenedScalesMakesEvolveTwoCountersAndOneEvolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me)
	raptor := pushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1)

	enterCreature(t, g, me, "Grizzly Bears", 2, 2)
	settleProwess(t, g)
	if n := plusOnes(g, raptor); n != 2 {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	krasis := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Renegade Krasis", OracleID: oracleRenegadeKrasis,
		TypeLine: "Creature — Beast Mutant", Power: 3, Toughness: 2, Owner: me, Controller: me,
	})
	grown := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grown", TypeLine: "Creature — Elf", Power: 9, Toughness: 9,
		Owner: me, Controller: me,
	})
	plain := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain", TypeLine: "Creature — Elf", Power: 9, Toughness: 9,
		Owner: me, Controller: me,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(grown, game.CounterPlusOne, 1) })
	// A counter from something else is not an evolution.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(krasis, game.CounterPlusOne, 1) })
	settleProwess(t, g)
	if n := plusOnes(g, grown); n != 1 {
		t.Fatalf("a plain counter on the Krasis triggered it: Grown has %d", n)
	}

	// 4/3 now; a 5/1 is stronger.
	enterCreature(t, g, me, "Hasty Ogre", 5, 1)
	settleProwess(t, g)
	if n := plusOnes(g, krasis); n != 2 {
		t.Fatalf("Krasis: %d counters, want 2", n)
	}
	if n := plusOnes(g, grown); n != 2 {
		t.Errorf("the other creature with a counter: %d, want 2", n)
	}
	if n := plusOnes(g, plain); n != 0 {
		t.Errorf("a creature with no counter got %d", n)
	}
}

// TestWatchfulRadstagCopiesItselfWhenItEvolves.
func TestWatchfulRadstagCopiesItselfWhenItEvolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Watchful Radstag", OracleID: oracleWatchfulRadstag,
		TypeLine: "Creature — Elk Mutant", Power: 2, Toughness: 2, Owner: me, Controller: me,
	})
	enterCreature(t, g, me, "Hill Giant", 3, 3)
	settleProwess(t, g)
	if n := onBattlefieldNamed(g, "Watchful Radstag"); n != 2 {
		t.Errorf("%d Watchful Radstags after it evolved, want 2 (it and its token copy)", n)
	}
}

// TestDinosaurEggEvolves — the named card (ADR 0106): its evolve now
// works, and the discover reads the grown toughness.
func TestDinosaurEggEvolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	egg := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dinosaur Egg", OracleID: oracleDinosaurEgg,
		TypeLine: "Creature — Dinosaur Egg", Power: 0, Toughness: 3, Owner: me, Controller: me,
	})
	enterCreature(t, g, me, "Grizzly Bears", 1, 1)
	settleProwess(t, g)
	if n := plusOnes(g, egg); n != 1 {
		t.Errorf("Dinosaur Egg: %d counters after a 1/1 entered, want 1", n)
	}
}

// TestScurryOakMakesASquirrelWhenItGrows — "one or more +1/+1 counters
// are put on this creature", here by its own evolve.
func TestScurryOakMakesASquirrelWhenItGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Scurry Oak", OracleID: oracleScurryOak,
		TypeLine: "Creature — Treefolk", Power: 1, Toughness: 2, Owner: me, Controller: me,
	})
	enterCreature(t, g, me, "Grizzly Bears", 2, 2)
	settleProwess(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("the Oak's counter did not ask about a Squirrel")
	}
	if err := g.ResolveConfirm(ask.ID, me, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	settleProwess(t, g)
	if n := onBattlefieldNamed(g, "Squirrel"); n != 1 {
		t.Errorf("%d Squirrels after the Oak evolved, want 1", n)
	}
}

// TestLonisInvestigatesOncePerCounterPlaced — "that many times".
func TestLonisInvestigatesOncePerCounterPlaced(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lonis, Genetics Expert", OracleID: oracleLonis,
		TypeLine: "Legendary Creature — Snake Elf Detective", Power: 1, Toughness: 2, Owner: me, Controller: me,
	})
	enterCreature(t, g, me, "Grizzly Bears", 2, 2)
	settleProwess(t, g)
	if n := onBattlefieldNamed(g, "Clue"); n != 2 {
		t.Errorf("%d Clues after Lonis evolved under Hardened Scales, want 2", n)
	}
}

// TestGyreSageTapsForAGreenPerCounter.
func TestGyreSageTapsForAGreenPerCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	sage := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gyre Sage", OracleID: oracleGyreSage,
		TypeLine: "Creature — Elf Druid", Power: 1, Toughness: 2, Owner: me, Controller: me,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sage, game.CounterPlusOne, 3) })
	var produced string
	g.WithWriteLock(func() { produced = gyreSageProduced(g, me, sage) })
	if produced != "{G}{G}{G}" {
		t.Errorf("Gyre Sage with three counters produces %q, want {G}{G}{G}", produced)
	}
}

// TestSimicFluxmageMovesACounter — CR 122.5: off the Fluxmage, onto the
// target; with no counter to move, nothing happens.
func TestSimicFluxmageMovesACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	mage := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Simic Fluxmage", OracleID: oracleSimicFluxmage,
		TypeLine: "Creature — Merfolk Wizard", Power: 1, Toughness: 2, Owner: me, Controller: me,
	})
	other := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me, Controller: me,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mage, game.CounterPlusOne, 1) })
	item := &game.StackItem{
		Kind: game.StackItemActivated, Controller: me, Owner: me, SourceCardID: mage,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}},
	}
	g.WithWriteLock(func() {
		if err := simicFluxmageMove(g, item); err != nil {
			t.Fatalf("move: %v", err)
		}
	})
	if a, b := plusOnes(g, mage), plusOnes(g, other); a != 0 || b != 1 {
		t.Errorf("after the move: Fluxmage %d, target %d; want 0, 1", a, b)
	}
	g.WithWriteLock(func() { _ = simicFluxmageMove(g, item) })
	if a, b := plusOnes(g, mage), plusOnes(g, other); a != 0 || b != 1 {
		t.Errorf("a move with no counter to move changed something: Fluxmage %d, target %d", a, b)
	}
}

// TestPollywogProdigyDrawsOnOpponentsCheapNoncreatureSpells — mana value
// LESS than its power, an opponent's, and not a creature.
func TestPollywogProdigyDrawsOnOpponentsCheapNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	prodigy := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pollywog Prodigy", OracleID: oraclePollywog,
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
			got = pollywogProdigyApplies(game.Event{Kind: game.EventCast, Actor: me.ID, CardID: id}, &src, game.Characteristic{}, g)
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
		mine = pollywogProdigyApplies(game.Event{Kind: game.EventCast, Actor: opp.ID, CardID: uuid.New()}, &src, game.Characteristic{}, g)
	})
	if mine {
		t.Error("its controller's own spell triggered it")
	}
}
