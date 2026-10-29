package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// manifest_from_hands_test.go — ADR 0189 decision 3 (#3538): the engine
// door under "manifest N cards from your hand".

// Cards from two different hands enter in one batch, each under its
// owner, each face down and known to its owner alone, and `then` is told
// both permanents exactly once.
func TestManifestFromHandsPutsEachCardUnderItsOwner(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := handCardFor(a, "Bear", "Creature — Bear")
	theirs := handCardFor(b, "Ancestral Recall", "Instant")

	resetProbes()
	g.WithWriteLock(func() {
		if err := g.ManifestFromHandsKeyedForEffect([]uuid.UUID{mine, theirs}, probeTail()); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	calls := len(testProbeHeard)
	var got []uuid.UUID
	if calls > 0 {
		got = testProbeHeard[0].IDs
	}
	if calls != 1 || len(got) != 2 {
		t.Fatalf("then ran %d times with %d permanents, want once with 2", calls, len(got))
	}
	for _, tc := range []struct {
		id    uuid.UUID
		owner *Player
	}{{mine, a}, {theirs, b}} {
		card, zone, ok := cardAnywhere(g, tc.id)
		if !ok || zone != ZoneBattlefield {
			t.Fatalf("card is in %q, want the battlefield", zone)
		}
		if !card.FaceDown || card.FaceDownKind != FaceDownManifested {
			t.Errorf("face-down state = (%v, %q), want manifested", card.FaceDown, card.FaceDownKind)
		}
		if card.Controller != tc.owner.ID {
			t.Errorf("controller %s, want the owner %s", card.Controller, tc.owner.ID)
		}
		if !card.IsKnownTo(tc.owner.ID) {
			t.Error("the owner may not look at their own manifest (CR 708.5)")
		}
		for _, other := range []*Player{a, b, c} {
			if other.ID != tc.owner.ID && card.IsKnownTo(other.ID) {
				t.Errorf("seat %s knows another player's manifest", other.ID)
			}
		}
		if !card.IsCreature() {
			t.Error("a manifested card is a 2/2 creature whatever it printed (CR 701.40a)")
		}
	}
	if a.Hand.Contains(mine) {
		t.Error("the card is still in its owner's hand")
	}
}

// One card not in a hand refuses the whole call: nothing moves, and
// `then` still runs exactly once (the batch's own contract).
func TestManifestFromHandsRefusesAStaleIDAndMovesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	inHand := handCardFor(me, "Bear", "Creature — Bear")
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: uuid.New(), Name: "Elsewhere", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	stale := g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID

	resetProbes()
	var err error
	g.WithWriteLock(func() {
		err = g.ManifestFromHandsKeyedForEffect([]uuid.UUID{inHand, stale}, probeTail())
	})
	calls := len(testProbeHeard)
	if calls > 0 && len(testProbeHeard[0].IDs) != 0 {
		t.Errorf("then was told %d permanents after a refusal", len(testProbeHeard[0].IDs))
	}
	if !errors.Is(err, ErrCardNotFound) {
		t.Fatalf("err = %v, want ErrCardNotFound", err)
	}
	if calls != 1 {
		t.Errorf("then ran %d times, want once", calls)
	}
	if !me.Hand.Contains(inHand) {
		t.Error("a refused call still moved the good card")
	}
}
