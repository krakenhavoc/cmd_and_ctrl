package game

import (
	"testing"

	"github.com/google/uuid"
)

// attach_legality_test.go pins #3160, CR 701.3b: an effect that moves an
// Aura, Equipment or Fortification that is already attached to an object
// it can't legally be attached to does nothing, and the attachment stays
// where it is. It does NOT cover an Aura entering the battlefield
// attached (CR 303.4f) or an Aura spell resolving (CR 608.3c).

func attachLegalityCatalog(t *testing.T) {
	t.Helper()
	withCatalogTargetSpec(t, func(string) *TargetSpec { return enchantCreatureSpec() })
}

func attachNow(t *testing.T, g *Game, attachment, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.recomputeLayersLocked()
		if err := g.AttachForEffect(attachment, TargetRef{Kind: TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
		g.runStateChecksLocked()
	})
}

// An Aura attached to a creature, moved to a permanent its enchant clause
// does not admit, stays on the creature instead of being buried.
func TestMovingAnAuraOntoAnIllegalHostLeavesItWhereItIs(t *testing.T) {
	attachLegalityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	rock := pushAttachTestCard(g, me.ID, "Mana Rock", "Artifact")
	aura := pushCataloguedAura(g, me.ID, "Test Pacifism")
	attachNow(t, g, aura, bear)
	before := countEventsOfKind(g, EventAttachSkipped)
	attachNow(t, g, aura, rock)

	got, ok := battlefieldCardByID(g, aura)
	if !ok {
		t.Fatal("CR 701.3b: the Aura was put into the graveyard instead of staying put")
	}
	if !got.IsAttachedTo(bear) {
		t.Errorf("the Aura moved to the illegal host: %+v", got.AttachedTo)
	}
	if countEventsOfKind(g, EventAttachSkipped) != before+1 {
		t.Error("the refused move left no EventAttachSkipped")
	}
}

// CR 702.16c: a creature with protection from the Aura's colour can't be
// enchanted by it, so a move onto it does not happen either.
func TestMovingAnAuraOntoAProtectedCreatureLeavesItWhereItIs(t *testing.T) {
	attachLegalityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	pro := pushColouredCreature(g, me, "Pro-Red Bear", []string{"W"}, "protection from red")
	aura := pushCataloguedAura(g, me.ID, "Test Red Aura")
	g.WithWriteLock(func() { findBattlefieldCard(g, aura).Colors = []string{"R"} })
	attachNow(t, g, aura, bear)
	attachNow(t, g, aura, pro)

	got, ok := battlefieldCardByID(g, aura)
	if !ok || !got.IsAttachedTo(bear) {
		t.Fatalf("a red Aura moved onto a pro-red creature: on=%v %+v", ok, got.AttachedTo)
	}
}

// The check must not stop a legal move.
func TestMovingAnAuraOntoALegalHostMoves(t *testing.T) {
	attachLegalityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	ox := pushAttachTestCard(g, me.ID, "Ox", "Creature — Ox")
	aura := pushCataloguedAura(g, me.ID, "Test Pacifism")
	attachNow(t, g, aura, bear)
	attachNow(t, g, aura, ox)
	got, ok := battlefieldCardByID(g, aura)
	if !ok || !got.IsAttachedTo(ox) {
		t.Fatalf("a legal move did not happen: on=%v %+v", ok, got.AttachedTo)
	}
}

// CR 303.4f / 608.3c: an Aura that is not attached yet (it is entering,
// or just resolved) is not "moved". The attach happens and the
// state-based action decides, as before.
func TestAnAuraThatIsNotAttachedYetIsNotRefused(t *testing.T) {
	attachLegalityCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	rock := pushAttachTestCard(g, me.ID, "Mana Rock", "Artifact")
	aura := pushCataloguedAura(g, me.ID, "Test Pacifism")
	before := countEventsOfKind(g, EventAttachSkipped)
	g.WithWriteLock(func() {
		g.recomputeLayersLocked()
		_ = g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: rock})
	})
	if countEventsOfKind(g, EventAttachSkipped) != before {
		t.Error("an entering Aura was refused by the 701.3b move check")
	}
	if c, _ := battlefieldCardByID(g, aura); !c.IsAttachedTo(rock) {
		t.Errorf("the entering Aura should attach (and be undone by the SBA): %+v", c.AttachedTo)
	}
}

// Equipment attached to a creature stays there when an effect tries to
// attach it to a non-creature; an unattached one stays unattached.
func TestMovingEquipmentOntoANonCreatureLeavesItWhereItIs(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	rock := pushAttachTestCard(g, me.ID, "Mana Rock", "Artifact")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	attachNow(t, g, sword, bear)
	attachNow(t, g, sword, rock)
	if got, _ := battlefieldCardByID(g, sword); !got.IsAttachedTo(bear) {
		t.Errorf("the Equipment let go of its creature: %+v", got.AttachedTo)
	}

	loose := pushAttachTestCard(g, me.ID, "Loose Sword", "Artifact — Equipment")
	attachNow(t, g, loose, rock)
	if got, _ := battlefieldCardByID(g, loose); got.IsAttached() {
		t.Errorf("an Equipment attached to a non-creature: %+v", got.AttachedTo)
	}
}

// Moving Equipment between two creatures is untouched by the check.
func TestMovingEquipmentBetweenCreaturesStillMoves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	ox := pushAttachTestCard(g, me.ID, "Ox", "Creature — Ox")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	attachNow(t, g, sword, bear)
	attachNow(t, g, sword, ox)
	if got, _ := battlefieldCardByID(g, sword); !got.IsAttachedTo(ox) {
		t.Errorf("the Equipment did not move: %+v", got.AttachedTo)
	}
}
