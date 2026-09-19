package game

import (
	"testing"

	"github.com/google/uuid"
)

// attach_departed_source_test.go — #812, CR 702.6a / CR 608.2 /
// CR 301.5c / CR 400.7.
//
// An ability resolves whether or not its source is still around. Equip
// is one of the few that cannot DO anything without it: only a
// permanent can be attached (CR 301.5c), so an equip whose Equipment
// left the battlefield in response resolves and does nothing. It is
// not an error, and before this it was logged as one.
//
// The engine half is here; the card half (Loxodon Warhammer sacrificed
// to its own equip) is in
// cards/effects/attachments_departed_source_test.go.

// activatedItemFor builds the stack item an equip activation would
// leave, with the CR 400.7 epoch stamp both announce paths take.
func activatedItemFor(g *Game, controller, source uuid.UUID, targets ...TargetRef) *StackItem {
	item := &StackItem{
		ID:           uuid.New(),
		Kind:         StackItemActivated,
		Controller:   controller,
		Owner:        controller,
		SourceCardID: source,
		Label:        "Equip {2}",
		Targets:      targets,
	}
	g.ReadSnapshot(func() { item.SourceEpoch = g.cardObjectEpochLocked(source) })
	return item
}

// The control: the source is still the permanent it was, so the equip
// attaches exactly as it always has.
func TestAttachSourceForEffectAttachesWhenTheSourceIsStillThere(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	host := TargetRef{Kind: TargetCard, ID: bear}
	item := activatedItemFor(g, me.ID, sword, host)

	g.WithWriteLock(func() {
		if err := g.AttachSourceForEffect(item, host); err != nil {
			t.Fatalf("AttachSourceForEffect: %v", err)
		}
	})
	if got, _ := battlefieldCardByID(g, sword); !got.IsAttachedTo(bear) {
		t.Fatalf("AttachedTo = %+v, want card %s", got.AttachedTo, bear)
	}
	if n := countEventsOfKind(g, EventAttachSkipped); n != 0 {
		t.Errorf("EventAttachSkipped count = %d, want 0", n)
	}
}

// CR 702.6a with the Equipment gone: the ability resolves, nothing is
// attached, the target is untouched, and no error is reported.
func TestAttachSourceForEffectDoesNothingWhenTheSourceLeft(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	host := TargetRef{Kind: TargetCard, ID: bear}
	item := activatedItemFor(g, me.ID, sword, host)

	var err error
	g.WithWriteLock(func() {
		if sacErr := g.SacrificePermanentForEffect(sword); sacErr != nil {
			t.Fatalf("sacrifice: %v", sacErr)
		}
		err = g.AttachSourceForEffect(item, host)
	})
	if err != nil {
		t.Fatalf("AttachSourceForEffect with a departed source: %v, want nil", err)
	}
	if got, ok := battlefieldCardByID(g, sword); ok {
		t.Fatalf("the sacrificed sword is still on the battlefield: %+v", got.AttachedTo)
	}
	var attachments []uuid.UUID
	g.ReadSnapshot(func() { attachments = g.AttachmentsOf(bear) })
	if len(attachments) != 0 {
		t.Errorf("the target picked up %v", attachments)
	}
	if n := countEventsOfKind(g, EventAttachSkipped); n != 1 {
		t.Errorf("EventAttachSkipped count = %d, want 1", n)
	}
	if n := countEventsOfKind(g, EventEffectError); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0 — the ability resolved and did nothing", n)
	}
}

// CR 400.7: an Equipment bounced and replayed while its equip is on
// the stack keeps its instance ID and is a NEW OBJECT. The old
// ability attaches nothing, even though a card with that ID is back on
// the battlefield — which is the case a bare "is it on the
// battlefield" check gets wrong.
func TestAttachSourceForEffectRefusesANewObjectWithTheSameID(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	host := TargetRef{Kind: TargetCard, ID: bear}
	item := activatedItemFor(g, me.ID, sword, host)

	var err error
	g.WithWriteLock(func() {
		if bounceErr := g.BounceToHandForEffect(sword); bounceErr != nil {
			t.Fatalf("bounce: %v", bounceErr)
		}
		if _, playErr := g.PutFromHandOntoBattlefieldForEffect(sword, HandEntryOptions{Controller: me.ID}); playErr != nil {
			t.Fatalf("replay: %v", playErr)
		}
		err = g.AttachSourceForEffect(item, host)
	})
	if err != nil {
		t.Fatalf("AttachSourceForEffect on a new object: %v, want nil", err)
	}
	got, ok := battlefieldCardByID(g, sword)
	if !ok {
		t.Fatal("the replayed sword is not on the battlefield")
	}
	if got.ObjectEpoch == item.SourceEpoch {
		t.Fatalf("setup: the epoch did not move (%d), so this test proves nothing", got.ObjectEpoch)
	}
	if got.IsAttached() {
		t.Errorf("the NEW object was attached to %+v — CR 400.7 says it has no memory of the ability", got.AttachedTo)
	}
	if n := countEventsOfKind(g, EventAttachSkipped); n != 1 {
		t.Errorf("EventAttachSkipped count = %d, want 1", n)
	}
	if n := countEventsOfKind(g, EventEffectError); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0", n)
	}
}

// The epoch half is asked only of an ACTIVATED item, because those are
// the only items whose announce path stamps one. A trigger reaching
// the same primitive gets the presence test and nothing worse.
func TestAttachSourceForEffectAsksTheEpochOnlyOfAnActivatedItem(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	aura := pushAttachTestCard(g, me.ID, "Test Aura", "Enchantment — Aura")
	host := TargetRef{Kind: TargetCard, ID: bear}
	// A harvested trigger: no epoch stamp, and the source has not moved.
	item := &StackItem{
		ID:           uuid.New(),
		Kind:         StackItemTriggered,
		Controller:   me.ID,
		Owner:        me.ID,
		SourceCardID: aura,
		Label:        "attach it to a creature you control",
	}
	g.WithWriteLock(func() {
		if err := g.AttachSourceForEffect(item, host); err != nil {
			t.Fatalf("AttachSourceForEffect from a trigger: %v", err)
		}
	})
	if got, _ := battlefieldCardByID(g, aura); !got.IsAttachedTo(bear) {
		t.Errorf("a trigger with an unstamped epoch was refused: %+v", got.AttachedTo)
	}
}
