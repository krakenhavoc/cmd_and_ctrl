package game

import (
	"testing"

	"github.com/google/uuid"
)

// attach_transform_test.go — the engine half of Curse of Leeches //
// Leeching Lurker (#2586). The card half, with the catalogued Aura's
// CR 704.5m exemption while the question is open, is
// cards/effects/curse_of_leeches_test.go.

// CR 704.5p's first sentence: a creature attached to an object or
// player becomes unattached and stays on the battlefield. It is what
// happens to Curse of Leeches when it turns into Leeching Lurker while
// still attached to a player.
func TestSBAUnattachesACreatureFromAPlayer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lurker := pushAttachTestCard(g, me.ID, "Leeching Lurker", "Creature — Leech Horror")
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(lurker, TargetRef{Kind: TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
		g.runStateChecksLocked()
	})
	c, ok := battlefieldCardByID(g, lurker)
	if !ok {
		t.Fatal("the creature left the battlefield; CR 704.5p only unattaches it")
	}
	if c.IsAttached() {
		t.Errorf("the creature is still attached to %+v", c.AttachedTo)
	}
}

// QueueAttachSourceToPlayerForEffect refuses a permanent that is not an
// Aura on the battlefield: there is nothing to attach.
func TestQueueAttachSourceToPlayerNeedsAnAura(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	var queued uuid.UUID
	g.WithWriteLock(func() { queued = g.QueueAttachSourceToPlayerForEffect(bear, "attach it") })
	if queued != uuid.Nil || len(g.PendingChoices) != 0 {
		t.Errorf("queued %v / %d prompts for a non-Aura", queued, len(g.PendingChoices))
	}
}

// For an Aura the prompt is addressed to its controller, offers every
// live seat, and the answer attaches.
func TestQueueAttachSourceToPlayerAttachesOnTheAnswer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, third := g.Seats[0], g.Seats[2]
	aura := pushAttachTestCard(g, me.ID, "Test Curse", "Enchantment — Aura Curse")
	var queued uuid.UUID
	g.WithWriteLock(func() { queued = g.QueueAttachSourceToPlayerForEffect(aura, "attach it") })
	if queued == uuid.Nil {
		t.Fatal("no prompt queued for an Aura on the battlefield")
	}
	var choice *PendingChoice
	for _, c := range g.PendingChoices {
		if c.ID == queued {
			choice = c
		}
	}
	if choice == nil || choice.Chooser != me.ID || len(choice.PickOptions) != len(g.Seats) {
		t.Fatalf("prompt %+v, want addressed to the controller with every seat", choice)
	}
	idx := -1
	for i, o := range choice.PickOptions {
		if o.Player == third.ID {
			idx = i
		}
	}
	if err := g.ResolveOptionPick(queued, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	c, _ := battlefieldCardByID(g, aura)
	if c.AttachedTo.Kind != TargetPlayer || c.AttachedTo.ID != third.ID {
		t.Errorf("attached to %+v, want player %s", c.AttachedTo, third.ID)
	}
}
