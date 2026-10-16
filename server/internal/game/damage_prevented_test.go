package game

import (
	"testing"

	"github.com/google/uuid"
)

// damage_prevented_test.go — CR 615.13 (#2047): EventDamagePrevented is
// emitted once per application of one prevention effect to one damage
// instance, per recipient, with what it prevented, and never for damage
// that can't be prevented.

// preventedEvents is every EventDamagePrevented in the log, oldest first.
func preventedEvents(g *Game) []Event {
	var out []Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventDamagePrevented {
				out = append(out, ev)
			}
		}
	})
	return out
}

// followUpsOnly is the owed CR 615.5 follow-ups, without the CR 615.13
// announcements owed beside them.
func followUpsOnly(in []PreventionFollowUp) []PreventionFollowUp {
	var out []PreventionFollowUp
	for _, f := range in {
		if !f.Announce {
			out = append(out, f)
		}
	}
	return out
}

// Selfless Squire's "prevent all damage that would be dealt to you this
// turn": two sources' damage in one instance is one application, so one
// event with the total; a second instruction is a second application.
func TestPreventionAnnouncesOncePerApplicationPerInstance(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushColouredCreature(g, opp, "A", []string{"R"})
	b := pushColouredCreature(g, opp, "B", []string{"G"})
	sourceShield(t, g, me.ID, uuid.Nil, 0, BodyRef{})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		err := g.DamageInstanceForEffect(func() error {
			if err := g.DealDamageToPlayerForEffect(a, me.ID, 3); err != nil {
				return err
			}
			return g.DealDamageToPlayerForEffect(b, me.ID, 2)
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	evs := preventedEvents(g)
	if len(evs) != 1 || evs[0].Amount != 5 || evs[0].Target != me.ID {
		t.Fatalf("events %+v, want one for 5 to me: one application to one instance (CR 615.13)", evs)
	}
	if evs[0].Source != a || evs[0].Actor != opp.ID {
		t.Errorf("event source %v / actor %v, want the first source and its controller", evs[0].Source, evs[0].Actor)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(a, me.ID, 4); err != nil {
			t.Fatal(err)
		}
	})
	if evs := preventedEvents(g); len(evs) != 2 || evs[1].Amount != 4 {
		t.Fatalf("events %+v, want a second for 4: a second instruction is a second application", evs)
	}
	if got := lifeOf(g, me.ID); got != start {
		t.Errorf("life %d, want %d", got, start)
	}
}

// CR 615.12 / 615.13: a prevention effect applied to damage that can't be
// prevented prevents nothing, so nothing triggers — while its CR 615.5
// follow-up still runs, with zero.
func TestPreventionAnnouncesNothingForUnpreventableDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
	var calls []followUpCall
	sourceShield(t, g, me.ID, uuid.Nil, 0, testFollowUp(&calls))
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() {
		g.DamageCantBePreventedThisTurnForEffect(uuid.Nil, "Skullcrack")
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 3)
	})
	g.RunStateChecksForTest()
	if got := lifeOf(g, me.ID); got != start-3 {
		t.Fatalf("life %d, want %d", got, start-3)
	}
	if evs := preventedEvents(g); len(evs) != 0 {
		t.Fatalf("events %+v, want none: nothing was prevented", evs)
	}
	if len(calls) != 1 || calls[0].amount != 0 {
		t.Errorf("follow-ups %+v, want one with 0 (CR 615.12)", calls)
	}
}

// A charged shield that covers part of the damage announces the part it
// prevented ("that many"), not the damage it met.
func TestPartialPreventionAnnouncesWhatWasPrevented(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Pinger", []string{"R"})
	sourceShield(t, g, me.ID, src, 2, BodyRef{})
	start := lifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 5) })
	if got := lifeOf(g, me.ID); got != start-3 {
		t.Fatalf("life %d, want %d", got, start-3)
	}
	if evs := preventedEvents(g); len(evs) != 1 || evs[0].Amount != 2 {
		t.Fatalf("events %+v, want one for 2", evs)
	}
}

// CR 615.13 counts applications of EFFECTS: two shields each preventing
// one source's damage in one instance are two triggers.
func TestTwoPreventionEffectsAreTwoApplications(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushColouredCreature(g, opp, "A", []string{"R"})
	b := pushColouredCreature(g, opp, "B", []string{"R"})
	sourceShield(t, g, me.ID, a, 0, BodyRef{})
	sourceShield(t, g, me.ID, b, 0, BodyRef{})
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToPlayerForEffect(a, me.ID, 3)
			return g.DealDamageToPlayerForEffect(b, me.ID, 2)
		})
	})
	evs := preventedEvents(g)
	if len(evs) != 2 || evs[0].Amount != 3 || evs[1].Amount != 2 {
		t.Fatalf("events %+v, want 3 then 2: one per prevention effect", evs)
	}
}

// One application that covered you and a creature you control is one
// event for each recipient, each with its own amount.
func TestPreventionAnnouncesPerRecipient(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Src", []string{"R"})
	bear := pushColouredCreature(g, me, "Bear", []string{"G"})
	g.WithWriteLock(func() {
		if !g.PreventDamageFromSourceThisTurnForEffect(DamageShield{
			Controller: me.ID, ProtectPlayer: me.ID, ProtectTypes: []string{"creature"}, Label: "you and your creatures",
		}) {
			t.Fatal("no shield")
		}
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToPlayerForEffect(src, me.ID, 3)
			return g.DealDamageToCreatureForEffect(src, bear, 2)
		})
	})
	evs := preventedEvents(g)
	if len(evs) != 2 {
		t.Fatalf("events %+v, want one for me and one for the bear", evs)
	}
	got := map[uuid.UUID]int{evs[0].Target: evs[0].Amount, evs[1].Target: evs[1].Amount}
	if got[me.ID] != 3 || got[bear] != 2 {
		t.Errorf("amounts %v, want 3 to me and 2 to the bear", got)
	}
}

// Protection is a prevention effect (CR 702.16e): it announces too.
func TestProtectionAnnouncesPrevention(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grantProtectionFromEverything(g, me)
	src := pushColouredCreature(g, opp, "Src", []string{"R"})
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 4) })
	if evs := preventedEvents(g); len(evs) != 1 || evs[0].Amount != 4 || evs[0].Target != me.ID {
		t.Fatalf("events %+v, want one for 4 to me", evs)
	}
}

// A damage instance still open as a restore point is taken keeps the
// announcement it owes: the restored game emits it.
func TestAnnouncementSurvivesARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushColouredCreature(g, me, "Bear", []string{"G"})
	a := pushColouredCreature(g, opp, "A", []string{"R"})
	g.WithWriteLock(func() {
		if !g.PreventDamageFromSourceThisTurnForEffect(DamageShield{
			Controller: me.ID, ProtectPermanent: bear, Label: "the bear",
		}) {
			t.Fatal("no shield")
		}
		g.markCombatDamageOnCardLocked(bear, 2, a, CombatStepRegular)
	})
	snap := g.CaptureSnapshot()
	if snap == nil || len(snap.PreventionFollowUps) != 1 || !snap.PreventionFollowUps[0].Announce {
		t.Fatalf("snapshot owes %+v, want the announcement", snap)
	}
	r, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	r.WithWriteLock(func() { r.runStateChecksLocked() })
	if evs := preventedEvents(r); len(evs) != 1 || evs[0].Amount != 2 || evs[0].Target != bear {
		t.Fatalf("restored game emitted %+v, want one for 2 to the bear", evs)
	}
}
