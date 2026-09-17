package game

import (
	"testing"

	"github.com/google/uuid"
)

// destroyed_this_way_test.go pins #815: "for each creature destroyed
// this way" counts what was DESTROYED, read off the landed outcome.
//
// destroyPermanentsLocked used to count a leg as destroyed whenever
// the call returned no error, which is true of three things that are
// not a destruction: a permanent the CR 614 window saved outright
// (indestructible granted mid-window, "it isn't destroyed instead"), a
// permanent a replacement sent somewhere other than a graveyard, and a
// commander whose exit is merely PAUSED on the CR 903.9 prompt and has
// not happened yet.
//
// The rule, per CR 701.7a ("to destroy a permanent, move it from the
// battlefield to its owner's graveyard"): a graveyard is a
// destruction, the command zone is a destruction (CR 903.9 replaces
// where the commander goes, not whether it was destroyed — the
// engine's declared carry-over, see simultaneous.go), and everything
// else is not. The count is a continuation rather than a return value
// because any leg can pause.

// wipeTarget puts a plain creature on the battlefield under owner.
func wipeTarget(g *Game, owner *Player) uuid.UUID {
	return pushBear(g, owner.ID)
}

// destroyAllThen runs the destroy-all sweep under the write lock and
// returns the landed list the continuation was handed, plus how many
// times the continuation ran (it must run exactly once, and only when
// the whole sweep has settled).
func destroyAllThen(t *testing.T, g *Game, ids []uuid.UUID) (*[]uuid.UUID, *int) {
	t.Helper()
	var got []uuid.UUID
	ran := 0
	g.WithWriteLock(func() {
		err := g.DestroyPermanentsThenForEffect(ids, func(_ *Game, destroyed []uuid.UUID) error {
			ran++
			got = destroyed
			return nil
		})
		if err != nil {
			t.Fatalf("DestroyPermanentsThenForEffect: %v", err)
		}
	})
	return &got, &ran
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

// TestDestroyedThisWayCountsOnlyWhatWasDestroyed is the issue. A wipe
// hits three creatures: one is destroyed, one has its destruction
// cancelled outright, one is exiled instead. Only the first was
// destroyed.
func TestDestroyedThisWayCountsOnlyWhatWasDestroyed(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	destroyed := wipeTarget(g, owner)
	saved := wipeTarget(g, owner)
	exiled := wipeTarget(g, owner)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(saved,
			"it can't be destroyed", func(ev *ReplacementEvent) { ev.Cancel() }))
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(exiled,
			"exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
	})

	got, ran := destroyAllThen(t, g, []uuid.UUID{destroyed, saved, exiled})

	if *ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", *ran)
	}
	if !idsEqual(*got, []uuid.UUID{destroyed}) {
		t.Errorf("destroyed this way = %v, want just the one that reached a graveyard (%v)", *got, destroyed)
	}
	if !owner.Graveyard.Contains(destroyed) {
		t.Error("the plain creature is in its owner's graveyard")
	}
	if findBattlefieldCard(g, saved) == nil {
		t.Error("the cancelled destruction leaves its permanent on the battlefield")
	}
	if !g.Exile.Contains(exiled) {
		t.Error("the redirected destruction put its permanent in exile")
	}
}

// TestAnExiledDestructionIsNotADestruction states the CR 701.7a call
// on its own, because it is the half that is a judgement rather than
// an obvious bug: the permanent really did leave the battlefield, and
// it still pays out nothing. CR 701.7a defines destroying as moving
// the permanent to its owner's GRAVEYARD; a replacement that sends it
// somewhere else means that never happened.
func TestAnExiledDestructionIsNotADestruction(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	exiled := wipeTarget(g, owner)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(exiled,
			"exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
	})

	got, ran := destroyAllThen(t, g, []uuid.UUID{exiled})

	if *ran != 1 || len(*got) != 0 {
		t.Errorf("destroyed this way = %v (continuation ran %d times), want none — "+
			"the permanent left the battlefield but was never put into a graveyard (CR 701.7a)", *got, *ran)
	}
	if !g.Exile.Contains(exiled) {
		t.Error("the permanent is in exile")
	}
}

// TestAPausedCommanderLegIsCountedOnceItLands — the CR 903.9 half.
// The count cannot be taken while the prompt is open, so the whole
// sweep waits for it; when the answer arrives the commander is
// counted, and so is the creature queued behind it. Both answers, and
// the creature ordered AFTER the paused leg so the wait is observable.
func TestAPausedCommanderLegIsCountedOnceItLands(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{{"to the command zone", true}, {"to the graveyard", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			owner := g.Seats[0]
			commander := seatCommander(t, g.Battlefield, owner)
			bear := wipeTarget(g, owner)

			got, ran := destroyAllThen(t, g, []uuid.UUID{commander, bear})

			if *ran != 0 {
				t.Fatalf("the continuation ran %d times with the CR 903.9 prompt still open, want 0", *ran)
			}
			if findBattlefieldCard(g, bear) == nil {
				t.Error("the rest of the sweep waits for the paused leg")
			}

			prompt := expectCommanderPrompt(t, g, owner)
			if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, tc.commandZone); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}

			if *ran != 1 {
				t.Fatalf("the continuation ran %d times after the answer, want 1", *ran)
			}
			if !idsEqual(*got, []uuid.UUID{commander, bear}) {
				t.Errorf("destroyed this way = %v, want [commander bear] = %v %v — CR 903.9 "+
					"replaces the zone change, not the destruction", *got, commander, bear)
			}
			landed := owner.Graveyard
			if tc.commandZone {
				landed = owner.Command
			}
			if !landed.Contains(commander) {
				t.Errorf("the commander is in the %s", landed.Kind)
			}
		})
	}
}

// TestUndoAcrossAPausedDestroyLegReplays is the undo contract for the
// new continuation. Rewinding into the open CR 903.9 prompt and
// answering it again has to destroy the same permanents and report the
// same list: the landed list is carried forward by value, so a
// replayed answer cannot see the first run's entry.
func TestUndoAcrossAPausedDestroyLegReplays(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	commander := seatCommander(t, g.Battlefield, owner)
	bear := wipeTarget(g, owner)

	got, ran := destroyAllThen(t, g, []uuid.UUID{commander, bear})
	promptOpen := g.Clone()

	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	first := append([]uuid.UUID(nil), (*got)...)
	if len(first) != 2 {
		t.Fatalf("first run destroyed %v, want both", first)
	}

	*ran = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if findBattlefieldCard(g, bear) == nil || findBattlefieldCard(g, commander) == nil {
		t.Fatal("the rewind puts both permanents back on the battlefield")
	}

	replay := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(replay.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement (replay): %v", err)
	}
	if *ran != 1 {
		t.Errorf("the continuation ran %d times on the replay, want 1", *ran)
	}
	if !idsEqual(*got, first) {
		t.Errorf("replayed destroyed this way = %v, want the first run's %v", *got, first)
	}
}

// TestFireAndForgetDestroyStillCountsWhatLanded — the synchronous
// entry point keeps its int and its meaning is now the same rule. A
// cancelled destruction is not in it. A leg that paused cannot be,
// which is exactly why a card that reads the number uses the Then
// form; that is asserted by the callers above.
func TestFireAndForgetDestroyStillCountsWhatLanded(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	destroyed := wipeTarget(g, owner)
	saved := wipeTarget(g, owner)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(saved,
			"it can't be destroyed", func(ev *ReplacementEvent) { ev.Cancel() }))
	})

	var n int
	g.WithWriteLock(func() {
		n = g.DestroyPermanentsForEffect([]uuid.UUID{destroyed, saved})
	})
	if n != 1 {
		t.Errorf("DestroyPermanentsForEffect = %d, want 1 — a destruction the window cancelled destroyed nothing", n)
	}
}
