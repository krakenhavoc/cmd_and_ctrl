package game

import (
	"testing"

	"github.com/google/uuid"
)

// attach_new_object_test.go — #3159. A host that leaves the battlefield
// and returns inside one resolution, keeping its instance ID, is a new
// object (CR 400.7), and no state-based check ran in between. The
// attachment pinned to the old object must not follow it: an Aura goes
// to the graveyard (CR 303.4c, 704.5m), an Equipment is unattached and
// stays (CR 301.5c, 704.5n).

// bounceHostAndBack moves the host to its owner's graveyard and back in
// one write lock, as a same-resolution return does, then runs the
// state-based checks.
func bounceHostAndBack(t *testing.T, g *Game, owner *Player, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, owner.Graveyard, host); err != nil {
			t.Fatalf("MoveCard out: %v", err)
		}
		if _, err := MoveCard(owner.Graveyard, g.Battlefield, host); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: host, OldZone: ZoneGraveyard, NewZone: ZoneBattlefield})
		g.runStateChecksLocked()
	})
}

func TestHostReturningAsANewObjectKeepsNoAura(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	aura := pushAttachTestCard(g, me.ID, "Test Aura", "Enchantment — Aura")
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	bounceHostAndBack(t, g, me, bear)

	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Fatal("CR 303.4c: the Aura enchanted the old object and must leave the battlefield")
	}
	if !graveyardHas(me, aura) {
		t.Error("the Aura should be in its owner's graveyard")
	}
	if _, ok := battlefieldCardByID(g, bear); !ok {
		t.Error("the returned creature should be on the battlefield")
	}
}

func TestHostReturningAsANewObjectKeepsNoEquipment(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	bounceHostAndBack(t, g, me, bear)

	got, ok := battlefieldCardByID(g, sword)
	if !ok {
		t.Fatal("CR 301.5c: the Equipment stays on the battlefield")
	}
	if got.IsAttached() {
		t.Errorf("the Equipment is still attached to %+v after its host became a new object", got.AttachedTo)
	}
}

// TestHandSetAttachmentIsNotPinned is the compatibility half: a link
// nobody pinned (a fixture that sets AttachedTo by hand, or a game
// restored from a file written before the pin existed) keeps the old
// behaviour rather than falling off on its first state-based check.
func TestHandSetAttachmentIsNotPinned(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			switch c.InstanceID {
			case bear:
				c.ObjectEpoch = 7
			case sword:
				c.AttachedTo = TargetRef{Kind: TargetCard, ID: bear}
			}
		}
		g.runStateChecksLocked()
	})
	got, ok := battlefieldCardByID(g, sword)
	if !ok || !got.IsAttachedTo(bear) {
		t.Errorf("an unpinned link must stay attached; got %+v", got.AttachedTo)
	}
}

// TestReequipOntoTheSameObjectIsNotAMove pins CR 701.3c with the pin in
// place: attaching to the host the Equipment is already on, while it is
// still the same object, does nothing — no new timestamp.
func TestReequipOntoTheSameObjectIsNotAMove(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	host := TargetRef{Kind: TargetCard, ID: bear}
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, host); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	first, _ := battlefieldCardByID(g, sword)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, host); err != nil {
			t.Fatalf("AttachForEffect again: %v", err)
		}
	})
	second, _ := battlefieldCardByID(g, sword)
	if !second.AttachedPinned {
		t.Error("AttachForEffect should pin the link to the host's object")
	}
	if first.AttachedEpoch != second.AttachedEpoch {
		t.Errorf("re-attach to the same object changed the pin: %d -> %d", first.AttachedEpoch, second.AttachedEpoch)
	}
}
