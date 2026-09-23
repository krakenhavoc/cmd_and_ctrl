package game

import "testing"

// lki_counters_test.go covers #1218: the counters-at-departure half
// of "Post-departure LKI / 'exiled with this' record"
// (docs/engine-seams.md). The Ozolith's "whenever a creature you
// control leaves the battlefield, if it had counters on it, put those
// counters on The Ozolith" needs the departing creature's counters —
// and MoveCard's battlefield-exit cleanup zeroes Card.Counters
// (zone.go) before any watcher, self or bystander, sees the event.
// game/lastKnownCounters and LastKnownCountersForEffect are the fix;
// the effects package's TestOzolith* files are the card-level proof
// that a BYSTANDER trigger (Ozolith is not the departing permanent)
// reads the right answer through the real harvester.

// TestSnapshotLKICapturesCountersBeforeMoveCardClearsThem is the unit
// test at the level of the two calls battlefieldExitLocked sits
// between: the snapshot must see the counters, because the very next
// line (MoveCard) is what zeroes them.
func TestSnapshotLKICapturesCountersBeforeMoveCardClearsThem(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		c.Counters = map[string]int{"+1/+1": 3, "stun": 1}
	})

	g.WithWriteLock(func() {
		g.battlefieldExitLocked(id)
		got := g.LastKnownCountersForEffect(id)
		if got["+1/+1"] != 3 || got["stun"] != 1 {
			t.Errorf("LastKnownCountersForEffect = %v, want +1/+1:3 stun:1", got)
		}
	})
}

// TestSnapshotLKICountersIsADeepCopy proves the snapshot does not
// alias the live card's map — a mutation of one must not be visible
// through the other, which matters because MoveCard nils the live
// map out from under it a line later.
func TestSnapshotLKICountersIsADeepCopy(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	live := map[string]int{"+1/+1": 2}
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		c.Counters = live
	})

	g.WithWriteLock(func() {
		g.battlefieldExitLocked(id)
	})
	live["+1/+1"] = 99
	g.WithWriteLock(func() {
		if got := g.LastKnownCountersForEffect(id); got["+1/+1"] != 2 {
			t.Errorf("snapshot aliased the live map: got %v after mutating the source to 99", got)
		}
	})
}

// TestLastKnownCountersIsClearedAfterTheHarvest is the other half of
// the contract: the entry does not leak past the event it was taken
// for. A real destroy (not a direct battlefieldExitLocked call) runs
// the whole pipeline including harvestLTB's cleanup.
func TestLastKnownCountersIsClearedAfterTheHarvest(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		c.Counters = map[string]int{"+1/+1": 2}
	})

	destroy(t, g, id)

	g.WithWriteLock(func() {
		if got := g.LastKnownCountersForEffect(id); got != nil {
			t.Errorf("lastKnownCounters leaked past the harvest: %v", got)
		}
	})
}

// TestLastKnownCountersIsNilForACounterlessDeparture is the common
// case, and it is what keeps this free: nothing is allocated when
// there is nothing to carry.
func TestLastKnownCountersIsNilForACounterlessDeparture(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := regenBear(g, me.ID)

	destroy(t, g, id)

	g.WithWriteLock(func() {
		if got := g.LastKnownCountersForEffect(id); got != nil {
			t.Errorf("LastKnownCountersForEffect = %v for a counterless departure, want nil", got)
		}
		if g.lastKnownCounters != nil {
			t.Error("lastKnownCounters map was allocated for a counterless departure")
		}
	})
}
