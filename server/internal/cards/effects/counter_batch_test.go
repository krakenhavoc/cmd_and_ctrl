package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_batch_test.go — #2150: "whenever you put one or more
// counters on ~" with the kinds put together, through Aragorn,
// Company Leader.

const aragornCompanyLeaderOracle = "1b841e8f-9484-4737-bd37-dfa11bd06883"

func aragornBoard(t *testing.T) (g *game.Game, me, opp *game.Player, aragorn, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	aragorn = pushCatalogPermanent(g, me.ID, "Aragorn, Company Leader", "Legendary Creature — Human Ranger", aragornCompanyLeaderOracle, false)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	return g, me, opp, aragorn, bear
}

// putCountersAsAResolution is one resolution of `who`'s putting each
// (kind, n) pair on `id`: one event batch.
func putCountersAsAResolution(g *game.Game, who, id uuid.UUID, kinds map[string]int, order ...string) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: who})
		for _, k := range order {
			_ = g.AddCounterByForEffect(who, id, k, kinds[k])
		}
	})
}

func batchCounterOn(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	return ringBFCard(t, g, id).Counters[kind]
}

// One placement of three kinds is ONE trigger: one target, one of each
// kind, however many of that kind were put on Aragorn.
func TestAragornCopiesEveryKindOfOneBatchToOneTarget(t *testing.T) {
	g, me, _, aragorn, bear := aragornBoard(t)
	other := b12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)

	putCountersAsAResolution(g, me.ID, aragorn,
		map[string]int{game.CounterPlusOne: 2, game.CounterFlying: 1, game.CounterLifelink: 1},
		game.CounterPlusOne, game.CounterFlying, game.CounterLifelink)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); hasID(p.PickTargetCards, aragorn) {
		t.Error("Aragorn himself is not a legal target: it says other")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a second trigger asked for a second target: three kinds in one batch are one occurrence")
	}

	for _, kind := range []string{game.CounterPlusOne, game.CounterFlying, game.CounterLifelink} {
		if got := batchCounterOn(t, g, bear, kind); got != 1 {
			t.Errorf("Bear has %d %q counters, want exactly one of each kind", got, kind)
		}
	}
	if got := batchCounterOn(t, g, other, game.CounterPlusOne); got != 0 {
		t.Errorf("the other creature got counters: %d", got)
	}
	if got := batchCounterOn(t, g, aragorn, game.CounterPlusOne); got != 2 {
		t.Errorf("Aragorn lost counters: %d", got)
	}
}

// The kinds are what was put on Aragorn in THAT batch, by you: a
// removal in the same resolution and counters on other creatures are
// not in it.
func TestAragornReadsOnlyTheKindsPutOnItByYou(t *testing.T) {
	g, me, _, aragorn, bear := aragornBoard(t)
	other := b12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterByForEffect(me.ID, aragorn, "charge", 1)
	})
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if got := batchCounterOn(t, g, other, "charge"); got != 1 {
		t.Fatalf("setup: charge copied %d times", got)
	}

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterByForEffect(me.ID, aragorn, "charge", -1) // a removal
		_ = g.AddCounterByForEffect(me.ID, bear, game.CounterTrample, 1)
		_ = g.AddCounterByForEffect(me.ID, aragorn, game.CounterReach, 1)
	})
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if got := batchCounterOn(t, g, other, game.CounterReach); got != 1 {
		t.Errorf("the reach counter put on Aragorn was not copied: %d", got)
	}
	if got := batchCounterOn(t, g, other, game.CounterTrample); got != 0 {
		t.Errorf("a counter put on another creature was copied: %d", got)
	}
	if got := batchCounterOn(t, g, other, "charge"); got != 1 {
		t.Errorf("a removed counter was copied: charge = %d, want the earlier 1", got)
	}
}

// Two placements that are NOT simultaneous are two occurrences, each
// copying only its own kinds.
func TestAragornSeparateBatchesTriggerSeparately(t *testing.T) {
	g, me, _, aragorn, bear := aragornBoard(t)

	putCountersAsAResolution(g, me.ID, aragorn, map[string]int{game.CounterPlusOne: 1}, game.CounterPlusOne)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	putCountersAsAResolution(g, me.ID, aragorn, map[string]int{game.CounterFlying: 1}, game.CounterFlying)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if got := batchCounterOn(t, g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the first batch's kind was copied %d times, want once", got)
	}
	if got := batchCounterOn(t, g, bear, game.CounterFlying); got != 1 {
		t.Errorf("the second batch did not trigger on its own: flying = %d", got)
	}
}

func TestAragornIgnoresCountersAnOpponentPutsOnIt(t *testing.T) {
	g, _, opp, aragorn, _ := aragornBoard(t)
	putCountersAsAResolution(g, opp.ID, aragorn, map[string]int{game.CounterPlusOne: 1}, game.CounterPlusOne)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, g.Seats[0].ID) != nil {
		t.Error("an opponent's counter triggered it: it says you put")
	}
}

// The first ability: the Ring tempts you, you choose another creature,
// then put your choice of a counter on Aragorn. That counter is a
// counter you put on Aragorn, so it triggers the second ability.
func TestAragornRingTemptPutsAChosenCounterThatTriggersTheCopy(t *testing.T) {
	g, me, _, aragorn, bear := aragornBoard(t)

	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, aragorn)
	ringSettle(t, g)
	if latestOptionPickFor(g, me.ID) != nil {
		t.Fatal("choosing Aragorn himself triggers nothing")
	}

	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 3) // lifelink
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if got := batchCounterOn(t, g, aragorn, game.CounterLifelink); got != 1 {
		t.Errorf("Aragorn has %d lifelink counters, want 1", got)
	}
	if got := batchCounterOn(t, g, bear, game.CounterLifelink); got != 1 {
		t.Errorf("the other creature has %d lifelink counters, want 1", got)
	}
	if !effectiveHasKeyword(ringBFCard(t, g, bear), "lifelink") {
		t.Error("a lifelink counter is lifelink")
	}
}
