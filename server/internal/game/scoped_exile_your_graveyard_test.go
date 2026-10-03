package game

import (
	"testing"

	"github.com/google/uuid"
)

// exileInsteadOfYourGraveyard (ADR 0108 §4, #1823): every card that
// would go to the named player's graveyard this turn, from any zone,
// is exiled instead. A token is not a card, another player's graveyard
// is not "yours", and the record ends with the turn.

func TestExileInsteadOfYourGraveyardCoversEveryRouteIntoItsGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dying := pushScopedTestCreature(g, me.ID, 2, 2)
	theirs := pushScopedTestCreature(g, opp.ID, 2, 2)
	discarded := uuid.New()
	me.Hand.Cards = nil // the random discard can then only pick this card
	me.Hand.PushTop(Card{InstanceID: discarded, Name: "Discard Fodder", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if !g.ExileInsteadOfYourGraveyardThisTurnForEffect(uuid.Nil, me.ID, "Yawgmoth's Will") {
			t.Fatal("setup: registered nothing")
		}
		_ = g.DestroyPermanentForEffect(dying)
		_ = g.DestroyPermanentForEffect(theirs)
		if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	if me.Graveyard.Contains(dying) || !g.Exile.Contains(dying) {
		t.Error("a destroyed permanent of mine was not exiled instead")
	}
	if me.Graveyard.Contains(discarded) || !g.Exile.Contains(discarded) {
		t.Error("a discarded card was not exiled instead (a discard is its own replacement window)")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("a card going to an opponent's graveyard was exiled; the clause says YOUR graveyard")
	}
}

func TestExileInsteadOfYourGraveyardNamesOwnersGraveyardNotControllers(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// I control it, but it is the opponent's card: it dies into THEIR
	// graveyard.
	stolen := uuid.New()
	g.Battlefield.PushTop(Card{InstanceID: stolen, Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		g.ExileInsteadOfYourGraveyardThisTurnForEffect(uuid.Nil, me.ID, "Yawgmoth's Will")
		_ = g.DestroyPermanentForEffect(stolen)
	})
	if !opp.Graveyard.Contains(stolen) {
		t.Error("a card owned by an opponent was exiled when it died under my control")
	}
}

func TestExileInsteadOfYourGraveyardIgnoresTokens(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := pushScopedTestCreature(g, me.ID, 1, 1)
	token := uuid.New()
	g.Battlefield.PushTop(Card{InstanceID: token, Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		g.ExileInsteadOfYourGraveyardThisTurnForEffect(uuid.Nil, me.ID, "Yawgmoth's Will")
		e := g.ScopedEffects[0]
		for id, want := range map[uuid.UUID]bool{card: true, token: false} {
			ev := &ReplacementEvent{Kind: RepEventMove, CardID: id, OldZone: ZoneBattlefield, NewZone: ZoneGraveyard, NewZoneOwner: me.ID}
			if got := scopedReplacementAppliesLocked(g, e, 0, e.Mods[0], ev); got != want {
				t.Errorf("applies to %s = %v, want %v (a token is not a card)", id, got, want)
			}
		}
	})
}

func TestExileInsteadOfYourGraveyardEndsWithTheTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	dying := pushScopedTestCreature(g, me.ID, 2, 2)
	g.WithWriteLock(func() {
		g.ExileInsteadOfYourGraveyardThisTurnForEffect(uuid.Nil, me.ID, "Yawgmoth's Will")
		g.ClearEndOfTurnScopedStaticsLocked()
		_ = g.DestroyPermanentForEffect(dying)
	})
	if !me.Graveyard.Contains(dying) {
		t.Error("the replacement outlived its turn")
	}
}

func TestExileInsteadOfYourGraveyardRefusesAnUnknownPlayer(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if g.ExileInsteadOfYourGraveyardThisTurnForEffect(uuid.Nil, uuid.New(), "Yawgmoth's Will") {
			t.Error("registered for a player who is not at the table")
		}
	})
}
