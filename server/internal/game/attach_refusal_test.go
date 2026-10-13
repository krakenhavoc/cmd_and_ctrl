package game

import (
	"testing"

	"github.com/google/uuid"
)

// attach_refusal_test.go — #3160, CR 701.3b / 303.4j: an effect that
// tries to attach an Aura, Equipment or Fortification to an object or
// player it can't be attached to does nothing. The attachment stays
// where it was and the log says why (EventAttachSkipped).
//
// The enchant-clause refusal for a CATALOGUED Aura lives in
// cards/effects/attachments_refusal_test.go: it needs a real TargetSpec.
// Everything here is the part the engine owns without a catalog.

// attachSkippedFor returns the Label of the latest EventAttachSkipped
// naming `attachment`, or "" when there is none.
func attachSkippedFor(g *Game, attachment uuid.UUID) string {
	var out string
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventAttachSkipped && ev.CardID == attachment {
				out = ev.Label
			}
		}
	})
	return out
}

func attachTo(t *testing.T, g *Game, attachment uuid.UUID, host TargetRef) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(attachment, host); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
}

func cardHost(id uuid.UUID) TargetRef { return TargetRef{Kind: TargetCard, ID: id} }

// A red Aura aimed at a creature with protection from red doesn't move
// (CR 702.16c, 701.3b): it stays on the creature it was already on.
func TestAttachRefusesAHostProtectedFromTheAttachment(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	plain := pushAttachTestCard(g, me.ID, "Plain Bear", "Creature — Bear")
	pro := pushColouredCreature(g, me, "Pro-Red Knight", []string{"W"}, "protection from red")
	g.WithWriteLock(func() { g.recomputeLayersLocked() })
	aura := pushAttachTestCard(g, me.ID, "Red Aura", "Enchantment — Aura")
	g.WithWriteLock(func() {
		g.Battlefield.Cards[findCardOnBattlefield(g, aura)].Colors = []string{"R"}
	})
	attachTo(t, g, aura, cardHost(plain))

	attachTo(t, g, aura, cardHost(pro))

	got, ok := battlefieldCardByID(g, aura)
	if !ok {
		t.Fatal("the Aura left the battlefield: a refused attach must not move or bury it")
	}
	if !got.IsAttachedTo(plain) {
		t.Errorf("AttachedTo = %+v, want it still on the plain bear %s", got.AttachedTo, plain)
	}
	if why := attachSkippedFor(g, aura); why == "" {
		t.Error("no EventAttachSkipped for the refused attach")
	}
}

// The same attach to a creature with no such protection is unaffected,
// and it pins the host's object.
func TestALegalAttachIsUnaffectedAndPinsTheHostObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	first := pushAttachTestCard(g, me.ID, "First Bear", "Creature — Bear")
	second := pushAttachTestCard(g, me.ID, "Second Bear", "Creature — Bear")
	aura := pushAttachTestCard(g, me.ID, "Test Aura", "Enchantment — Aura")
	attachTo(t, g, aura, cardHost(first))
	attachTo(t, g, aura, cardHost(second))

	got, _ := battlefieldCardByID(g, aura)
	if !got.IsAttachedTo(second) {
		t.Fatalf("AttachedTo = %+v, want the second bear", got.AttachedTo)
	}
	host, _ := battlefieldCardByID(g, second)
	if got.AttachedEpoch != host.ObjectEpoch+1 {
		t.Errorf("AttachedEpoch = %d, want the host's ObjectEpoch+1 = %d", got.AttachedEpoch, host.ObjectEpoch+1)
	}
	if why := attachSkippedFor(g, aura); why != "" {
		t.Errorf("a legal attach was skipped: %q", why)
	}
}

// CR 301.5c: an Equipment attaches to a creature. An effect that aims
// it at anything else does nothing, and the Equipment keeps its host.
func TestAttachRefusesEquipmentOntoANoncreature(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	rock := pushAttachTestCard(g, me.ID, "Test Rock", "Artifact")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	attachTo(t, g, sword, cardHost(bear))

	attachTo(t, g, sword, cardHost(rock))

	got, ok := battlefieldCardByID(g, sword)
	if !ok || !got.IsAttachedTo(bear) {
		t.Errorf("the Equipment should stay on the bear (onBattlefield=%v, AttachedTo=%+v)", ok, got.AttachedTo)
	}
	if attachSkippedFor(g, sword) == "" {
		t.Error("no EventAttachSkipped for the refused attach")
	}
}

// Equipment attaches to a permanent, never a player (CR 301.5).
func TestAttachRefusesEquipmentOntoAPlayer(t *testing.T) {
	g := newActiveGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")

	attachTo(t, g, sword, TargetRef{Kind: TargetPlayer, ID: foe.ID})

	got, _ := battlefieldCardByID(g, sword)
	if got.IsAttached() {
		t.Errorf("an Equipment attached to a player: %+v", got.AttachedTo)
	}
	if attachSkippedFor(g, sword) == "" {
		t.Error("no EventAttachSkipped for the refused attach")
	}
}

// An Equipment attaches to a creature it is not protected from; a
// creature that is itself an Equipment may do it only with reconfigure.
func TestAttachRefusesAnAnimatedEquipmentWithoutReconfigure(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	golem := pushAttachTestCard(g, me.ID, "Animated Sword", "Artifact Creature — Equipment Golem")

	attachTo(t, g, golem, cardHost(bear))

	if got, _ := battlefieldCardByID(g, golem); got.IsAttached() {
		t.Errorf("an Equipment creature without reconfigure attached to %+v", got.AttachedTo)
	}
	if attachSkippedFor(g, golem) == "" {
		t.Error("no EventAttachSkipped for the refused attach")
	}
}

// CR 301.6: a Fortification attaches to a land.
func TestAttachFortificationNeedsALand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	forest := pushAttachTestCard(g, me.ID, "Test Forest", "Land — Forest")
	fort := pushAttachTestCard(g, me.ID, "Test Fortification", "Artifact — Fortification")

	attachTo(t, g, fort, cardHost(bear))
	if got, _ := battlefieldCardByID(g, fort); got.IsAttached() {
		t.Fatalf("a Fortification attached to a creature: %+v", got.AttachedTo)
	}
	if attachSkippedFor(g, fort) == "" {
		t.Error("no EventAttachSkipped for the refused attach")
	}

	attachTo(t, g, fort, cardHost(forest))
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	got, ok := battlefieldCardByID(g, fort)
	if !ok || !got.IsAttachedTo(forest) {
		t.Errorf("a Fortification on a land should stay (onBattlefield=%v, AttachedTo=%+v)", ok, got.AttachedTo)
	}
}

// A player with protection from everything can't be enchanted
// (CR 702.16i): a Curse aimed at them stays on its current player.
func TestAttachRefusesAPlayerWithProtection(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, foe, third := g.Seats[0], g.Seats[1], g.Seats[2]
	curse := pushAttachTestCard(g, me.ID, "Test Curse", "Enchantment — Aura Curse")
	attachTo(t, g, curse, TargetRef{Kind: TargetPlayer, ID: foe.ID})
	grantProtectionFromEverything(g, third)

	attachTo(t, g, curse, TargetRef{Kind: TargetPlayer, ID: third.ID})

	got, ok := battlefieldCardByID(g, curse)
	if !ok {
		t.Fatal("the Curse left the battlefield")
	}
	if got.AttachedTo.Kind != TargetPlayer || got.AttachedTo.ID != foe.ID {
		t.Errorf("AttachedTo = %+v, want it to stay on the first player", got.AttachedTo)
	}
	if attachSkippedFor(g, curse) == "" {
		t.Error("no EventAttachSkipped for the refused attach")
	}
}

// An attachment with no pin (a fixture, or a game saved before
// AttachedEpoch existed) keeps its host after the host's epoch moves:
// the pre-#3159 behaviour, and the safe one for old saves.
func TestAnUnpinnedAttachmentKeepsItsHost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		g.Battlefield.Cards[findCardOnBattlefield(g, sword)].AttachedTo = cardHost(bear)
		g.Battlefield.Cards[findCardOnBattlefield(g, bear)].ObjectEpoch += 3
		g.runStateChecksLocked()
	})
	if got, ok := battlefieldCardByID(g, sword); !ok || !got.IsAttachedTo(bear) {
		t.Errorf("an unpinned Equipment lost its host (onBattlefield=%v, AttachedTo=%+v)", ok, got.AttachedTo)
	}
}

// Re-attaching to the same ID as a new object re-pins the epoch, so
// the Equipment is on the host as it is now.
func TestReattachingToTheNewObjectRepinsTheEpoch(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	attachTo(t, g, sword, cardHost(bear))
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, bear); err != nil {
			t.Fatalf("MoveCard out: %v", err)
		}
		if _, err := MoveCard(me.Graveyard, g.Battlefield, bear); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
	})
	attachTo(t, g, sword, cardHost(bear))
	g.WithWriteLock(func() { g.runStateChecksLocked() })

	got, ok := battlefieldCardByID(g, sword)
	if !ok || !got.IsAttachedTo(bear) {
		t.Fatalf("a fresh equip to the new object should hold (onBattlefield=%v, AttachedTo=%+v)", ok, got.AttachedTo)
	}
	host, _ := battlefieldCardByID(g, bear)
	if got.AttachedEpoch != host.ObjectEpoch+1 {
		t.Errorf("AttachedEpoch = %d, want %d", got.AttachedEpoch, host.ObjectEpoch+1)
	}
}

// Clone (undo) carries the pin, and the snapshot mirror does too: a
// restored game must not let a flickered host keep its Aura, nor
// forget a pin it was given.
func TestAttachedEpochSurvivesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	attachTo(t, g, sword, cardHost(bear))
	want, _ := battlefieldCardByID(g, sword)
	if want.AttachedEpoch == 0 {
		t.Fatal("setup: the attach did not pin the host")
	}

	clone := g.Clone()
	var cloned Card
	clone.ReadSnapshot(func() {
		for _, c := range clone.Battlefield.Cards {
			if c.InstanceID == sword {
				cloned = c
			}
		}
	})
	if cloned.AttachedEpoch != want.AttachedEpoch {
		t.Errorf("Clone: AttachedEpoch = %d, want %d", cloned.AttachedEpoch, want.AttachedEpoch)
	}

	mirror := snapshotCard(want, nil)
	if mirror.AttachedEpoch != want.AttachedEpoch {
		t.Errorf("snapshot mirror: AttachedEpoch = %d, want %d", mirror.AttachedEpoch, want.AttachedEpoch)
	}
	if back := restoreCard(&mirror); back.AttachedEpoch != want.AttachedEpoch {
		t.Errorf("snapshot restore: AttachedEpoch = %d, want %d", back.AttachedEpoch, want.AttachedEpoch)
	}
}
