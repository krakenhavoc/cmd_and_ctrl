package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const fearOfSleepParalysisOracle = "a3a1f149-febe-4ff6-9e34-8b82ef21c2ee"

func stunnedTapped(g *game.Game, owner uuid.UUID, stun int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stunned", TypeLine: "Creature — Test",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		Tapped: true, Counters: map[string]int{game.CounterStun: stun},
	})
}

func fearOnBattlefield(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fear of Sleep Paralysis",
		TypeLine: "Enchantment Creature — Nightmare", OracleID: fearOfSleepParalysisOracle,
		Power: 6, Toughness: 6, Owner: controller, Controller: controller,
	})
}

// An opponent's stun counter is not removed by their untap step: the
// permanent stays tapped with the counter intact.
func TestFearOfSleepParalysisKeepsOpponentsStunCounterThroughUntap(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	fearOnBattlefield(g, me.ID)
	victim := stunnedTapped(g, opp.ID, 2)

	advanceToNextSeatsTurn(t, g) // the opponent's untap step runs
	c, _ := battlefieldCard(g, victim)
	if !c.Tapped || c.Counters[game.CounterStun] != 2 {
		t.Fatalf("locked permanent after its untap step: tapped=%v stun=%d, want tapped with 2 counters", c.Tapped, c.Counters[game.CounterStun])
	}
}

// The control: without the lock the same untap step removes one counter.
func TestStunCounterIsRemovedByUntapWithoutTheLock(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := g.Seats[1]
	victim := stunnedTapped(g, opp.ID, 2)

	advanceToNextSeatsTurn(t, g)
	c, _ := battlefieldCard(g, victim)
	if !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Fatalf("unlocked permanent after its untap step: tapped=%v stun=%d, want tapped with 1 counter", c.Tapped, c.Counters[game.CounterStun])
	}
}

// "Your opponents": the controller's own permanents are not covered.
func TestFearOfSleepParalysisDoesNotLockItsControllersOwnStunCounters(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := g.Seats[1]
	fearOnBattlefield(g, opp.ID)
	own := stunnedTapped(g, opp.ID, 1)

	advanceToNextSeatsTurn(t, g)
	c, _ := battlefieldCard(g, own)
	if c.Counters[game.CounterStun] != 0 || !c.Tapped {
		t.Fatalf("own permanent: tapped=%v stun=%d, want the counter removed in place of the untap", c.Tapped, c.Counters[game.CounterStun])
	}
}

// The lock is on the counter, not on the untap: a removal effect can't
// take it either, and other counter kinds are untouched.
func TestFearOfSleepParalysisBlocksEffectRemovalOfStunOnly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	fearOnBattlefield(g, me.ID)
	victim := stunnedTapped(g, opp.ID, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Counters["+1/+1"] = 2
			}
		}
		if err := g.AddCounterForEffect(victim, game.CounterStun, -1); err != nil {
			t.Fatalf("remove stun: %v", err)
		}
		if err := g.AddCounterForEffect(victim, "+1/+1", -1); err != nil {
			t.Fatalf("remove +1/+1: %v", err)
		}
	})
	c, _ := battlefieldCard(g, victim)
	if c.Counters[game.CounterStun] != 1 || c.Counters["+1/+1"] != 1 {
		t.Fatalf("counters = %v, want stun 1 (locked) and +1/+1 1 (removed)", c.Counters)
	}
}

// Losing the source lifts the lock with no bookkeeping.
func TestFearOfSleepParalysisLeavingLiftsTheLock(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	fear := fearOnBattlefield(g, me.ID)
	victim := stunnedTapped(g, opp.ID, 1)
	eerieDestroy(t, g, fear)

	advanceToNextSeatsTurn(t, g)
	c, _ := battlefieldCard(g, victim)
	if c.Counters[game.CounterStun] != 0 {
		t.Fatalf("stun after the lock source left = %d, want 0", c.Counters[game.CounterStun])
	}
}

// A counter-removal cost can't be paid from a permanent whose counters
// are locked: the option is withheld before anything is spent
// (CR 118.3), and lifts when the lock source goes.
func TestFearOfSleepParalysisWithholdsRemovalCostOption(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	fear := fearOnBattlefield(g, opp.ID)
	mine := stunnedTapped(g, me.ID, 1)
	cost := RemoveCountersAmong(game.CounterStun, 1, "a permanent you control")
	options := func() int {
		n := 0
		g.WithWriteLock(func() {
			n = len(g.CounterCostOptionsForEffect(me.ID, mine, cost.RemoveCounters))
		})
		return n
	}
	if got := options(); got != 0 {
		t.Fatalf("options with the lock in place = %d, want 0", got)
	}
	eerieDestroy(t, g, fear)
	if got := options(); got != 1 {
		t.Fatalf("options after the lock source left = %d, want 1", got)
	}
}

// Eerie: an enchantment entering taps up to one target creature and
// stuns it.
func TestFearOfSleepParalysisEerieTapsAndStuns(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		me, opp := g.Seats[0], g.Seats[1]
		fearOnBattlefield(g, me.ID)
		victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Target")
		fire(eerieAnswers{Target: game.TargetRef{Kind: game.TargetCard, ID: victim}})
		c, _ := battlefieldCard(g, victim)
		if !c.Tapped || c.Counters[game.CounterStun] != 1 {
			t.Fatalf("target after eerie: tapped=%v stun=%d, want tapped with 1 stun counter", c.Tapped, c.Counters[game.CounterStun])
		}
	})
}
