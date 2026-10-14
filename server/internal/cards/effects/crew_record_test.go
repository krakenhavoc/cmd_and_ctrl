package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// crew_record_test.go — ADR 0186 slice 1 (#3543): who crewed or saddled
// what this turn. The record is written when the creatures are TAPPED
// (CR 702.122b-c, CR 702.171c), keyed by object, and read back by object.

// crewBy activates the Vehicle's crew ability (ability 0) with the named
// creatures and does NOT let it resolve.
func crewBy(t *testing.T, g *game.Game, owner, vehicle uuid.UUID, crewers ...uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(owner, vehicle, 0, game.ActivateAbilityParams{CrewIDs: crewers}); err != nil {
		t.Fatalf("crew: %v", err)
	}
}

// objectOf is the ObjectRef of a card on the battlefield now.
func objectOf(t *testing.T, g *game.Game, id uuid.UUID) game.ObjectRef {
	t.Helper()
	c, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}
}

func crewRecords(g *game.Game, permanent game.ObjectRef, saddle bool) (out []game.CrewRecord) {
	g.WithWriteLock(func() { out = g.CrewRecordsForEffect(permanent, saddle) })
	return out
}

func crewersOf(g *game.Game, permanent game.ObjectRef) (out []uuid.UUID) {
	g.WithWriteLock(func() { out = g.CreaturesThatCrewedForEffect(permanent) })
	return out
}

func saddlersOfObject(g *game.Game, permanent game.ObjectRef) (out []uuid.UUID) {
	g.WithWriteLock(func() { out = g.CreaturesThatSaddledForEffect(permanent) })
	return out
}

func flickerForTest(t *testing.T, g *game.Game, controller, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (Flicker{Target: id}).Apply(ctxFor(g, &game.StackItem{Controller: controller})); err != nil {
			t.Fatalf("flicker: %v", err)
		}
	})
}

func idsEqual(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Crewing with two creatures writes two records, in payment order, naming
// the Vehicle's object and each creature's object, and no saddle record.
func TestCrewingWritesOneRecordPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	a := pushCrewerForTest(g, owner.ID, "A", 1)
	b := pushCrewerForTest(g, owner.ID, "B", 1)

	crewBy(t, g, owner.ID, copter, a, b)

	vehicle := objectOf(t, g, copter)
	want := []game.CrewRecord{
		{Permanent: vehicle, Creature: objectOf(t, g, a)},
		{Permanent: vehicle, Creature: objectOf(t, g, b)},
	}
	got := crewRecords(g, vehicle, false)
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("crew records = %+v, want %+v", got, want)
	}
	if n := len(crewRecords(g, vehicle, true)); n != 0 {
		t.Errorf("a crew payment wrote %d saddle records", n)
	}
	if got := crewersOf(g, vehicle); !idsEqual(got, []uuid.UUID{a, b}) {
		t.Errorf("CreaturesThatCrewedForEffect = %v, want [a b] in payment order", got)
	}
	// The record is written at payment: the ability has not resolved.
	if v, _ := battlefieldCardByID(g, copter); v.IsCreature() {
		t.Error("the crew ability resolved before the check")
	}
}

// A saddle writes saddle records, and no crew records.
func TestSaddlingWritesSaddleRecords(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := pushMountForTest(g, me.ID, "Gilded Ghoda", gildedGhodaOracle, 2)
	a := pushCrewerForTest(g, me.ID, "A", 1)
	b := pushCrewerForTest(g, me.ID, "B", 1)

	if err := saddleAbility(t, g, me.ID, mount, a, b); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	ref := objectOf(t, g, mount)
	got := crewRecords(g, ref, true)
	if len(got) != 2 || !got[0].Saddle || !got[1].Saddle ||
		got[0].Creature.ID != a || got[1].Creature.ID != b || got[0].Permanent != ref {
		t.Fatalf("saddle records = %+v", got)
	}
	if n := len(crewRecords(g, ref, false)); n != 0 {
		t.Errorf("a saddle payment wrote %d crew records", n)
	}
	if got := saddlersOfObject(g, ref); !idsEqual(got, []uuid.UUID{a, b}) {
		t.Errorf("CreaturesThatSaddledForEffect = %v", got)
	}
}

// CR 702.122c: a crew ability countered after payment still leaves its
// records, because the creatures were tapped to pay.
func TestCounteredCrewAbilityStillLeavesItsRecords(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	crewer := pushCrewerForTest(g, owner.ID, "Crewer", 1)
	crewBy(t, g, owner.ID, copter, crewer)

	var abilityID uuid.UUID
	for id, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemActivated && it.SourceCardID == copter {
			abilityID = id
		}
	}
	if abilityID == uuid.Nil {
		t.Fatal("the crew ability is not on the stack")
	}
	if err := g.CounterAbility(abilityID); err != nil {
		t.Fatalf("CounterAbility: %v", err)
	}
	if v, _ := battlefieldCardByID(g, copter); v.IsCreature() {
		t.Fatal("a countered crew ability crewed the Vehicle")
	}
	if got := crewersOf(g, objectOf(t, g, copter)); !idsEqual(got, []uuid.UUID{crewer}) {
		t.Errorf("crewers after a counter = %v, want the creature that paid", got)
	}
}

// CR 400.7: a crewer flickered after paying is a new object that crewed
// nothing, so the creature reader drops it; the record stays.
func TestFlickeredCrewerIsDroppedButItsRecordStays(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	stays := pushCrewerForTest(g, owner.ID, "Stays", 1)
	blinked := pushCrewerForTest(g, owner.ID, "Blinked", 1)
	crewBy(t, g, owner.ID, copter, stays, blinked)
	vehicle := objectOf(t, g, copter)
	blinkedThen := objectOf(t, g, blinked)

	flickerForTest(t, g, owner.ID, blinked)

	if got := crewersOf(g, vehicle); !idsEqual(got, []uuid.UUID{stays}) {
		t.Errorf("crewers = %v, want only the creature that did not leave", got)
	}
	var recorded bool
	for _, r := range crewRecords(g, vehicle, false) {
		if r.Creature == blinkedThen {
			recorded = true
		}
	}
	if !recorded {
		t.Error("the flickered crewer's record is gone; CrewRecordsForEffect keeps departed creatures")
	}
}

// crewTestIDByName finds the (single) battlefield card with this name.
func crewTestIDByName(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	t.Fatalf("no %q on the battlefield", name)
	return uuid.Nil
}

// saddledMountForTest saddles a Gilded Ghoda with one creature and lets
// the ability resolve.
func saddledMountForTest(t *testing.T, g *game.Game) (me *game.Player, mount, saddler uuid.UUID) {
	t.Helper()
	toMain(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	mount = pushMountForTest(g, me.ID, "Gilded Ghoda", gildedGhodaOracle, 2)
	saddler = pushCrewerForTest(g, me.ID, "Saddler", 1)
	if err := saddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)
	return me, mount, saddler
}

// A Mount that left (flickered here) still answers by its old
// ObjectRef (Fortune, Loyal Steed's ruling), and the object it became
// answers nothing (CR 400.7).
func TestFlickeredMountStillAnswersByItsOldObject(t *testing.T) {
	g := newCatalogGame(t)
	me, mount, saddler := saddledMountForTest(t, g)
	old := objectOf(t, g, mount)

	flickerForTest(t, g, me.ID, mount)

	now := crewTestIDByName(t, g, "Gilded Ghoda")
	nowRef := objectOf(t, g, now)
	if nowRef == old {
		t.Fatal("the flickered Mount is the same object")
	}
	if got := saddlersOfObject(g, old); !idsEqual(got, []uuid.UUID{saddler}) {
		t.Errorf("the old object's saddlers = %v, want [saddler]", got)
	}
	if got := saddlersOfObject(g, nowRef); len(got) != 0 {
		t.Errorf("the new object answered %v; it was never saddled", got)
	}
	var viaID []uuid.UUID
	g.WithWriteLock(func() { viaID = g.SaddlersOf(now) })
	if len(viaID) != 0 {
		t.Errorf("SaddlersOf(new id) = %v", viaID)
	}
}

// A Mount bounced to hand is on no battlefield and still answers by the
// object it was.
func TestBouncedMountStillAnswersByItsOldObject(t *testing.T) {
	g := newCatalogGame(t)
	_, mount, saddler := saddledMountForTest(t, g)
	old := objectOf(t, g, mount)

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(mount); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})

	if got := saddlersOfObject(g, old); !idsEqual(got, []uuid.UUID{saddler}) {
		t.Errorf("saddlers after the Mount left = %v, want [saddler]", got)
	}
	var viaID []uuid.UUID
	g.WithWriteLock(func() { viaID = g.SaddlersOf(mount) })
	if len(viaID) != 0 {
		t.Errorf("SaddlersOf(id) = %v for a Mount that is not on the battlefield", viaID)
	}
}

// Records are gone once the next turn begins.
func TestCrewRecordsAreGoneAfterTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	crewer := pushCrewerForTest(g, owner.ID, "Crewer", 1)
	crewBy(t, g, owner.ID, copter, crewer)
	vehicle := objectOf(t, g, copter)
	if len(crewRecords(g, vehicle, false)) != 1 {
		t.Fatal("no record to begin with")
	}

	advancePastCleanupOf(t, g, seat)

	if n := len(g.TurnTally.Crews); n != 0 {
		t.Errorf("TurnTally.Crews has %d records in the next turn", n)
	}
	if got := crewersOf(g, objectOf(t, g, copter)); len(got) != 0 {
		t.Errorf("crewers next turn = %v", got)
	}
}

// Undo restores the clone taken before the activation, which has no
// records; and the clone does not share the live slice, so a later crew
// does not reach the restore point's view of the turn.
func TestCrewRecordsCloneAndUndo(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	a := pushCrewerForTest(g, owner.ID, "A", 1)
	b := pushCrewerForTest(g, owner.ID, "B", 1)

	before := g.Clone()
	crewBy(t, g, owner.ID, copter, a)
	if len(before.TurnTally.Crews) != 0 {
		t.Fatal("the pre-activation clone carries the record: undo would keep it")
	}
	mid := g.Clone()
	if len(mid.TurnTally.Crews) != 1 {
		t.Fatalf("mid clone records = %d, want 1", len(mid.TurnTally.Crews))
	}
	crewBy(t, g, owner.ID, copter, b)
	if len(g.TurnTally.Crews) != 2 {
		t.Fatalf("live records = %d, want 2", len(g.TurnTally.Crews))
	}
	if len(mid.TurnTally.Crews) != 1 {
		t.Errorf("the clone's slice grew to %d with the live game's", len(mid.TurnTally.Crews))
	}
}

// A game restored from a file with saddledBy and no crews still gives
// Rambling Possum its saddlers (the v8 fallback), and the fallback
// reads only the Mount object that was saddled.
func TestRestoredGameWithoutCrewsStillGivesPossumItsSaddlers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Rambling Possum", ramblingPossumOracle, 2)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)
	toMain(t, g)
	if err := saddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)

	snap := g.CaptureSnapshot()
	if len(snap.TurnTally.Crews) != 1 {
		t.Fatalf("captured crews = %+v", snap.TurnTally.Crews)
	}
	// What a file written before this change looks like.
	snap.TurnTally.Crews = nil
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	var got []uuid.UUID
	restored.WithWriteLock(func() { got = restored.CreaturesThatSaddledForEffect(objectOf(t, restored, mount)) })
	if !idsEqual(got, []uuid.UUID{saddler}) {
		t.Fatalf("fallback saddlers = %v, want [saddler]", got)
	}

	declareAttack(t, restored, opp.ID, mount)
	passPriorityAroundTable(t, restored)
	choice := latestChooseCards(restored, me.ID)
	if choice == nil || !idsEqual(choice.ChooseCards, []uuid.UUID{saddler}) {
		t.Fatalf("Rambling Possum's prompt = %+v, want the saddler", choice)
	}
}

// A record that does survive a snapshot round trip is read back.
func TestCrewRecordsSurviveASnapshotRoundTrip(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	crewer := pushCrewerForTest(g, owner.ID, "Crewer", 1)
	crewBy(t, g, owner.ID, copter, crewer)
	vehicle := objectOf(t, g, copter)

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got := crewRecords(restored, vehicle, false); len(got) != 1 || got[0].Creature.ID != crewer {
		t.Errorf("restored records = %+v", got)
	}
}

// Rambling Possum bounced in response to its own attack trigger still
// offers the creatures that saddled it: the trigger reads its saddlers by
// the Possum's object, not by a battlefield lookup.
func TestRamblingPossumBouncedInResponseStillOffersItsSaddlers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Rambling Possum", ramblingPossumOracle, 2)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)
	toMain(t, g)
	if err := saddleAbility(t, g, me.ID, mount, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, mount)
	if triggerOnStack(g, mount) == nil {
		t.Fatal("the attack trigger is not on the stack")
	}

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(mount); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	choice := latestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("no prompt for the saddlers after the Possum was bounced")
	}
	if !idsEqual(choice.ChooseCards, []uuid.UUID{saddler}) {
		t.Errorf("offered %v, want [saddler]", choice.ChooseCards)
	}
}

func battlefieldViewCard(t *testing.T, g *game.Game, viewer string, id uuid.UUID) protocol.CardView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer)
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("%s is not on %q's battlefield view", id, viewer)
	return protocol.CardView{}
}

// The Vehicle's view carries crewed_by with both crewers for every seat
// and a spectator; it drops a crewer that left, and it is absent next turn.
func TestCrewedByOnTheWire(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	a := pushCrewerForTest(g, owner.ID, "A", 1)
	b := pushCrewerForTest(g, owner.ID, "B", 1)
	for _, id := range []uuid.UUID{copter, a, b} {
		knownToTable(g, id)
	}
	crewBy(t, g, owner.ID, copter, a, b)
	passPriorityAroundTable(t, g)

	viewers := []string{protocol.SpectatorViewerID, ""}
	for _, s := range g.Seats {
		viewers = append(viewers, s.ID.String())
	}
	want := []string{a.String(), b.String()}
	for _, viewer := range viewers {
		got := battlefieldViewCard(t, g, viewer, copter).CrewedBy
		if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Errorf("viewer %q: crewed_by = %v, want %v", viewer, got, want)
		}
		if sb := battlefieldViewCard(t, g, viewer, copter).SaddledBy; len(sb) != 0 {
			t.Errorf("viewer %q: a Vehicle carries saddled_by %v", viewer, sb)
		}
	}
	if cb := battlefieldViewCard(t, g, owner.ID.String(), a).CrewedBy; len(cb) != 0 {
		t.Errorf("a crewer carries crewed_by %v", cb)
	}

	flickerForTest(t, g, owner.ID, b)
	got := battlefieldViewCard(t, g, g.Seats[1].ID.String(), copter).CrewedBy
	if len(got) != 1 || got[0] != a.String() {
		t.Errorf("after one crewer left: crewed_by = %v, want [a]", got)
	}

	advancePastCleanupOf(t, g, seat)
	if got := battlefieldViewCard(t, g, g.Seats[1].ID.String(), copter).CrewedBy; len(got) != 0 {
		t.Errorf("crewed_by next turn = %v, want none", got)
	}
}

// A saddled Mount's view carries saddled_by.
func TestSaddledByOnTheWire(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mount := pushMountForTest(g, me.ID, "Gilded Ghoda", gildedGhodaOracle, 2)
	a := pushCrewerForTest(g, me.ID, "A", 1)
	knownToTable(g, mount)
	if err := saddleAbility(t, g, me.ID, mount, a); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, viewer := range []string{me.ID.String(), g.Seats[1].ID.String(), protocol.SpectatorViewerID} {
		c := battlefieldViewCard(t, g, viewer, mount)
		if len(c.SaddledBy) != 1 || c.SaddledBy[0] != a.String() {
			t.Errorf("viewer %q: saddled_by = %v, want [a]", viewer, c.SaddledBy)
		}
		if len(c.CrewedBy) != 0 {
			t.Errorf("viewer %q: a Mount carries crewed_by %v", viewer, c.CrewedBy)
		}
	}
}
