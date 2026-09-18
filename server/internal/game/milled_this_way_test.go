package game

import (
	"testing"

	"github.com/google/uuid"
)

// milled_this_way_test.go pins #893: a mill counts what LANDED in the
// zone it was aimed at, the way the destroy sweep has since #815 and
// the exile and bounce sweeps since #866.
//
// MillToZoneForEffect used to report every leg that did not pause as
// having reached `dest`, which is three different wrong answers: a leg
// the CR 614 window cancelled never moved at all, a leg a replacement
// sent somewhere else moved but not this way ("if a card would be put
// into a graveyard from anywhere, exile it instead"), and a commander
// that took CR 903.9's offer went to the command zone. Oona, Queen of
// the Fae reads that list as "exiled this way" and made a Faerie for
// each of them.
//
// The rule is CR 400.7, landedInZoneLocked's: the object that ARRIVED
// in `dest` is the one that was milled this way. The mill is the same
// batch body the other three verbs use, with the mill's own template
// (millRoute) — library origin, the caller's destination, EventMill
// only when a graveyard is where the card really ended up.

// libraryCard pushes a plain card onto the top of p's library and
// returns its ID. The top of a library is the LAST element, so the
// last call is the first card milled.
func libraryCard(p *Player, name string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// intoGraveyardReplacement is "if a card would be put into a graveyard
// from anywhere, <rewrite> instead" (Rest in Peace, Leyline of the Void,
// Stone of Erech), gated to one card. The mill's own destination is
// what it replaces, which is what makes it the counterexample this
// file is about: the card really leaves the library, and it is not
// milled.
func intoGraveyardReplacement(only uuid.UUID, label string, rewrite func(ev *ReplacementEvent)) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == only && ev.NewZone == ZoneGraveyard
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			rewrite(ev)
			return nil
		},
		Label: label,
	}
}

// millThen runs a mill under the write lock and returns the milled
// list the continuation was handed, plus how many times it ran.
func millThen(t *testing.T, g *Game, p *Player, n int, dest ZoneKind) (*[]uuid.UUID, *int) {
	t.Helper()
	var got []uuid.UUID
	ran := 0
	g.WithWriteLock(func() {
		err := g.MillToZoneThenForEffect(p.ID, n, dest, nil, func(_ *Game, milled []uuid.UUID) error {
			ran++
			got = milled
			return nil
		})
		if err != nil {
			t.Fatalf("MillToZoneThenForEffect: %v", err)
		}
	})
	return &got, &ran
}

// TestMilledThisWayCountsOnlyWhatLanded is the issue with no prompt in
// it. Three cards are milled: one reaches the graveyard, one is
// redirected to exile by a graveyard replacement, one has its move
// cancelled outright. Only the first was milled.
func TestMilledThisWayCountsOnlyWhatLanded(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	saved := libraryCard(owner, "Saved")
	stolen := libraryCard(owner, "Stolen")
	milled := libraryCard(owner, "Milled")

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(stolen,
			"if it would be put into a graveyard, exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
		g.RegisterReplacementForTest(intoGraveyardReplacement(saved,
			"it isn't put into a graveyard", func(ev *ReplacementEvent) { ev.Cancel() }))
	})

	got, ran := millThen(t, g, owner, 3, ZoneGraveyard)

	if *ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", *ran)
	}
	if !idsEqual(*got, []uuid.UUID{milled}) {
		t.Errorf("milled this way = %v, want just the card that reached the graveyard (%v) — "+
			"CR 400.7, the object that arrived", *got, milled)
	}
	if !owner.Graveyard.Contains(milled) {
		t.Error("the unprotected card is in its owner's graveyard")
	}
	if !g.Exile.Contains(stolen) {
		t.Error("the redirected card is in exile")
	}
	if !owner.Library.Contains(saved) {
		t.Error("the cancelled leg leaves its card on the library")
	}
	// CR 701.17a defines a mill by its destination, so the card that
	// went to exile instead was never milled and no mill payoff may
	// see it.
	if hasEventFor(g, EventMill, stolen) {
		t.Error("EventMill fired for a card a replacement sent to exile instead")
	}
	if !hasEventFor(g, EventMill, milled) {
		t.Error("EventMill fired for the card that really was milled")
	}
}

// TestAPausedCommanderMillLegIsCountedOnlyWhenItLandsInTheGraveyard is
// the CR 903.9 half. The list cannot be taken while the prompt is
// open, so the rest of the mill waits for it; when the answer arrives
// the commander counts only if the graveyard is where it went.
func TestAPausedCommanderMillLegIsCountedOnlyWhenItLandsInTheGraveyard(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{{"to the command zone", true}, {"to the graveyard", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			owner := g.Seats[0]
			// The top of a library is the LAST element, so the
			// commander is milled first and the bear second.
			bear := libraryCard(owner, "Bear")
			commander := seatCommander(t, owner.Library, owner)

			got, ran := millThen(t, g, owner, 2, ZoneGraveyard)

			if *ran != 0 {
				t.Fatalf("the continuation ran %d times with the CR 903.9 prompt open, want 0", *ran)
			}
			if !owner.Library.Contains(bear) {
				t.Error("the rest of the mill waits for the paused leg")
			}

			prompt := expectCommanderPrompt(t, g, owner)
			if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, tc.commandZone); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}

			if *ran != 1 {
				t.Fatalf("the continuation ran %d times after the answer, want 1", *ran)
			}
			want := []uuid.UUID{commander, bear}
			landed := owner.Graveyard
			if tc.commandZone {
				// It left the library, and it was not milled: the
				// command zone is not a graveyard, and CR 701.17a
				// names the graveyard.
				want = []uuid.UUID{bear}
				landed = owner.Command
			}
			if !idsEqual(*got, want) {
				t.Errorf("milled this way = %v, want %v", *got, want)
			}
			if !landed.Contains(commander) {
				t.Errorf("the commander is in the %s", landed.Kind)
			}
			if !owner.Graveyard.Contains(bear) {
				t.Error("the rest of the mill lands once the prompt is answered")
			}
			if tc.commandZone && hasEventFor(g, EventMill, commander) {
				t.Error("EventMill fired for a commander that went to the command zone")
			}
		})
	}
}

// TestUndoAcrossAPausedMillLegReplays is the undo contract the
// continuation signs, the one DestroyPermanentsThenForEffect and
// ExileCardsThenForEffect already sign: rewinding into the open
// CR 903.9 prompt and answering again mills the same cards and reports
// the same list, because the landed list is carried forward by value.
func TestUndoAcrossAPausedMillLegReplays(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	bear := libraryCard(owner, "Bear")
	commander := seatCommander(t, owner.Library, owner)

	got, ran := millThen(t, g, owner, 2, ZoneGraveyard)
	promptOpen := g.Clone()

	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	first := append([]uuid.UUID(nil), (*got)...)
	if !idsEqual(first, []uuid.UUID{commander, bear}) {
		t.Fatalf("first run milled %v, want both", first)
	}

	*ran = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if !g.Seats[0].Library.Contains(commander) || !g.Seats[0].Library.Contains(bear) {
		t.Fatal("the rewind puts both cards back on the library")
	}

	replay := expectCommanderPrompt(t, g, g.Seats[0])
	if err := g.ResolveOptionalReplacement(replay.ID, g.Seats[0].ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement (replay): %v", err)
	}
	if *ran != 1 {
		t.Errorf("the continuation ran %d times on the replay, want 1", *ran)
	}
	if !idsEqual(*got, first) {
		t.Errorf("replayed milled this way = %v, want the first run's %v", *got, first)
	}
}

// TestFireAndForgetMillReturnsWhatLanded — the synchronous entry point
// keeps its slice and its signature, and the slice now means what the
// sweeps' counts mean: the legs that landed in `dest`. It also keeps
// the one thing the Then form gives up, and #529 put there on purpose:
// a paused card does not hold the rest of the mill up, because nothing
// is waiting on the answer.
func TestFireAndForgetMillReturnsWhatLanded(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	stolen := libraryCard(owner, "Stolen")
	bear := libraryCard(owner, "Bear")
	commander := seatCommander(t, owner.Library, owner)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(stolen,
			"if it would be put into a graveyard, exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
	})

	var moved []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		moved, err = g.MillToZoneForEffect(owner.ID, 3, ZoneGraveyard, nil)
		if err != nil {
			t.Fatalf("MillToZoneForEffect: %v", err)
		}
	})

	if !idsEqual(moved, []uuid.UUID{bear}) {
		t.Errorf("MillToZoneForEffect = %v, want just the bear — the commander is still being asked "+
			"and the third card was exiled instead of milled", moved)
	}
	if !owner.Graveyard.Contains(bear) || !g.Exile.Contains(stolen) {
		t.Error("the mill proceeds AROUND the paused card rather than stopping on it")
	}
	if !owner.Library.Contains(commander) {
		t.Error("the paused leg has moved nothing")
	}

	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !owner.Graveyard.Contains(commander) {
		t.Error("the answered leg lands afterwards, outside the returned list")
	}
}

// TestMillToExileCountsWhatReachedExile is Oona's destination: "exile
// the top X cards of their library" is not a mill (no EventMill, no
// mill payoff), and "exiled this way" is the same CR 400.7 reading
// against exile.
func TestMillToExileCountsWhatReachedExile(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	bear := libraryCard(owner, "Bear")
	commander := seatCommander(t, owner.Library, owner)

	got, ran := millThen(t, g, owner, 2, ZoneExile)
	if *ran != 0 {
		t.Fatalf("the continuation ran %d times with the CR 903.9 prompt open, want 0", *ran)
	}

	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}

	if !idsEqual(*got, []uuid.UUID{bear}) {
		t.Errorf("exiled this way = %v, want just %v — the commander went to the command zone", *got, bear)
	}
	if !owner.Command.Contains(commander) || !g.Exile.Contains(bear) {
		t.Error("the commander is in the command zone and the bear in exile")
	}
	if hasEventFor(g, EventMill, bear) {
		t.Error(`"exile the top N cards" is not a mill (CR 701.17a), so EventMill must not fire for it`)
	}
}
