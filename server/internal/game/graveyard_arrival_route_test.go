package game

import (
	"testing"

	"github.com/google/uuid"
)

// graveyard_arrival_route_test.go pins #931: EVERY way a card reaches
// a graveyard opens the CR 614 window.
//
// Two did not. A library search with a graveyard destination (Entomb,
// Buried Alive, Unmarked Grave) moved the card with a raw MoveCard,
// and so did surveil's graveyard leg — so "if a card would be put into
// a graveyard from anywhere, exile it instead" (Rest in Peace, Leyline
// of the Void) could not see either of them, and neither could
// CR 903.9: a tutored or surveilled commander was never offered the
// command zone. Mill has gone through the window since #529, the
// battlefield exit since S17, a discard since #650.
//
// Both now take the shared batch body (routeAllThenLocked) — the
// search with its own template (searchRoute), surveil with the mill's,
// which is the same library → graveyard move under a different
// keyword. So both can PAUSE, and everything after them is a
// continuation: the search's shuffle and Then, surveil's EventSurveil
// and the rest of the effect.

// searchToGraveyard runs a one-card search for `name` into the
// searcher's graveyard and returns the found list the continuation was
// handed, plus how many times it ran.
func searchToGraveyard(t *testing.T, g *Game, p *Player, name string, dest ZoneKind) (*[]uuid.UUID, *int) {
	t.Helper()
	var found []uuid.UUID
	ran := 0
	g.WithWriteLock(func() {
		err := g.SearchLibraryThenForEffect(SearchLibrarySpec{
			Player:  p.ID,
			Pred:    func(c Card) bool { return c.Name == name },
			Dest:    dest,
			Limit:   1,
			Shuffle: true,
			Then:    func(_ *Game, ids []uuid.UUID) error { ran++; found = ids; return nil },
		})
		if err != nil {
			t.Fatalf("SearchLibraryThenForEffect: %v", err)
		}
	})
	return &found, &ran
}

// TestSearchToGraveyardOpensTheReplacementWindow is the issue. Entomb
// finds the card, and a standing "if a card would be put into a
// graveyard from anywhere, exile it instead" takes it — which it could
// not do at all before, because the take never asked.
func TestSearchToGraveyardOpensTheReplacementWindow(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	target := libraryCard(me, "Reanimation Target")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(target,
			"if it would be put into a graveyard, exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
	})

	found, ran := searchToGraveyard(t, g, me, "Reanimation Target", ZoneGraveyard)

	if *ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", *ran)
	}
	if !g.Exile.Contains(target) {
		t.Fatal(`the tutored card is in exile — "if a card would be put into a graveyard from anywhere"`)
	}
	if me.Graveyard.Contains(target) {
		t.Error("the replacement rewrote the destination, so nothing reached the graveyard")
	}
	if len(*found) != 0 {
		t.Errorf("found = %v, want nothing: the card never reached the destination the search named (CR 400.7)", *found)
	}
	// The search still HAPPENED — it is the destination that was
	// replaced, not the search — so its own event and its shuffle are
	// owed either way.
	if !hasAnyEvent(g, EventSearchLibrary) {
		t.Error("EventSearchLibrary fires even when the found card was taken by a replacement")
	}
}

// TestSearchToGraveyardCancelledLegLeavesTheCardInTheLibrary is
// CR 614.10 with a null replacement: "it isn't put into a graveyard".
// Nothing moves, and the search behind it still finishes.
func TestSearchToGraveyardCancelledLegLeavesTheCardInTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	target := libraryCard(me, "Reanimation Target")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(target,
			"it isn't put into a graveyard", func(ev *ReplacementEvent) { ev.Cancel() }))
	})

	found, ran := searchToGraveyard(t, g, me, "Reanimation Target", ZoneGraveyard)

	if *ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", *ran)
	}
	if !me.Library.Contains(target) {
		t.Error("a cancelled leg moves nothing, so the card is still in the library")
	}
	if len(*found) != 0 {
		t.Errorf("found = %v, want nothing", *found)
	}
}

// TestATutoredCommanderIsOfferedTheCommandZone is CR 903.9's "from
// anywhere" reaching the last two movers that never asked. The search
// cannot finish while the question is open — the found list is not a
// list yet — so the shuffle and the Then wait for the answer.
func TestATutoredCommanderIsOfferedTheCommandZone(t *testing.T) {
	for _, tc := range []struct {
		name        string
		dest        ZoneKind
		commandZone bool
	}{
		{"to the graveyard, declined", ZoneGraveyard, false},
		{"to the graveyard, taken", ZoneGraveyard, true},
		{"to the hand, taken", ZoneHand, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			commander := seatCommander(t, me.Library, me)

			found, ran := searchToGraveyard(t, g, me, "Atraxa", tc.dest)
			if *ran != 0 {
				t.Fatalf("the continuation ran %d times with the CR 903.9 prompt open, want 0", *ran)
			}
			if !me.Library.Contains(commander) {
				t.Fatal("a paused leg has moved nothing: the commander is still in the library")
			}

			prompt := expectCommanderPrompt(t, g, me)
			if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, tc.commandZone); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}

			if *ran != 1 {
				t.Fatalf("the continuation ran %d times after the answer, want 1", *ran)
			}
			want := []uuid.UUID{commander}
			landed := me.Graveyard
			if tc.dest == ZoneHand {
				landed = me.Hand
			}
			if tc.commandZone {
				// It left the library, but not to where the search
				// aimed it, so it was not found this way (CR 400.7).
				want = nil
				landed = me.Command
			}
			if !idsEqual(*found, want) {
				t.Errorf("found = %v, want %v", *found, want)
			}
			if !landed.Contains(commander) {
				t.Errorf("the commander is in the %s", landed.Kind)
			}
		})
	}
}

// TestUndoAcrossAPausedSearchLegReplays is the undo contract every
// continuation in the engine signs: rewind into the open CR 903.9
// prompt, answer again, and the same card is taken and the same list
// reported, because the found list is carried forward by value.
func TestUndoAcrossAPausedSearchLegReplays(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	commander := seatCommander(t, me.Library, me)

	found, ran := searchToGraveyard(t, g, me, "Atraxa", ZoneGraveyard)
	promptOpen := g.Clone()

	prompt := expectCommanderPrompt(t, g, me)
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	first := append([]uuid.UUID(nil), (*found)...)
	if !idsEqual(first, []uuid.UUID{commander}) {
		t.Fatalf("first run found %v, want the commander", first)
	}

	*ran = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if !g.Seats[0].Library.Contains(commander) {
		t.Fatal("the rewind puts the commander back in the library")
	}

	replay := expectCommanderPrompt(t, g, g.Seats[0])
	if err := g.ResolveOptionalReplacement(replay.ID, g.Seats[0].ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement (replay): %v", err)
	}
	if *ran != 1 {
		t.Errorf("the continuation ran %d times on the replay, want 1", *ran)
	}
	if !idsEqual(*found, first) {
		t.Errorf("replayed found = %v, want the first run's %v", *found, first)
	}
}

// --- surveil ------------------------------------------------------

// openSurveil queues a surveil of n and returns the prompt.
func openSurveil(t *testing.T, g *Game, p *Player, n int, after func(g *Game) error) *PendingChoice {
	t.Helper()
	g.WithWriteLock(func() {
		g.SurveilThenForEffect(p.ID, uuid.Nil, n, after)
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceSurveil && c.Chooser == p.ID {
			return c
		}
	}
	t.Fatal("no surveil prompt was queued")
	return nil
}

// TestSurveilGraveyardLegOpensTheReplacementWindow is the other half
// of the issue, and the one that is not a mill: surveil bins by
// choice, from the top of the library, and the card it bins is put
// into a graveyard like any other.
func TestSurveilGraveyardLegOpensTheReplacementWindow(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	// The top of a library is the LAST element, so the last push is
	// the first card surveilled.
	kept := libraryCard(me, "Kept")
	binned := libraryCard(me, "Binned")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(binned,
			"if it would be put into a graveyard, exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
	})

	ran := 0
	c := openSurveil(t, g, me, 2, func(*Game) error { ran++; return nil })
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{binned}, []uuid.UUID{kept}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}

	if ran != 1 {
		t.Fatalf("the rest of the effect ran %d times, want 1", ran)
	}
	if !g.Exile.Contains(binned) {
		t.Fatal("the surveilled card is in exile: the graveyard leg opened the CR 614 window")
	}
	if me.Graveyard.Contains(binned) {
		t.Error("nothing reached the graveyard")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != kept {
		t.Error("the kept card is the top of the library once the binned one has left")
	}
	// The event says how many went to the GRAVEYARD, and none did.
	for _, ev := range g.Events {
		if ev.Kind == EventSurveil && ev.Amount != 0 {
			t.Errorf("EventSurveil amount = %d, want 0 — the card was exiled, not binned", ev.Amount)
		}
	}
	if hasEventFor(g, EventMill, binned) {
		t.Error("EventMill fired for a card a replacement sent to exile instead")
	}
}

// TestSurveilledCommanderIsOfferedTheCommandZone — the leg pauses, and
// the surveil is not finished until it is answered: EventSurveil and
// the rest of the effect both wait.
func TestSurveilledCommanderIsOfferedTheCommandZone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	kept := libraryCard(me, "Kept")
	commander := seatCommander(t, me.Library, me)

	ran := 0
	c := openSurveil(t, g, me, 2, func(*Game) error { ran++; return nil })
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{commander}, []uuid.UUID{kept}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if ran != 0 {
		t.Fatalf("the rest of the effect ran %d times with the CR 903.9 prompt open, want 0", ran)
	}
	if hasAnyEvent(g, EventSurveil) {
		t.Error("the surveil is not finished while a leg is still being asked about")
	}

	prompt := expectCommanderPrompt(t, g, me)
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}

	if ran != 1 {
		t.Fatalf("the rest of the effect ran %d times after the answer, want 1", ran)
	}
	if !me.Command.Contains(commander) {
		t.Error("the commander took CR 903.9's offer and is in the command zone")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != kept {
		t.Error("the kept card is the top of the library")
	}
}

// hasAnyEvent reports whether any event of this kind was emitted, for
// the assertions that are about the kind rather than about one card.
func hasAnyEvent(g *Game, kind EventKind) bool {
	for _, ev := range g.Events {
		if ev.Kind == kind {
			return true
		}
	}
	return false
}
