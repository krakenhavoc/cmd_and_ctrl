package game

import (
	"testing"

	"github.com/google/uuid"
)

// scry_size_test.go — #1036, the engine half.
//
// A finished scry carries TWO numbers and they are different numbers:
// `Amount` is how many cards moved (to the bottom, to the graveyard),
// which is the motion the table watched, and `LookedAt` is the size of
// the action, the "2" in "scry 2" that the player announced. Until
// #1036 only the first was on the event, so the log could say what
// moved and never how big the look was.
//
// The size is the size the action ACTUALLY had, which is the reason it
// is read off the prompt rather than off the call: the CR 614
// keyword-action window (#976) may have rewritten the count before the
// prompt was queued, and a short library clamps it. The tests below
// pin all three — the plain case, the replaced case, the clamped case
// — plus the fact that `Amount` still means what
// cards/effects/surveil_test.go and graveyard_arrival_route_test.go
// say it means.

// openScry queues a scry of n and returns the prompt. The sibling of
// openSurveil in graveyard_arrival_route_test.go.
func openScry(t *testing.T, g *Game, p *Player, n int) *PendingChoice {
	t.Helper()
	g.WithWriteLock(func() {
		g.ScryForEffect(p.ID, uuid.Nil, n)
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceScry && c.Chooser == p.ID {
			return c
		}
	}
	t.Fatal("no scry prompt was queued")
	return nil
}

// onlyEventOfKind returns the one event of `kind` in the log, failing
// when there is not exactly one — both numbers are on a single
// announcement per finished action, and two would mean the emit moved.
func onlyEventOfKind(t *testing.T, g *Game, kind EventKind) Event {
	t.Helper()
	var found []Event
	for _, ev := range g.Events {
		if ev.Kind == kind {
			found = append(found, ev)
		}
	}
	if len(found) != 1 {
		t.Fatalf("%d %s events in the log, want exactly 1", len(found), kind)
	}
	return found[0]
}

// The floor: a scry 2 that bottoms one card says both numbers.
func TestFinishedScryCarriesItsSizeAndItsMotion(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	kept := libraryCard(me, "Kept")
	bottomed := libraryCard(me, "Bottomed")

	c := openScry(t, g, me, 2)
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{bottomed}, []uuid.UUID{kept}); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}

	ev := onlyEventOfKind(t, g, EventScry)
	if ev.LookedAt != 2 {
		t.Errorf("EventScry looked_at = %d, want 2 — the size of the scry", ev.LookedAt)
	}
	if ev.Amount != 1 {
		t.Errorf("EventScry amount = %d, want 1 — one card went to the bottom", ev.Amount)
	}
}

// A scry that keeps everything still announces its size: "scry 2" is
// what happened even when nothing moved, and the size is the half an
// opponent cannot otherwise infer from a motionless scry.
func TestAScryThatMovedNothingStillCarriesItsSize(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	libraryCard(me, "A")
	libraryCard(me, "B")

	c := openScry(t, g, me, 2)
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}

	ev := onlyEventOfKind(t, g, EventScry)
	if ev.LookedAt != 2 || ev.Amount != 0 {
		t.Errorf("EventScry looked_at/amount = %d/%d, want 2/0", ev.LookedAt, ev.Amount)
	}
}

// The point of reading the size off the PROMPT rather than off the
// call: "if you would scry, scry that many plus one instead" rewrites
// the count in the CR 614 window before the prompt is queued
// (keyword_action.go, #976), and the amount scried is the one the
// window settled on — a scry 2 under a Crystal Ball is a scry 3.
func TestAReplacedScryReportsTheAmountActuallyScried(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	for _, n := range []string{"A", "B", "C"} {
		libraryCard(me, n)
	}

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionScry, plusOneForTest, "scry one more"))
	})
	c := openScry(t, g, me, 2)
	if len(c.ScryCards) != 3 {
		t.Fatalf("the prompt offers %d cards, want 3 — the window settled on 2+1", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards[:1], c.ScryCards[1:]); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}

	ev := onlyEventOfKind(t, g, EventScry)
	if ev.LookedAt != 3 {
		t.Errorf("EventScry looked_at = %d, want 3 — the size after the replacement, not the 2 the card printed", ev.LookedAt)
	}
	if ev.Amount != 1 {
		t.Errorf("EventScry amount = %d, want 1", ev.Amount)
	}
}

// The other adjustment: CR 701.22a looks at as many cards as there
// are. A scry 3 into a two-card library is a scry of 2, and saying
// "scried 3" over two cards would be a number the table can see is
// wrong.
func TestAScryIntoAShortLibraryReportsWhatItLookedAt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	libraryCard(me, "A")
	libraryCard(me, "B")

	c := openScry(t, g, me, 3)
	if len(c.ScryCards) != 2 {
		t.Fatalf("the prompt offers %d cards, want the whole 2-card library", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}

	if ev := onlyEventOfKind(t, g, EventScry); ev.LookedAt != 2 {
		t.Errorf("EventScry looked_at = %d, want 2 — the library had no third card", ev.LookedAt)
	}
}

// Surveil is the same two numbers with the other away lane.
func TestFinishedSurveilCarriesItsSizeAndItsMotion(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	kept := libraryCard(me, "Kept")
	binned := libraryCard(me, "Binned")

	c := openSurveil(t, g, me, 2, nil)
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{binned}, []uuid.UUID{kept}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}

	ev := onlyEventOfKind(t, g, EventSurveil)
	if ev.LookedAt != 2 {
		t.Errorf("EventSurveil looked_at = %d, want 2 — the size of the surveil", ev.LookedAt)
	}
	if ev.Amount != 1 {
		t.Errorf("EventSurveil amount = %d, want 1 — one card reached the graveyard", ev.Amount)
	}
}

// The surveil emit lives in the mill route's CONTINUATION (#931), so
// the two numbers come from two different places and can disagree:
// `Amount` is what the route actually put in the graveyard and moves
// when a Rest in Peace exiles a leg instead, while the size of the
// surveil is the whole looked-at set and does not.
func TestASurveilExiledByAReplacementKeepsItsSize(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	kept := libraryCard(me, "Kept")
	binned := libraryCard(me, "Binned")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(binned,
			"if it would be put into a graveyard, exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
	})

	c := openSurveil(t, g, me, 2, nil)
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{binned}, []uuid.UUID{kept}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}

	ev := onlyEventOfKind(t, g, EventSurveil)
	if ev.LookedAt != 2 {
		t.Errorf("EventSurveil looked_at = %d, want 2 — the surveil was still a surveil 2", ev.LookedAt)
	}
	if ev.Amount != 0 {
		t.Errorf("EventSurveil amount = %d, want 0 — the card was exiled, not binned", ev.Amount)
	}
}

// And the replaced count reaches the surveil emit too, through the
// same prompt.
func TestAReplacedSurveilReportsTheAmountActuallySurveilled(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	for _, n := range []string{"A", "B"} {
		libraryCard(me, n)
	}

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionSurveil, plusOneForTest, "surveil one more"))
	})
	c := openSurveil(t, g, me, 1, nil)
	if len(c.ScryCards) != 2 {
		t.Fatalf("the prompt offers %d cards, want 2 — the window settled on 1+1", len(c.ScryCards))
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}

	if ev := onlyEventOfKind(t, g, EventSurveil); ev.LookedAt != 2 {
		t.Errorf("EventSurveil looked_at = %d, want 2 — the size after the replacement", ev.LookedAt)
	}
}
