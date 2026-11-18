package game

import (
	"testing"

	"github.com/google/uuid"
)

// exiled_by_ability_test.go — #5876: an exile made by a resolving
// ABILITY writes its controller on the card (Card.ExiledByAbilityOf), for
// Evelyn, the Covetous's "if it was exiled by an ability you controlled".

// An ability I control exiles an opponent's permanent and a card off the
// top of their library: both name me, on both exit paths.
func TestAbilityExileStampsItsController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushBear(g, opp.ID)
	top := uuid.New()
	opp.Library.PushTop(Card{InstanceID: top, Name: "Top Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	g.mu.Lock()
	resolvingAbilityOf(g, me.ID)
	err1 := g.ExileCardForEffect(victim)
	err2 := g.ExileCardForEffect(top)
	g.mu.Unlock()
	if err1 != nil || err2 != nil {
		t.Fatalf("ExileCardForEffect: %v / %v", err1, err2)
	}
	for _, id := range []uuid.UUID{victim, top} {
		if got := exiledCard(t, g, id).ExiledByAbilityOf; got != me.ID {
			t.Errorf("ExiledByAbilityOf = %s, want the ability's controller", got)
		}
	}
}

// A resolving SPELL's exile is not an ability's, and stamps nothing.
func TestSpellExileStampsNoAbilityController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushBear(g, opp.ID)
	g.mu.Lock()
	g.beginResolvingLocked(&StackItem{ID: uuid.New(), Kind: StackItemSpell, Controller: me.ID, Owner: me.ID})
	g.resolutionOpen = true
	err := g.ExileCardForEffect(victim)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	if got := exiledCard(t, g, victim).ExiledByAbilityOf; got != uuid.Nil {
		t.Errorf("a spell's exile stamped %s", got)
	}
}

// Nothing resolving, nothing stamped; and a card that leaves exile and
// comes back by a manual move was exiled by nobody (CR 400.7).
func TestExiledByAbilityClearsOnTheNextMove(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushBear(g, opp.ID)
	g.mu.Lock()
	resolvingAbilityOf(g, me.ID)
	err := g.ExileCardForEffect(victim)
	g.resolving, g.resolutionOpen = nil, false
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneExile}, ZoneRef{Kind: ZoneHand, Owner: opp.ID}, victim); err != nil {
		t.Fatal(err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: opp.ID}, ZoneRef{Kind: ZoneExile}, victim); err != nil {
		t.Fatal(err)
	}
	if got := exiledCard(t, g, victim).ExiledByAbilityOf; got != uuid.Nil {
		t.Errorf("the mark survived a move out of exile: %s", got)
	}
}

// The mark survives a snapshot round trip.
func TestExiledByAbilitySurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushBear(g, opp.ID)
	g.mu.Lock()
	resolvingAbilityOf(g, me.ID)
	err := g.ExileCardForEffect(victim)
	g.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	c := exiledCard(t, g, victim)
	snap := snapshotCard(*c, nil)
	if back := restoreCard(&snap); back.ExiledByAbilityOf != me.ID {
		t.Errorf("restored ExiledByAbilityOf = %s, want %s", back.ExiledByAbilityOf, me.ID)
	}
}
