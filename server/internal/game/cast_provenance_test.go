package game

import (
	"testing"

	"github.com/google/uuid"
)

// cast_provenance_test.go — CR 400.7d at the engine level (#653),
// against fixtures with no catalog entry. The card-level proof (a real
// Phlage, sacrificing itself unless it escaped) is in
// internal/cards/effects/phlage_test.go.
//
// Four facts, and the last two are the ones a per-entry field gets
// wrong if nobody pins them:
//
//	the record is written for a spell cast for an alternative cost
//	it is EMPTY for one cast for its mana cost, and for a permanent
//	  that was never a spell at all
//	it is written BEFORE EventETB, or "sacrifice it unless it
//	  escaped" has nothing to read
//	it ENDS with the permanent (CR 400.7) — a flicker is a new
//	  object that did not escape

// escapedCreature seeds a creature in the active seat's graveyard
// with an escape cost, and returns it plus the fodder that pays.
func escapedCreature(t *testing.T, g *Game, me *Player, oracle string) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	return seedEscapeCreature(t, g, me, oracle, 0)
}

// TestAnEscapedPermanentRemembersIt is the whole issue in one pass.
func TestAnEscapedPermanentRemembersIt(t *testing.T) {
	const oracle = "test-provenance-escape"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := escapedCreature(t, g, me, oracle)

	if err := castEscape(g, me, id, yard); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	resolveTop(t, g)

	prov := g.CastProvenanceForEffect(id)
	if !prov.Escaped() {
		t.Fatalf("provenance = %+v, want an escaped permanent", prov)
	}
	if prov.AltCost != AltCostKeyEscape {
		t.Errorf("AltCost = %q, want %q", prov.AltCost, AltCostKeyEscape)
	}
	if prov.FromZone != ZoneGraveyard {
		t.Errorf("FromZone = %q, want the graveyard it was cast from", prov.FromZone)
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || !c.Escaped() {
		t.Error("Card.Escaped() disagrees with the record")
	}
}

// The same creature cast for its printed cost out of hand remembers
// nothing, and that is the answer "sacrifice it unless it escaped"
// needs: a hard-cast Phlage is sacrificed.
func TestAHardCastPermanentRemembersNothing(t *testing.T) {
	const oracle = "test-provenance-hand"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, _ := escapedCreature(t, g, me, oracle)
	if _, err := MoveCard(me.Graveyard, me.Hand, id); err != nil {
		t.Fatalf("seed to hand: %v", err)
	}

	if err := g.CastSpell(me.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("hand cast: %v", err)
	}
	resolveTop(t, g)

	prov := g.CastProvenanceForEffect(id)
	if prov.Escaped() {
		t.Errorf("a hard-cast permanent reports escaped: %+v", prov)
	}
	// FromZone is still recorded — "cast from your hand" is a fact
	// about the spell too — but the cost half is empty, which is what
	// every reader of this record branches on.
	if prov.AltCost != "" {
		t.Errorf("AltCost = %q, want empty", prov.AltCost)
	}
}

// A permanent that was never a spell — put onto the battlefield,
// reanimated, a token — has the zero record. CR 400.7d is written
// about "the spell that became that permanent", and there wasn't one.
func TestAPermanentThatWasNeverASpellRemembersNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := NewCard("Reanimated Titan", me.ID)
	c.TypeLine = "Creature — Elder Giant"
	c.Power, c.Toughness = 6, 6
	id := c.InstanceID
	me.Graveyard.PushTop(c)

	g.WithWriteLock(func() {
		if _, err := MoveCard(me.Graveyard, g.Battlefield, id); err != nil {
			t.Fatalf("reanimate: %v", err)
		}
	})
	if prov := g.CastProvenanceForEffect(id); prov.Any() {
		t.Errorf("a reanimated permanent carries %+v", prov)
	}
}

// CR 400.7: the record ends with the permanent. A creature that
// escaped, was exiled and came back is a NEW object that did not
// escape — which is the difference between Phlage as a recursion
// engine and Phlage as an infinite one.
func TestProvenanceIsGoneAfterAFlicker(t *testing.T) {
	const oracle = "test-provenance-flicker"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := escapedCreature(t, g, me, oracle)

	if err := castEscape(g, me, id, yard); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	resolveTop(t, g)
	if !g.CastProvenanceForEffect(id).Escaped() {
		t.Fatal("pre-flicker: the record never landed")
	}

	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, g.Exile, id); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	// Cleared on the way OUT, so the card sitting in exile is already
	// telling the truth about itself.
	exiled := findCardForTest(g.Exile, id)
	if exiled == nil || exiled.Provenance.Any() {
		t.Fatalf("the exiled card carries %+v", exiled)
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Exile, g.Battlefield, id); err != nil {
			t.Fatalf("return: %v", err)
		}
	})
	if prov := g.CastProvenanceForEffect(id); prov.Escaped() {
		t.Errorf("a flickered permanent still reports escaped: %+v", prov)
	}
}

// The ordering requirement, and the only one that is not obvious from
// the field's existence: the record has to be written before EventETB,
// because "when this enters, sacrifice it unless it escaped" is a
// trigger harvested off that event. Asserted through a listener rather
// than through a catalog card so it pins the ENGINE's ordering.
func TestProvenanceIsWrittenBeforeTheETBEvent(t *testing.T) {
	const oracle = "test-provenance-etb"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := escapedCreature(t, g, me, oracle)

	probe := &etbProvenanceProbe{card: id}
	g.RegisterListener(probe)

	if err := castEscape(g, me, id, yard); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	resolveTop(t, g)

	if !probe.sawETB {
		t.Fatal("no ETB event for the escaped creature")
	}
	if !probe.escapedAtETB {
		t.Error("the ETB event fired before the permanent knew it had escaped")
	}
}

// etbProvenanceProbe reads the permanent's record at the instant its
// ETB event fires — the moment a triggered ability is harvested.
// OnEvent runs with g.mu already held, so the *ForEffect read is the
// right one.
type etbProvenanceProbe struct {
	card         uuid.UUID
	sawETB       bool
	escapedAtETB bool
}

func (p *etbProvenanceProbe) OnEvent(g *Game, ev Event) {
	if ev.Kind != EventETB || ev.CardID != p.card {
		return
	}
	p.sawETB = true
	p.escapedAtETB = g.CastProvenanceForEffect(p.card).Escaped()
}

// Snapshot and undo. The record was copied off a StackItem that no
// longer exists, so neither a restore nor a rewind can rebuild it
// from anything else.
func TestProvenanceSurvivesASnapshotAndUndoTakesItBack(t *testing.T) {
	const oracle = "test-provenance-snapshot"
	g := newActiveGame(t)
	me := g.Seats[0]
	id, yard := escapedCreature(t, g, me, oracle)

	if err := castEscape(g, me, id, yard); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	snap := g.Clone()
	resolveTop(t, g)
	if !g.CastProvenanceForEffect(id).Escaped() {
		t.Fatal("pre-snapshot: the record never landed")
	}

	_, restored := roundTrip(t, g)
	if !restored.CastProvenanceForEffect(id).Escaped() {
		t.Error("a snapshot round trip lost the record")
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if g.CastProvenanceForEffect(id).Escaped() {
		t.Error("undo rewound past the resolution and left the record behind")
	}
}
