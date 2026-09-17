package game

import (
	"testing"

	"github.com/google/uuid"
)

// exiled_this_way_test.go pins #866: an exile or bounce batch counts
// what LANDED, the way the destroy sweep has since #815.
//
// The rule for these two is CR 400.7 — a card that changes zones
// becomes a new object in the zone it arrives in — so "for each card
// exiled this way" means the cards that reached EXILE. A leg the
// CR 614 window cancelled never moved; a leg a replacement sent
// somewhere else moved, but not this way; a leg that has merely PAUSED
// on the CR 903.9 prompt has not happened yet. All three used to be
// counted, which is why Settle the Wreckage handed its victim a basic
// land for every creature whose owner had been ASKED about the command
// zone.
//
// This is the one place the exile/bounce rule differs from destroy's:
// destroyedThisWayLocked counts the command zone as a destruction by a
// declared carry-over (ADR 0013 §5i), because CR 903.9 replaces where
// the commander goes rather than whether it was destroyed. Nothing
// replaces the fact that an exile put the card in exile — it did not —
// so an exile batch does not count it.

// exileAllThen runs the exile sweep under the write lock and returns
// the landed list the continuation was handed, plus how many times the
// continuation ran.
func exileAllThen(t *testing.T, g *Game, ids []uuid.UUID) (*[]uuid.UUID, *int) {
	t.Helper()
	var got []uuid.UUID
	ran := 0
	g.WithWriteLock(func() {
		err := g.ExileCardsThenForEffect(ids, func(_ *Game, exiled []uuid.UUID) error {
			ran++
			got = exiled
			return nil
		})
		if err != nil {
			t.Fatalf("ExileCardsThenForEffect: %v", err)
		}
	})
	return &got, &ran
}

// TestExiledThisWayCountsOnlyWhatLanded is the issue. A sweep hits
// three creatures: one is exiled, one has its exile cancelled
// outright, one is sent to a graveyard instead. Only the first was
// exiled this way.
func TestExiledThisWayCountsOnlyWhatLanded(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	exiled := wipeTarget(g, owner)
	saved := wipeTarget(g, owner)
	binned := wipeTarget(g, owner)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(saved,
			"it can't be exiled", func(ev *ReplacementEvent) { ev.Cancel() }))
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(binned,
			"put it into its owner's graveyard instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneGraveyard
				ev.NewZoneOwner = owner.ID
			}))
	})

	got, ran := exileAllThen(t, g, []uuid.UUID{exiled, saved, binned})

	if *ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", *ran)
	}
	if !idsEqual(*got, []uuid.UUID{exiled}) {
		t.Errorf("exiled this way = %v, want just the one that reached exile (%v) — "+
			"CR 400.7, the object that arrived", *got, exiled)
	}
	if !g.Exile.Contains(exiled) {
		t.Error("the plain creature is in exile")
	}
	if findBattlefieldCard(g, saved) == nil {
		t.Error("the cancelled exile leaves its permanent on the battlefield")
	}
	if !owner.Graveyard.Contains(binned) {
		t.Error("the redirected exile put its permanent in a graveyard")
	}
}

// TestAPausedCommanderExileLegIsCountedOnlyWhenItLandsInExile — the
// CR 903.9 half, and the row that separates this rule from destroy's.
// The count cannot be taken while the prompt is open, so the whole
// sweep waits for it; when the answer arrives the commander counts only
// if it actually went to exile.
func TestAPausedCommanderExileLegIsCountedOnlyWhenItLandsInExile(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{{"to the command zone", true}, {"to exile", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			owner := g.Seats[0]
			commander := seatCommander(t, g.Battlefield, owner)
			bear := wipeTarget(g, owner)

			got, ran := exileAllThen(t, g, []uuid.UUID{commander, bear})

			if *ran != 0 {
				t.Fatalf("the continuation ran %d times with the CR 903.9 prompt open, want 0", *ran)
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
			want := []uuid.UUID{commander, bear}
			landed := g.Exile
			if tc.commandZone {
				// It left the battlefield, and it was not exiled: the
				// command zone is not exile, and nothing replaced that.
				want = []uuid.UUID{bear}
				landed = owner.Command
			}
			if !idsEqual(*got, want) {
				t.Errorf("exiled this way = %v, want %v", *got, want)
			}
			if !landed.Contains(commander) {
				t.Errorf("the commander is in the %s", landed.Kind)
			}
		})
	}
}

// TestReturnedThisWayCountsOnlyWhatLanded — the bounce half, on the
// same body. A commander that takes CR 903.9's offer was not returned
// to its owner's hand.
func TestReturnedThisWayCountsOnlyWhatLanded(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	commander := seatCommander(t, g.Battlefield, owner)
	bear := wipeTarget(g, owner)

	var got []uuid.UUID
	ran := 0
	g.WithWriteLock(func() {
		err := g.BounceCardsToHandThenForEffect([]uuid.UUID{commander, bear},
			func(_ *Game, bounced []uuid.UUID) error {
				ran++
				got = bounced
				return nil
			})
		if err != nil {
			t.Fatalf("BounceCardsToHandThenForEffect: %v", err)
		}
	})
	if ran != 0 {
		t.Fatalf("the continuation ran %d times with the CR 903.9 prompt open, want 0", ran)
	}

	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}

	if ran != 1 {
		t.Fatalf("the continuation ran %d times after the answer, want 1", ran)
	}
	if !idsEqual(got, []uuid.UUID{bear}) {
		t.Errorf("returned this way = %v, want just %v — the commander went to the command zone, "+
			"not to a hand", got, bear)
	}
	if !owner.Command.Contains(commander) || !owner.Hand.Contains(bear) {
		t.Error("the commander is in the command zone and the bear is in its owner's hand")
	}
}

// TestUndoAcrossAPausedExileLegReplays is the undo contract for the new
// continuation, the same one the destroy sweep signs: rewinding into
// the open CR 903.9 prompt and answering it again has to exile the same
// cards and report the same list. The landed list is carried forward by
// value, so a replayed answer cannot see the first run's entry.
func TestUndoAcrossAPausedExileLegReplays(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	commander := seatCommander(t, g.Battlefield, owner)
	bear := wipeTarget(g, owner)

	got, ran := exileAllThen(t, g, []uuid.UUID{commander, bear})
	promptOpen := g.Clone()

	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	first := append([]uuid.UUID(nil), (*got)...)
	if len(first) != 2 {
		t.Fatalf("first run exiled %v, want both", first)
	}

	*ran = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if findBattlefieldCard(g, bear) == nil || findBattlefieldCard(g, commander) == nil {
		t.Fatal("the rewind puts both permanents back on the battlefield")
	}

	replay := expectCommanderPrompt(t, g, g.Seats[0])
	if err := g.ResolveOptionalReplacement(replay.ID, g.Seats[0].ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement (replay): %v", err)
	}
	if *ran != 1 {
		t.Errorf("the continuation ran %d times on the replay, want 1", *ran)
	}
	if !idsEqual(*got, first) {
		t.Errorf("replayed exiled this way = %v, want the first run's %v", *got, first)
	}
}

// TestFireAndForgetExileAndBounceCountWhatLanded — the synchronous
// entry points keep their int and it now means the same rule. A leg the
// window cancelled is not in it. A leg that paused cannot be, which is
// why a card that reads the number uses the Then form.
func TestFireAndForgetExileAndBounceCountWhatLanded(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	exiled := wipeTarget(g, owner)
	saved := wipeTarget(g, owner)
	bounced := wipeTarget(g, owner)
	stuck := wipeTarget(g, owner)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(saved,
			"it can't be exiled", func(ev *ReplacementEvent) { ev.Cancel() }))
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(stuck,
			"it can't leave the battlefield", func(ev *ReplacementEvent) { ev.Cancel() }))
	})

	var exiledN, bouncedN int
	g.WithWriteLock(func() {
		exiledN = g.ExileCardsForEffect([]uuid.UUID{exiled, saved})
		bouncedN = g.BounceCardsToHandForEffect([]uuid.UUID{bounced, stuck})
	})
	if exiledN != 1 {
		t.Errorf("ExileCardsForEffect = %d, want 1 — an exile the window cancelled exiled nothing", exiledN)
	}
	if bouncedN != 1 {
		t.Errorf("BounceCardsToHandForEffect = %d, want 1 — a bounce the window cancelled returned nothing", bouncedN)
	}
}
