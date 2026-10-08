package game

import (
	"testing"

	"github.com/google/uuid"
)

// #2664: a land put into a graveyard from the battlefield, under the
// controller it had as it left. A bounce or an exile is not it, and
// neither is a nonland permanent or an opponent's land.
func TestTurnTallyCountsLandsPutIntoAGraveyard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	forest, bounced, exiled, rock, theirs := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: bounced, Name: "Island", TypeLine: "Basic Land — Island", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: exiled, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: rock, Name: "Sol Ring", TypeLine: "Artifact", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: theirs, Name: "Plains", TypeLine: "Basic Land — Plains", Owner: opp, Controller: opp})
	})
	g.WithWriteLock(func() {
		g.BounceCardsToHandForEffect([]uuid.UUID{bounced})
		g.ExileCardsForEffect([]uuid.UUID{exiled})
		if err := g.SacrificePermanentForEffect(rock); err != nil {
			t.Fatal(err)
		}
	})
	if g.TurnTallyFor(me).LandsToGraveyard != 0 || g.LandToGraveyardThisTurn(me) {
		t.Fatalf("bounce, exile and an artifact counted: %+v", g.TurnTallyFor(me))
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(theirs); err != nil {
			t.Fatal(err)
		}
	})
	if g.LandToGraveyardThisTurn(me) || !g.LandToGraveyardThisTurn(opp) {
		t.Fatalf("an opponent's land counted for me, or not for them: me %+v opp %+v", g.TurnTallyFor(me), g.TurnTallyFor(opp))
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(forest); err != nil {
			t.Fatal(err)
		}
	})
	if got := g.TurnTallyFor(me).LandsToGraveyard; got != 1 {
		t.Errorf("sacrificed land: %d, want 1", got)
	}
}
