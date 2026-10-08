package game

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// speed_test.go pins ADR 0138 (#2122): start your engines! gives a
// player speed 1 as a state-based action (CR 702.179a), the inherent
// sourceless trigger raises it once on each of their turns when an
// opponent loses life (CR 702.179d), it stops at 4, and the player
// keeps it (nothing lowers a speed, CR 702.179b).

// pushSpeedPermanent puts a creature with start your engines! on the
// battlefield under owner's control.
func pushSpeedPermanent(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	c := NewCard("Speedy", owner.ID)
	c.TypeLine = "Creature — Test"
	c.Power, c.Toughness = 1, 1
	c.Keywords = []string{KeywordStartYourEngines}
	g.Battlefield.PushTop(c)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == c.InstanceID {
			g.Battlefield.Cards[i].effective = printedEffectiveWith(c, KeywordStartYourEngines)
		}
	}
	g.BumpLayerVersionForTest()
	return c.InstanceID
}

// loseLife takes n life from player as an effect would.
func loseLife(t *testing.T, g *Game, player uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, player, -n); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
}

// speedTriggersWaiting counts the speed triggers queued or on the stack.
func speedTriggersWaiting(g *Game) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.Body == speedIncreaseBody.Key() {
			n++
		}
	}
	for _, m := range g.StackMeta {
		if m != nil && m.Body == speedIncreaseBody.Key() {
			n++
		}
	}
	return n
}

func TestStartYourEnginesGivesSpeedOne(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	if me.Speed != 0 || them.Speed != 0 {
		t.Fatalf("a game starts with no speed, got %d and %d", me.Speed, them.Speed)
	}
	pushSpeedPermanent(t, g, me)
	g.RunStateChecksForTest()
	if me.Speed != 1 {
		t.Errorf("speed = %d, want 1 once a start-your-engines permanent is on the battlefield (CR 702.179a)", me.Speed)
	}
	if them.Speed != 0 {
		t.Errorf("the other player got speed %d from a permanent they don't control", them.Speed)
	}
}

func TestSpeedIsKeptAfterThePermanentLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushSpeedPermanent(t, g, me)
	g.RunStateChecksForTest()
	g.SetSpeedForTest(me.ID, 3)
	g.Battlefield.Remove(id)
	g.RunStateChecksForTest()
	if me.Speed != 3 {
		t.Errorf("speed = %d after the permanent left, want 3 kept (CR 702.179b)", me.Speed)
	}
}

func TestSpeedRisesOnceOnYourTurnWhenAnOpponentLosesLife(t *testing.T) {
	g := newActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	pushSpeedPermanent(t, g, me)
	g.RunStateChecksForTest()

	loseLife(t, g, me.ID, 1) // my own loss is never an opponent's
	if got := speedTriggersWaiting(g); got != 0 {
		t.Fatalf("my own life loss on my turn made %d speed triggers, want 0", got)
	}

	loseLife(t, g, them.ID, 1)
	if got := speedTriggersWaiting(g); got != 1 {
		t.Fatalf("an opponent losing life on my turn made %d speed triggers, want 1", got)
	}
	if me.Speed != 1 {
		t.Errorf("speed rose to %d before the trigger resolved; it uses the stack", me.Speed)
	}
	for _, it := range g.PendingTriggers {
		if it != nil && it.Body == speedIncreaseBody.Key() {
			if it.SourceCardID != uuid.Nil || it.Controller != me.ID {
				t.Errorf("speed trigger source=%v controller=%v, want no source and controller %v", it.SourceCardID, it.Controller, me.ID)
			}
			if !strings.Contains(it.Label, "speed") {
				t.Errorf("speed trigger label = %q", it.Label)
			}
		}
	}
	settleMonarchStack(t, g)
	if me.Speed != 2 {
		t.Fatalf("speed = %d after the trigger resolved, want 2", me.Speed)
	}

	// "This ability triggers only once each turn."
	loseLife(t, g, them.ID, 1)
	if got := speedTriggersWaiting(g); got != 0 {
		t.Errorf("a second loss this turn made %d speed triggers, want 0", got)
	}
	settleMonarchStack(t, g)
	if me.Speed != 2 {
		t.Errorf("speed = %d after a second loss this turn, want 2", me.Speed)
	}
}

func TestSpeedRisesFromDamageThatCostsLife(t *testing.T) {
	g := newActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	src := pushSpeedPermanent(t, g, me)
	g.RunStateChecksForTest()
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, them.ID, 2); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	settleMonarchStack(t, g)
	if me.Speed != 2 {
		t.Errorf("speed = %d after damage to an opponent on my turn, want 2", me.Speed)
	}
}

func TestSpeedDoesNotRiseOffYourTurnOrFromYourOwnLoss(t *testing.T) {
	g := newActiveGame(t)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	pushSpeedPermanent(t, g, me)
	g.RunStateChecksForTest()

	loseLife(t, g, them.ID, 1) // their turn: my ability is "during your turn"
	loseLife(t, g, me.ID, 1)   // my own loss is never an opponent's
	if got := speedTriggersWaiting(g); got != 0 {
		t.Errorf("made %d speed triggers off my turn, want 0", got)
	}
	settleMonarchStack(t, g)
	if me.Speed != 1 {
		t.Errorf("speed = %d, want 1", me.Speed)
	}
}

func TestSpeedStopsAtFour(t *testing.T) {
	g := newActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	pushSpeedPermanent(t, g, me)
	g.RunStateChecksForTest()
	g.SetSpeedForTest(me.ID, MaxSpeed)
	if !g.HasMaxSpeed(me.ID) {
		t.Fatal("speed 4 is max speed")
	}
	loseLife(t, g, them.ID, 1)
	if got := speedTriggersWaiting(g); got != 0 {
		t.Errorf("a player at max speed got %d speed triggers, want 0 (the intervening if)", got)
	}
	g.SetSpeedForTest(me.ID, MaxSpeed+1)
	if me.Speed != MaxSpeed {
		t.Errorf("speed = %d, the write must refuse anything above %d", me.Speed, MaxSpeed)
	}
}

func TestSpeedRisesAgainOnYourNextTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	pushSpeedPermanent(t, g, me)
	g.RunStateChecksForTest()
	loseLife(t, g, them.ID, 1)
	settleMonarchStack(t, g)

	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	advanceToStepOfSeat(t, g, 0, StepPrecombatMain)
	loseLife(t, g, them.ID, 1)
	settleMonarchStack(t, g)
	if me.Speed != 3 {
		t.Errorf("speed = %d after a loss on each of two turns, want 3", me.Speed)
	}
}

// CR 702.179c: a player with no speed told to increase it by 1 gets 1.
func TestIncreasingNoSpeedGivesOne(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { g.increaseSpeedLocked(me.ID) })
	if me.Speed != 1 {
		t.Errorf("speed = %d, want 1", me.Speed)
	}
}

func TestSpeedChangeIsAnEventAndSurvivesClone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	before := len(g.Events)
	g.SetSpeedForTest(me.ID, 2)
	found := false
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventSpeedChanged && ev.Actor == me.ID && ev.Amount == 2 {
			found = true
		}
	}
	if !found {
		t.Error("setting a speed emitted no EventSpeedChanged")
	}
	clone := g.Clone()
	if clone.Seats[0].Speed != 2 {
		t.Errorf("clone speed = %d, want 2", clone.Seats[0].Speed)
	}
}
