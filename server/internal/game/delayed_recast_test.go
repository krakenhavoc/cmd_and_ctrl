package game

import (
	"testing"

	"github.com/google/uuid"
)

// delayed_recast_test.go — #6221, Ertai's Meddling: "Target spell's
// controller exiles it with X delay counters on it. At the beginning of
// each of that player's upkeeps, if that card is exiled, remove a delay
// counter from it. If the card has no delay counters on it, the player
// puts it onto the stack as a copy of the original spell." CR 707.10:
// the copy has the original's targets.

const meddledOracle = "meddled-fixture-oracle"

type meddledRun struct{ target uuid.UUID }

// withMeddledFixture is a "target player loses 1 life" sorcery recording
// each resolution's target.
func withMeddledFixture(t *testing.T, runs *[]meddledRun) {
	t.Helper()
	def := &CardDef{
		Targets: &TargetSpec{Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1},
		Resolve: func(g *Game, item *StackItem) error {
			var r meddledRun
			if len(item.Targets) == 1 {
				r.target = item.Targets[0].ID
			}
			*runs = append(*runs, r)
			return nil
		},
	}
	prev := CatalogLookup
	t.Cleanup(func() { CatalogLookup = prev })
	CatalogLookup = func(k string) *CardDef {
		if k == meddledOracle {
			return def
		}
		if prev != nil {
			return prev(k)
		}
		return nil
	}
}

func meddledTriggers(g *Game) int {
	n := 0
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Body == delayRecastBody {
			n++
		}
	}
	return n
}

// castAndMeddle casts the fixture at `them` and meddles it with x delay
// counters, returning the card.
func castAndMeddle(t *testing.T, g *Game, me, them *Player, x int) uuid.UUID {
	t.Helper()
	c := NewCard("Meddled Bolt", me.ID)
	c.OracleID = meddledOracle
	c.TypeLine = "Sorcery"
	c.Controller = me.ID
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	var ok bool
	var err error
	g.WithWriteLock(func() { ok, err = g.ExileSpellWithDelayCountersForEffect(c.InstanceID, x, uuid.Nil, "") })
	if !ok || err != nil {
		t.Fatalf("meddle: ok=%v err=%v", ok, err)
	}
	return c.InstanceID
}

// The acceptance test: a spell targeting a player, meddled with X=1, comes
// back at its controller's next upkeep targeting the same player (CR
// 707.10), across a restore; the opponent's upkeep does nothing; the card
// ends in the graveyard and the trigger is gone.
func TestMeddledSpellComesBackAtTheNextUpkeepWithItsTargets(t *testing.T) {
	var runs []meddledRun
	withMeddledFixture(t, &runs)
	g := newActiveGame(t)
	toMainPhase(t, g)
	mySeat := g.Turn.ActiveSeat
	me, them := g.Seats[mySeat], g.Seats[1-mySeat]
	id := castAndMeddle(t, g, me, them, 1)

	if g.Stack.Contains(id) || exiledCardByIDLocked(g, id) == nil {
		t.Fatal("the spell is not in exile")
	}
	if n := exiledCardByIDLocked(g, id).Counters[CounterDelay]; n != 1 {
		t.Fatalf("delay counters = %d, want 1", n)
	}
	if meddledTriggers(g) != 1 {
		t.Fatal("no delay trigger queued")
	}

	_, g = roundTrip(t, g)
	me, them = g.Seats[mySeat], g.Seats[1-mySeat]

	advanceToSeatStep(t, g, 1-mySeat, StepUpkeep, true)
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Fatal("the trigger fired in the opponent's upkeep")
	}
	advanceToSeatStep(t, g, mySeat, StepUpkeep, true)
	settleStack(t, g)
	if len(runs) != 1 || runs[0].target != them.ID {
		t.Fatalf("runs = %+v, want the spell once, at its original target", runs)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the card did not end in its owner's graveyard")
	}
	if meddledTriggers(g) != 0 {
		t.Error("the trigger is still queued after the card left exile")
	}
}

// X=2 waits two upkeeps; a card that leaves exile meanwhile is never put
// on the stack, and the trigger stops (CR 603.4, 400.7).
func TestMeddledSpellWaitsAndStopsWhenTheCardLeavesExile(t *testing.T) {
	var runs []meddledRun
	withMeddledFixture(t, &runs)
	g := newActiveGame(t)
	toMainPhase(t, g)
	mySeat := g.Turn.ActiveSeat
	me, them := g.Seats[mySeat], g.Seats[1-mySeat]
	id := castAndMeddle(t, g, me, them, 2)

	advanceToSeatStep(t, g, mySeat, StepUpkeep, true)
	settleStack(t, g)
	if len(runs) != 0 {
		t.Fatalf("the spell came back with a delay counter left: %+v", runs)
	}
	c := exiledCardByIDLocked(g, id)
	if c == nil || c.Counters[CounterDelay] != 1 {
		t.Fatal("one upkeep did not leave the card exiled with one delay counter")
	}
	// The card leaves exile (to its owner's hand) before the next upkeep.
	if _, err := MoveCard(g.Exile, me.Hand, id); err != nil {
		t.Fatal(err)
	}
	advanceToSeatStep(t, g, mySeat, StepUpkeep, true)
	settleStack(t, g)
	if len(runs) != 0 {
		t.Fatalf("a card no longer in exile came back: %+v", runs)
	}
	if meddledTriggers(g) != 0 {
		t.Error("the trigger outlived the card in exile")
	}
}
