package game

import (
	"testing"

	"github.com/google/uuid"
)

// reconfigure_test.go — CR 702.151 and CR 301.5c (#2639), against a
// stubbed catalog. The real cards and the activation path are tested in
// cards/effects (reconfigure_cards_test.go).

const (
	reconfigureOracle = "test-reconfigure-equipment"
	animatedOracle    = "test-animated-equipment"
)

// stubReconfigure gives reconfigureOracle its two marked rows. The
// effects are never run here.
func stubReconfigure(t *testing.T) {
	t.Helper()
	noop := func(*Game, *StackItem) error { return nil }
	stubActivatedFor(t, reconfigureOracle,
		ActivatedAbilityShape{Label: "Reconfigure {2} (attach)", Cost: AbilityCost{Mana: "{2}"}, SorcerySpeed: true, Reconfigure: true, Effect: noop},
		ActivatedAbilityShape{Label: "Reconfigure {2} (unattach)", Cost: AbilityCost{Mana: "{2}"}, SorcerySpeed: true, Reconfigure: true, Effect: noop},
	)
}

func reconfigureBoard(t *testing.T) (*Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	stubReconfigure(t)
	host := pushGateCard(g, "Bear", "Creature — Bear", "", me.ID)
	blades := pushGateCard(g, "Lizard Blades", "Artifact Creature — Equipment Lizard", reconfigureOracle, me.ID)
	return g, host, blades
}

func attachForTest(t *testing.T, g *Game, attachment, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(attachment, TargetRef{Kind: TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
		g.runStateChecksLocked()
		g.RecomputeLayersIfStaleLocked()
	})
}

func reconfiguredCardForTest(t *testing.T, g *Game, id uuid.UUID) Card {
	t.Helper()
	var out Card
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c := findBattlefieldCard(g, id)
		if c == nil {
			t.Fatalf("%v is not on the battlefield", id)
		}
		out = *c
	})
	return out
}

// CR 702.151b: attached, it is an "Artifact — Equipment" — no creature
// type and no creature subtype (CR 205.3d). Unattached, it is a creature
// again with its Lizard type back.
func TestAttachedReconfigureEquipmentIsNotACreature(t *testing.T) {
	g, host, blades := reconfigureBoard(t)

	if c := reconfiguredCardForTest(t, g, blades); !c.IsCreature() || !c.HasSubtype("Lizard") {
		t.Fatalf("unattached: creature=%v Lizard=%v, want both", c.IsCreature(), c.HasSubtype("Lizard"))
	}

	attachForTest(t, g, blades, host)
	c := reconfiguredCardForTest(t, g, blades)
	if !c.IsAttachedTo(host) {
		t.Fatalf("the Equipment fell off: %+v (CR 301.5c lets a reconfigure Equipment equip)", c.AttachedTo)
	}
	if c.IsCreature() || c.HasSubtype("Lizard") {
		t.Errorf("attached: creature=%v Lizard=%v, want neither (CR 702.151b, 205.3d)", c.IsCreature(), c.HasSubtype("Lizard"))
	}
	if !c.IsArtifact() || !c.HasSubtype("Equipment") {
		t.Errorf("attached: artifact=%v Equipment=%v, want both kept", c.IsArtifact(), c.HasSubtype("Equipment"))
	}

	g.WithWriteLock(func() {
		if err := g.UnattachForEffect(blades); err != nil {
			t.Fatalf("unattach: %v", err)
		}
	})
	if c := reconfiguredCardForTest(t, g, blades); !c.IsCreature() || !c.HasSubtype("Lizard") {
		t.Errorf("unattached again: creature=%v Lizard=%v, want both back", c.IsCreature(), c.HasSubtype("Lizard"))
	}
}

// CR 701.3d: a host that leaves the battlefield unattaches the
// Equipment, which is a creature again — the state-based action then
// clears the link.
func TestReconfigureEquipmentIsACreatureAgainWhenItsHostLeaves(t *testing.T) {
	g, host, blades := reconfigureBoard(t)
	attachForTest(t, g, blades, host)

	removeFromBattlefieldForTest(t, g, host, false)
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	c := reconfiguredCardForTest(t, g, blades)
	if c.IsAttached() {
		t.Errorf("still attached to a host that left: %+v", c.AttachedTo)
	}
	if !c.IsCreature() {
		t.Error("not a creature after its host left")
	}
}

// CR 301.5c: an Equipment that is also a creature can't equip unless it
// has reconfigure. An animated Equipment without it falls off (CR
// 704.5n); a reconfigure Equipment that a later effect animates stays.
func TestAnEquipmentCreatureEquipsOnlyWithReconfigure(t *testing.T) {
	g, host, blades := reconfigureBoard(t)
	me := g.Seats[0]
	sword := pushGateCard(g, "Animated Sword", "Artifact Creature — Equipment", animatedOracle, me.ID)

	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: host})
		_ = g.AttachForEffect(blades, TargetRef{Kind: TargetCard, ID: host})
		g.runStateChecksLocked()
	})
	if c := reconfiguredCardForTest(t, g, sword); c.IsAttached() {
		t.Errorf("an Equipment creature without reconfigure stayed attached: %+v", c.AttachedTo)
	}
	if c := reconfiguredCardForTest(t, g, blades); !c.IsAttachedTo(host) {
		t.Fatalf("the reconfigure Equipment fell off")
	}

	// A later layer-4 effect makes it a creature again (CR 613.7: newer
	// than the attach). It has reconfigure, so it stays.
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(blades, g.PinnedObjectsLocked(blades),
			[]Mod{AddTypesMod("Creature")}, g.UntilEndOfTurnDuration(), "test — becomes a creature")
		g.RecomputeLayersIfStaleLocked()
		g.runStateChecksLocked()
	})
	c := reconfiguredCardForTest(t, g, blades)
	if !c.IsCreature() {
		t.Fatal("setup: the later effect did not animate it")
	}
	if !c.IsAttachedTo(host) {
		t.Error("an animated reconfigure Equipment fell off; CR 301.5c lets it stay")
	}
}

// The unattach half of CR 702.151a does nothing when its source left and
// came back as a new object (CR 400.7), and unattaches it otherwise.
func TestUnattachSourceForEffect(t *testing.T) {
	g, host, blades := reconfigureBoard(t)
	attachForTest(t, g, blades, host)

	item := &StackItem{Kind: StackItemActivated, SourceCardID: blades, Controller: g.Seats[0].ID}
	g.WithWriteLock(func() {
		item.SourceEpoch = g.cardObjectEpochLocked(blades)
		item.SourceEpoch++ // a different object
		if err := g.UnattachSourceForEffect(item); err != nil {
			t.Fatalf("unattach: %v", err)
		}
	})
	if c := reconfiguredCardForTest(t, g, blades); !c.IsAttachedTo(host) {
		t.Fatal("a new object's ability unattached the Equipment")
	}

	g.WithWriteLock(func() {
		item.SourceEpoch = g.cardObjectEpochLocked(blades)
		if err := g.UnattachSourceForEffect(item); err != nil {
			t.Fatalf("unattach: %v", err)
		}
	})
	if c := reconfiguredCardForTest(t, g, blades); c.IsAttached() || !c.IsCreature() {
		t.Errorf("after unattach: attached=%v creature=%v, want unattached creature", c.IsAttached(), c.IsCreature())
	}
}

func TestHasReconfigureReadsTheRows(t *testing.T) {
	g, _, blades := reconfigureBoard(t)
	c := reconfiguredCardForTest(t, g, blades)
	if !HasReconfigure(&c) || !printsReconfigure(&c) {
		t.Error("a card with reconfigure rows does not have reconfigure")
	}
	if ab := ActivatedAbilitiesForCard(c); len(ab) != 2 || !IsKeywordActivatedAbility(ab[0]) {
		t.Errorf("rows = %+v, want two keyword rows", ab)
	}
}
