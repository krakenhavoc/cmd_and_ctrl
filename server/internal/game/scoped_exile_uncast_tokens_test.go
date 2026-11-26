package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_exile_uncast_tokens_test.go — ModExileUncastCreaturesAndTokens,
// Hallowed Moonlight's "until end of turn, if a creature would enter and it
// wasn't cast, exile it instead". The catalog half is in cards/effects.

func TestExileUncastCreaturesAndTokensExilesCardsAndTokensUntilEndOfTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if !g.ExileUncastCreaturesAndTokensForEffect(uuid.Nil, me.ID, g.UntilEndOfTurnDuration(), "test: exile uncast") {
			t.Fatal("the replacement was not registered")
		}
	})
	id := handCreatureFor(g, opp.ID, "Put Bear")
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: opp.ID}, ZoneRef{Kind: ZoneBattlefield}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if g.Battlefield.Contains(id) || !g.Exile.Contains(id) {
		t.Error("a creature put onto the battlefield was not exiled instead")
	}
	before := len(g.BattlefieldCardsForEffect())
	token := Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID}
	if err := g.CreateTokenForEffect(opp.ID, token, 1); err != nil {
		t.Fatalf("CreateTokenForEffect: %v", err)
	}
	if got := len(g.BattlefieldCardsForEffect()); got != before {
		t.Errorf("a creature token entered the battlefield (%d permanents, want %d)", got, before)
	}
	// A card that is not a creature is untouched.
	land := uuid.New()
	g.Seats[0].Hand.PushTop(Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: me.ID}, ZoneRef{Kind: ZoneBattlefield}, land); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was exiled instead of entering")
	}
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	later := handCreatureFor(g, opp.ID, "Later Bear")
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: opp.ID}, ZoneRef{Kind: ZoneBattlefield}, later); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if !g.Battlefield.Contains(later) {
		t.Error("the replacement outlived the turn")
	}
}
