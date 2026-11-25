package game

import (
	"testing"

	"github.com/google/uuid"
)

// phasing_extras_test.go — #6201: "can't phase out" (Spatial Binding)
// against CR 702.26a's untap-step phasing and an effect's phase-out.

// A permanent with phasing that can't phase out stays phased in during
// its controller's untap step (CR 702.26a); once the record is gone it
// phases as usual.
func TestCantPhaseOutKeepsAPhasingPermanentIn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: "Tolarian Serpent", TypeLine: "Creature — Serpent",
		Owner: me, Controller: me, Keywords: []string{KeywordPhasing},
	})
	aura := pushPhasingTestCard(g, me, "Test Aura", "Enchantment — Aura")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == aura {
				g.Battlefield.Cards[i].AttachedTo = TargetRef{Kind: TargetCard, ID: id}
			}
		}
		if !g.CantPhaseOutUntilYourNextUpkeepForEffect(uuid.Nil, me, id, "test — can't phase out") {
			t.Fatal("no record was written")
		}
	})
	g.WithWriteLock(func() { g.performPhasingLocked(me) })
	if _, ok := battlefieldCardByID(g, id); !ok {
		t.Fatal("a permanent that can't phase out phased out in its controller's untap step")
	}
	if _, ok := battlefieldCardByID(g, aura); !ok {
		t.Fatal("its Aura phased out without it (CR 702.26g)")
	}
	// An effect's phase-out is refused too.
	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, id) })
	if _, ok := battlefieldCardByID(g, id); !ok {
		t.Fatal("an effect phased out a permanent that can't phase out")
	}

	g.WithWriteLock(func() {
		g.dropScopedEffectsLocked(func(_ int, e ScopedEffect) bool { return scopedEffectHasMod(&e, ModCantPhaseOut) })
	})
	g.WithWriteLock(func() { g.performPhasingLocked(me) })
	if _, ok := phasedOutCardByID(g, id); !ok {
		t.Fatal("with the record gone, it phases out as usual")
	}
}

// "Until your next upkeep": the record lasts through the untap step of
// that turn and ends as the upkeep begins (the duration's own tests pin
// the clock; this pins that the gate reads a live record only).
func TestCantPhaseOutReadsOnlyALiveRecord(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	id := pushPhasingTestCard(g, me, "Grizzly Bears", "Creature — Bear")
	if g.CantPhaseOutForEffect(id) {
		t.Fatal("a permanent with no record can't phase out")
	}
	g.WithWriteLock(func() { g.CantPhaseOutUntilYourNextUpkeepForEffect(uuid.Nil, me, id, "test") })
	if !g.CantPhaseOutForEffect(id) {
		t.Fatal("the record does not cover its permanent")
	}
}
