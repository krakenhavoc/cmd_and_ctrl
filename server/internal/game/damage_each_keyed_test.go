package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// damage_each_keyed_test.go — #4253: a keyed Each walk paused on one leg's
// CR 616 prompt is written with the rest of the walk as data, and the
// keyed creature door tells its tail the permanent and the damage dealt.
// The catalog restores (Ranger's Firebrand, Creeping Bloodsucker, Ballista
// Wielder) are in effects/sweep_x_restore_test.go.

// A walk over two opponents whose first leg pauses: the paused leg's damage
// event is data and its mirror carries the walk (the other opponent still
// to be dealt damage, the walk's tail) through a restore, and the answer
// deals the rest and tells the tail the total. (The replacements here are
// test closures, which keep this table off the disk; the catalog restores
// use registered ones.)
func TestAKeyedDamageWalkIsWrittenWithTheRestOfTheWalk(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	a, b := g.Seats[1].ID, g.Seats[2].ID
	resetProbes()
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(preventOneDamageTo(a))
		g.RegisterReplacementForTest(doubleDamageTo(a))
		if err := g.DealDamageEachKeyedForEffect(uuid.Nil, []uuid.UUID{a, b}, 3, KeyedTail{Ref: testProbeTail, N: 9}); err != nil {
			t.Fatalf("DealDamageEachKeyedForEffect: %v", err)
		}
	})
	var ev *ReplacementEvent
	for _, c := range g.PendingChoices {
		if c != nil && c.replacementResume != nil {
			ev = c.replacementResume.ev
		}
	}
	if ev == nil || ev.Kind != RepEventDamage {
		t.Fatal("the first leg's damage did not pause on its ordering prompt")
	}
	if !g.damageEventIsData(ev) {
		t.Fatal("a keyed walk's paused leg is not data")
	}
	var out replacementEventSnapshot
	snapshotDamageEvent(ev, &out)
	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var back replacementEventSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	restored := &ReplacementEvent{}
	restoreDamageEvent(&back, restored)
	each := restored.damageTail.each
	if each == nil || len(each.Rest) != 1 || each.Rest[0] != b || each.Amount != 3 || each.Instance != ev.DamageInstance {
		t.Fatalf("the restored walk is %+v, want the second opponent still to take 3 in instance %d", each, ev.DamageInstance)
	}
	if each.Then == nil || each.Then.Key != testProbeTail.Key() || each.Then.N != 9 {
		t.Fatalf("the restored walk's tail is %+v, want the probe with N 9", each.Then)
	}
	if len(testProbeHeard) != 0 {
		t.Fatal("the walk's tail ran before its paused leg landed")
	}
	answerOnlyPrompt(t, g)
	// a: 3 → 2 → 4. b: 3. Total dealt = 7.
	if got := heardAmounts(); len(got) != 1 || got[0] != 7 {
		t.Fatalf("the walk's tail heard %v, want [7]", got)
	}
	if len(testProbeHeard[0].IDs) != 0 {
		t.Errorf("the walk's tail was told %v, want no IDs (a total names no one recipient)", testProbeHeard[0].IDs)
	}
}

// The keyed creature door tells its tail the permanent and the damage that
// landed, and zero damage is an outcome too (CR 120.8).
func TestTheKeyedCreatureDoorTellsThePermanentAndTheDamage(t *testing.T) {
	g := newActiveGame(t)
	id := uuid.New()
	resetProbes()
	g.WithWriteLock(func() {
		pushTestCreature(g, id, g.Seats[1], 4, 9)
		g.RegisterReplacementForTest(doubleDamageTo(id))
		if err := g.DealDamageToCreatureKeyedForEffect(uuid.Nil, id, 2, probeTail()); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureKeyedForEffect(uuid.Nil, id, 0, probeTail()); err != nil {
			t.Fatal(err)
		}
	})
	if len(testProbeHeard) != 2 {
		t.Fatalf("the tail ran %d times, want 2", len(testProbeHeard))
	}
	for i, want := range []int{4, 0} {
		h := testProbeHeard[i]
		if h.Amount != want || len(h.IDs) != 1 || h.IDs[0] != id {
			t.Errorf("run %d told %+v, want the creature and %d", i, h, want)
		}
	}
}
