package aiseat_test

// census_tally_test.go — ADR 0041 P7's soak-side measurement (#1497,
// #1558 item 4): how far behind a restore point a table gets between
// clean captures, measured across every action of every game the
// catalog soak plays.
//
// The shutdown census (ws.LogShutdownCensus) reads one instant per
// deploy; across 19 real deploys every table happened to be idle at
// that instant, so it could not rank anything blocking a restore
// point. ws.Room already tallies what blocks a restore point PER ROOM
// as it plays (writeRestorePointLocked's skipTally, ws/room.go) — this
// is the same measurement, taken from outside the room and summed
// across a whole soak run, with one thing the room-level tally does
// not keep: the longest run of CONSECUTIVE actions any one KIND
// stayed the blocker, not just the room's overall run. It exists only
// here, is never wired into ws.Room or the server, and costs nothing
// on any production path.
//
// It deliberately reuses the engine's own capture and census code —
// game.Game.CaptureSnapshot, game.GameSnapshot.Restorable and
// game.ContinuationCensus.Kinds are exactly what
// ws.Room.writeRestorePointLocked calls — rather than declaring a
// second census of what blocks a restore point.

import "testing"

// censusTally accumulates the measurement across every game in a
// soak run. Zero value is not usable; construct with newCensusTally.
type censusTally struct {
	// Actions is every capture attempt this tally has observed, across
	// every game.
	Actions int `json:"actions"`
	// Captured is how many of those were restorable — a clean
	// capture, i.e. what writeRestorePointLocked would have WRITTEN.
	Captured int `json:"captured"`
	// ByKind is blocked captures per census kind
	// (game.ContinuationCensus.Kinds's keys), summed across every game
	// in the run. A capture blocked by two kinds at once counts once
	// under each, matching ws.Room.skipTally's ByKind.
	ByKind map[string]int `json:"by_kind,omitempty"`
	// LongestRunByKind is, per kind, the longest run of CONSECUTIVE
	// captures that kind blocked, across the whole run — the number
	// the shutdown census cannot give, since it only ever reads one
	// instant.
	LongestRunByKind map[string]int `json:"longest_run_by_kind,omitempty"`

	// runByKind is the current in-progress run per kind. Not
	// marshalled (unexported): it is working state, not a result.
	runByKind map[string]int
}

// newCensusTally returns a tally ready to accumulate. It must be used
// rather than a zero censusTally, so note has maps to write into.
func newCensusTally() *censusTally {
	return &censusTally{
		ByKind:           map[string]int{},
		LongestRunByKind: map[string]int{},
		runByKind:        map[string]int{},
	}
}

// note records one capture attempt's outcome. present is the set of
// kinds blocking it this instant — game.ContinuationCensus.Kinds()'s
// result, empty for a capture that was restorable.
//
// A run is per kind: a kind present in two consecutive calls extends
// its run by one; a kind that stops appearing ends its run (and may
// start a new one later, from 1, without disturbing the longest run
// already recorded).
func (c *censusTally) note(present map[string]int) {
	c.Actions++
	if len(present) == 0 {
		c.Captured++
	}
	for kind := range c.runByKind {
		if _, ok := present[kind]; !ok {
			delete(c.runByKind, kind)
		}
	}
	for kind := range present {
		c.ByKind[kind]++
		c.runByKind[kind]++
		if c.runByKind[kind] > c.LongestRunByKind[kind] {
			c.LongestRunByKind[kind] = c.runByKind[kind]
		}
	}
}

// newGame ends every in-progress run without recording a capture. A
// stale run is a fact about ONE table: the last blocked capture of one
// game and the first of the next are two tables, not one run.
func (c *censusTally) newGame() {
	clear(c.runByKind)
}

// TestCensusTally pins note's bookkeeping against synthetic census
// data — no Scryfall dump, no game, no bot. TestCatalogSoak (gated on
// CMDCTRL_SCRYFALL_DUMP) is the only thing that feeds it real data.
func TestCensusTally(t *testing.T) {
	tally := newCensusTally()

	// Three clean captures in a row: no kind is ever charged, and
	// every action counts as captured.
	tally.note(nil)
	tally.note(map[string]int{})
	tally.note(nil)

	// A five-action run of "choiceResumeFrames" blocking alone.
	for i := 0; i < 5; i++ {
		tally.note(map[string]int{"choiceResumeFrames": 1})
	}

	// A clean capture ends that run.
	tally.note(nil)

	// A shorter, later run of the same kind must not raise the
	// longest run past what the five-action run set.
	tally.note(map[string]int{"choiceResumeFrames": 1})
	tally.note(map[string]int{"choiceResumeFrames": 1})
	tally.note(nil)

	// Two kinds blocking together for two actions, then one drops
	// while the other continues for one more action. Each kind's run
	// tracks independently.
	tally.note(map[string]int{"stackEffects": 2, "stackTargetSpecs": 1})
	tally.note(map[string]int{"stackEffects": 1, "stackTargetSpecs": 3})
	tally.note(map[string]int{"stackEffects": 4})

	wantActions := 3 + 5 + 1 + 3 + 3
	if tally.Actions != wantActions {
		t.Errorf("Actions = %d, want %d", tally.Actions, wantActions)
	}
	wantCaptured := 3 + 1 + 1
	if tally.Captured != wantCaptured {
		t.Errorf("Captured = %d, want %d", tally.Captured, wantCaptured)
	}

	wantByKind := map[string]int{
		"choiceResumeFrames": 5 + 2,
		"stackEffects":       3,
		"stackTargetSpecs":   2,
	}
	for kind, want := range wantByKind {
		if got := tally.ByKind[kind]; got != want {
			t.Errorf("ByKind[%q] = %d, want %d", kind, got, want)
		}
	}
	for kind := range tally.ByKind {
		if _, ok := wantByKind[kind]; !ok {
			t.Errorf("unexpected kind in ByKind: %q = %d", kind, tally.ByKind[kind])
		}
	}

	wantLongest := map[string]int{
		"choiceResumeFrames": 5, // the earlier five-action run, not the later two-action one
		"stackEffects":       3, // present in all three of the last actions
		"stackTargetSpecs":   2, // present in only the first two
	}
	for kind, want := range wantLongest {
		if got := tally.LongestRunByKind[kind]; got != want {
			t.Errorf("LongestRunByKind[%q] = %d, want %d", kind, got, want)
		}
	}

	// A run does not carry across a game boundary: stackEffects was
	// blocking when the last game ended, and blocks the first capture
	// of the next, but that is two tables, not a run of four.
	tally.newGame()
	tally.note(map[string]int{"stackEffects": 1})
	if got, want := tally.LongestRunByKind["stackEffects"], 3; got != want {
		t.Errorf("LongestRunByKind across a game boundary = %d, want %d", got, want)
	}

	// A count in Kinds() above 1 (Doubling Season territory) still
	// counts the kind once per action, exactly as ws.Room.skipTally
	// does: presence blocks, not magnitude.
	solo := newCensusTally()
	solo.note(map[string]int{"turnScopedReplacements": 7})
	solo.note(map[string]int{"turnScopedReplacements": 1})
	if got, want := solo.ByKind["turnScopedReplacements"], 2; got != want {
		t.Errorf("ByKind ignores magnitude: got %d, want %d", got, want)
	}
	if got, want := solo.LongestRunByKind["turnScopedReplacements"], 2; got != want {
		t.Errorf("LongestRunByKind ignores magnitude: got %d, want %d", got, want)
	}
}
