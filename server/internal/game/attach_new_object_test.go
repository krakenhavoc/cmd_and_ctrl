package game

import "testing"

// attach_new_object_test.go — #3159. A host that leaves the battlefield
// and returns inside one resolution is a new object (CR 400.7) that
// kept its instance ID, and no state-based check ran in between. The
// attachment pointing at the old object must not follow it:
// the Aura goes to the graveyard (CR 303.4c, 704.5m), the Equipment is
// unattached and stays (CR 301.5c, 704.5n).

func TestFlickeredHostKeepsNoAura(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	aura := pushAttachTestCard(g, me.ID, "Test Aura", "Enchantment — Aura")

	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
		if _, err := MoveCard(g.Battlefield, me.Graveyard, bear); err != nil {
			t.Fatalf("MoveCard out: %v", err)
		}
		if _, err := MoveCard(me.Graveyard, g.Battlefield, bear); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: bear, OldZone: ZoneGraveyard, NewZone: ZoneBattlefield})
		g.runStateChecksLocked()
	})

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

func TestFlickeredHostKeepsNoEquipment(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")

	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
		if _, err := MoveCard(g.Battlefield, me.Graveyard, bear); err != nil {
			t.Fatalf("MoveCard out: %v", err)
		}
		if _, err := MoveCard(me.Graveyard, g.Battlefield, bear); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: bear, OldZone: ZoneGraveyard, NewZone: ZoneBattlefield})
		g.runStateChecksLocked()
	})

	got, ok := battlefieldCardByID(g, sword)
	if !ok {
		t.Fatal("CR 301.5c: the Equipment stays on the battlefield")
	}
	if got.IsAttached() {
		t.Errorf("the Equipment is still attached to %+v after its host became a new object", got.AttachedTo)
	}
}
