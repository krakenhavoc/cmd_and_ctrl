package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// damage_keyed_test.go — #4253: a damage event's frame is data when its
// tail is, and a keyed damage continuation parked behind a prevention
// follow-up's prompt is written to a restore point. The catalog restore
// (Screaming Nemesis paused on a CR 616 ordering prompt) is in
// effects/sweep_k3_restore_test.go.

// damageFrameEvent is a non-combat damage event to a player with a tail
// and no continuation.
func damageFrameEvent() *ReplacementEvent {
	return &ReplacementEvent{
		Kind: RepEventDamage, DamageSource: uuid.New(), DamageTarget: uuid.New(), DamageAmount: 2, DamageInstance: 3,
		damageTail: &damageTail{kind: damageTailPlayer},
	}
}

// A shield division or a redirection order held for the event's instance,
// and the event's release from a division, are written beside it (#4361
// S16), so none of them keeps the frame off the disk.
func TestADamageFrameWithAHeldDivisionOrRedirectOrderIsKeyed(t *testing.T) {
	g := NewGame()
	ev := damageFrameEvent()
	c := &PendingChoice{Kind: PendingChoiceReplacementOrder, replacementResume: &replacementResumeFrame{
		ev: ev, applicable: []activeReplacement{{id: builtinReplacementIDBase}},
	}}
	if key := g.replacementFrameKeyable(c); key != ReplacementResumeDamage {
		t.Fatalf("a plain damage event's frame is keyed %q, want %q", key, ReplacementResumeDamage)
	}
	g.shieldDivisions = []ShieldDivision{{Instance: 3}}
	if key := g.replacementFrameKeyable(c); key != ReplacementResumeDamage {
		t.Errorf("a frame with a division held for its instance is keyed %q", key)
	}
	g.redirectOrders = []redirectOrder{{instance: 3}}
	if key := g.replacementFrameKeyable(c); key != ReplacementResumeDamage {
		t.Errorf("a frame with a redirection order held for its instance is keyed %q", key)
	}
	ev.damageTail.released = true
	if key := g.replacementFrameKeyable(c); key != ReplacementResumeDamage {
		t.Errorf("a frame released from a division is keyed %q", key)
	}
	// #4404: redirected parts are written too, under a key of their own.
	ev.damageTail.redirected = []redirectedDamage{{}}
	if key := g.replacementFrameKeyable(c); key != ReplacementResumeDamageRedirected {
		t.Errorf("a frame with redirected parts is keyed %q, want %q", key, ReplacementResumeDamageRedirected)
	}
}

// parkProbe sets the told player's life to 100 plus the damage dealt plus
// the tail's N, so a test can read back that it ran and what it was told.
var parkProbe = RegisterTail(testEffectKeyPrefix+"damage-park-probe", func(g *Game, o TailOutcome) error {
	if len(o.IDs) == 1 {
		if p := g.playerByIDLocked(o.IDs[0]); p != nil {
			p.Life = 100 + o.Amount + o.Tail.N
		}
	}
	return nil
})

// A keyed park is written, restored with its instance counted, and runs
// its tail told the target and the damage dealt. (The closure park is gone,
// #4361 S8.)
func TestAKeyedFollowUpParkIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() {
		g.parkFollowUpKeyedLocked(7, KeyedTail{Ref: parkProbe, N: 1}, opp.ID, 3)
	})
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a keyed park is not a restore point: %+v", snap.Continuations)
	}
	if len(snap.FollowUpParks) != 1 {
		t.Fatalf("%d parks written, want 1", len(snap.FollowUpParks))
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var back GameSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	r, err := back.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if len(r.followUpParks) != 1 || r.followUpParks[0].inst != 7 {
		t.Fatalf("restored parks %+v, want the one of instance 7", r.followUpParks)
	}
	if r.damageInstanceSeq < 7 {
		t.Errorf("the instance counter resumed at %d, below the park's 7", r.damageInstanceSeq)
	}
	r.WithWriteLock(func() {
		if r.resumeFollowUpParksLocked() {
			t.Error("the park paused again")
		}
	})
	if got := r.Seats[1].Life; got != 104 {
		t.Errorf("the park's tail left the player at %d, want 104 (told 3 dealt, N 1)", got)
	}
}
