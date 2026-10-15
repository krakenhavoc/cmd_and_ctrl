package game

import (
	"testing"

	"github.com/google/uuid"
)

// targets_paid_for_test.go -- #3546, ADR 0186 section 3. A clause over
// creatures with TargetSpec.PaidForSource is "that crewed it" / "that
// saddled it": judged in the one shared walk against the turn's crew
// records of the clause's own source object, so the legal set, the
// announce check (CR 601.2c) and the resolution re-check (CR 608.2b)
// agree.

func paidSpec(kind PaidForSourceKind) *TargetSpec {
	return &TargetSpec{
		Mode: "creature", Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		PaidForSource: kind,
		Min:           1, Max: 1,
	}
}

func objRef(g *Game, id uuid.UUID) ObjectRef {
	var c Card
	g.ReadSnapshot(func() { c = *findBattlefieldCard(g, id) })
	return ObjectRef{ID: id, Epoch: c.ObjectEpoch}
}

func recordPaid(g *Game, saddle bool, permanent, creature uuid.UUID) {
	p, c := objRef(g, permanent), objRef(g, creature)
	g.WithWriteLock(func() { g.recordCrewLocked(saddle, p, c) })
}

// Without the clause, every creature qualifies; with it, only those
// that paid for THIS source, and only the kind of payment it names.
func TestPaidForSourceOffersOnlyThisSourcesPayers(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	mount := pushScopedTestCreature(g, me, 4, 4)
	other := pushScopedTestCreature(g, me, 4, 4)
	saddler := pushScopedTestCreature(g, me, 1, 1)
	elsewhere := pushScopedTestCreature(g, me, 1, 1)
	crewer := pushScopedTestCreature(g, me, 1, 1)
	idle := pushScopedTestCreature(g, me, 1, 1)
	recordPaid(g, true, mount, saddler)
	recordPaid(g, true, other, elsewhere)
	recordPaid(g, false, mount, crewer)

	src := sourceOf(g, me, mount)
	if got := legalIDs(g, src, paidSpec("")); !got[idle] || !got[saddler] || !got[elsewhere] {
		t.Errorf("no clause offered %v, want every creature", got)
	}
	got := legalIDs(g, src, paidSpec(PaidSaddled))
	if len(got) != 1 || !got[saddler] {
		t.Errorf("saddled-this offered %v, want only %s", got, saddler)
	}
	if got := legalIDs(g, src, paidSpec(PaidCrewed)); len(got) != 1 || !got[crewer] {
		t.Errorf("crewed-this offered %v, want only %s", got, crewer)
	}
	// Announce: the same answer, creature by creature.
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want bool
	}{"its saddler": {saddler, true}, "another Mount's saddler": {elsewhere, false}, "its crewer": {crewer, false}, "an idle creature": {idle, false}} {
		if legalIn(g, src, paidSpec(PaidSaddled), tc.id) != tc.want {
			t.Errorf("announce of %s = %v, want %v", name, !tc.want, tc.want)
		}
	}
}

// A walk with no source object admits nothing, never everything: the
// direction a simplification may err (#259).
func TestPaidForSourceNeedsASource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	mount := pushScopedTestCreature(g, me, 4, 4)
	saddler := pushScopedTestCreature(g, me, 1, 1)
	recordPaid(g, true, mount, saddler)
	if got := legalIDs(g, SourceChooser(me), paidSpec(PaidSaddled)); len(got) != 0 {
		t.Errorf("no source admits %v, want nothing", got)
	}
	if legalIn(g, SourceChooser(me), paidSpec(PaidSaddled), saddler) {
		t.Error("announce accepted a payer with no source to compare it to")
	}
}

// A saddler that leaves the battlefield is not offered, and a saddler
// that left and came back is a new object that saddled nothing (CR 400.7).
func TestPaidForSourceDropsASaddlerThatLeftOrReturned(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	mount := pushScopedTestCreature(g, me, 4, 4)
	saddler := pushScopedTestCreature(g, me, 1, 1)
	recordPaid(g, true, mount, saddler)
	src := sourceOf(g, me, mount)
	spec := paidSpec(PaidSaddled)
	if !legalIn(g, src, spec, saddler) {
		t.Fatal("setup: the saddler should be legal")
	}
	// It is a new object (flickered in response): same id, later epoch.
	g.WithWriteLock(func() { findBattlefieldCard(g, saddler).ObjectEpoch++ })
	if legalIn(g, src, spec, saddler) {
		t.Error("a re-entered saddler is still a legal target")
	}
	if got := legalIDs(g, src, spec); len(got) != 0 {
		t.Errorf("a re-entered saddler is offered: %v", got)
	}
}

// A source that has left the battlefield is read from its last-known
// record (CR 608.2h): its saddlers stay its saddlers for an ability
// already on the stack, while a new object under the same instance is not.
func TestPaidForSourceReadsTheDepartedSource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	mount := pushScopedTestCreature(g, me, 4, 4)
	saddler := pushScopedTestCreature(g, me, 1, 1)
	recordPaid(g, true, mount, saddler)

	var rec PermanentInfo
	g.WithWriteLock(func() {
		g.rememberDepartingPermanentLocked(mount)
		c := findBattlefieldCard(g, mount)
		rec, _ = g.lastKnownPermanentLocked(ObjectRef{ID: mount, Epoch: c.ObjectEpoch})
	})
	src := SourceSnapshot(me, &rec.Characteristic)
	src.Departed, src.DepartedID = &rec, mount
	if got := legalIDs(g, src, paidSpec(PaidSaddled)); len(got) != 1 || !got[saddler] {
		t.Errorf("departed source offered %v, want only %s", got, saddler)
	}
	later := src
	rec2 := rec
	rec2.Epoch++
	later.Departed = &rec2
	if got := legalIDs(g, later, paidSpec(PaidSaddled)); len(got) != 0 {
		t.Errorf("a later object of the same card offered %v, want nothing", got)
	}
}
