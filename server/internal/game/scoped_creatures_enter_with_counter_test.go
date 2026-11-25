package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// scoped_creatures_enter_with_counter_test.go — #5521:
// ModCreaturesEnterWithPlusOneCounter, "this turn, each creature you
// control enters with an additional +1/+1 counter on it" (Combine
// Guildmage). CR 614.1c; CR 514.2 ends it in cleanup.

func creaturesEnterWithCounter(t *testing.T, g *Game, player uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.CreaturesEnterWithAnAdditionalPlusOneCounterThisTurnForEffect(uuid.Nil, player, "test: counter on entry") {
			t.Fatal("the replacement was not registered")
		}
	})
}

func putOntoBattlefieldCounters(t *testing.T, g *Game, owner, id uuid.UUID) int {
	t.Helper()
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: owner}, ZoneRef{Kind: ZoneBattlefield}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	c, ok := g.battlefieldCardLocked(id)
	if !ok {
		t.Fatal("the permanent is not on the battlefield")
	}
	return c.Counters[CounterPlusOne]
}

func TestCreaturesEnterWithACounterOnlyForTheirControllerThisTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creaturesEnterWithCounter(t, g, me.ID)
	if n := putOntoBattlefieldCounters(t, g, me.ID, handTypedCardFor(g, me.ID, "Bear", "Creature — Bear")); n != 1 {
		t.Errorf("your creature entered with %d counters, want 1", n)
	}
	if n := putOntoBattlefieldCounters(t, g, me.ID, handTypedCardFor(g, me.ID, "Rod", "Artifact")); n != 0 {
		t.Errorf("your artifact entered with %d counters, want 0", n)
	}
	if n := putOntoBattlefieldCounters(t, g, opp.ID, handTypedCardFor(g, opp.ID, "Opp Bear", "Creature — Bear")); n != 0 {
		t.Errorf("an opponent's creature entered with %d counters, want 0", n)
	}
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	if n := putOntoBattlefieldCounters(t, g, me.ID, handTypedCardFor(g, me.ID, "Next Bear", "Creature — Bear")); n != 0 {
		t.Errorf("the replacement outlived its turn: %d counters", n)
	}
}

func TestTwoCreaturesEnterWithACounterRecordsStack(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	creaturesEnterWithCounter(t, g, me.ID)
	creaturesEnterWithCounter(t, g, me.ID)
	if n := putOntoBattlefieldCounters(t, g, me.ID, handTypedCardFor(g, me.ID, "Bear", "Creature — Bear")); n != 2 {
		t.Errorf("entered with %d counters, want 2 (two abilities resolved)", n)
	}
}

func TestCreaturesEnterWithACounterSurvivesASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	creaturesEnterWithCounter(t, g, me.ID)
	_, restored := roundTrip(t, g)
	if n := putOntoBattlefieldCounters(t, restored, me.ID, handTypedCardFor(restored, me.ID, "Bear", "Creature — Bear")); n != 1 {
		t.Errorf("the restored replacement gave %d counters, want 1", n)
	}
	if creaturesEnterWithPlusOneCounterModProblem(Mod{Kind: ModCreaturesEnterWithPlusOneCounter}) == "" {
		t.Error("a record naming no player was accepted")
	}
	snap := g.CaptureSnapshot()
	for i := range snap.ScopedEffects {
		if len(snap.ScopedEffects[i].Mods) == 1 && snap.ScopedEffects[i].Mods[0].Kind == ModCreaturesEnterWithPlusOneCounter {
			snap.ScopedEffects[i].Mods[0].Player = uuid.Nil
		}
	}
	if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("a restore point naming no player = %v, want ErrUnknownEffectKey", err)
	}
}
