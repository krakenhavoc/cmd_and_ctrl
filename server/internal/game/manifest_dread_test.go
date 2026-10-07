package game

import (
	"testing"

	"github.com/google/uuid"
)

// manifest_dread_test.go — CR 701.62a, ADR 0082's 2026-10-07 amendment
// (#2570): the look at two, the controller-only choice, the manifest of
// the CHOSEN card and the graveyard for the other.

// dreadLibrary stacks a creature over a land on seat 0's library, the
// creature on top, and returns (top, second).
func dreadLibrary(g *Game) (top, second uuid.UUID) {
	me := g.Seats[0]
	second = topOfLibraryFor(me, "Second Card", "Land")
	top = topOfLibraryFor(me, "Top Card", "Creature — Bear")
	return top, second
}

func runManifestDread(t *testing.T, g *Game, then func(*Game, ManifestDreadResult) error) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ManifestDreadThenForEffect(g.Seats[0].ID, uuid.Nil, then); err != nil {
			t.Fatalf("ManifestDreadThenForEffect: %v", err)
		}
	})
}

// The whole of CR 701.62a: the player is asked, the card they name is
// the one that enters face down, and the other one is in the graveyard.
// Naming the SECOND card is the point — manifest alone could only ever
// take the top.
func TestManifestDreadManifestsTheChosenCardAndGraveyardsTheOther(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	top, second := dreadLibrary(g)

	var got ManifestDreadResult
	calls := 0
	runManifestDread(t, g, func(_ *Game, res ManifestDreadResult) error {
		calls++
		got = res
		return nil
	})
	if calls != 0 {
		t.Fatal("the continuation ran before the controller chose")
	}
	answerChooseCards(t, g, me.ID, second)

	if calls != 1 {
		t.Fatalf("continuation ran %d times, want 1", calls)
	}
	c, zone, ok := cardAnywhere(g, second)
	if !ok || zone != ZoneBattlefield || !c.FaceDown || c.FaceDownKind != FaceDownManifested {
		t.Fatalf("chosen card: zone %q face down %v kind %q, want a manifested permanent", zone, c.FaceDown, c.FaceDownKind)
	}
	if _, zone, _ := cardAnywhere(g, top); zone != ZoneGraveyard {
		t.Errorf("the other card is in %q, want the graveyard", zone)
	}
	if got.Manifested != second || got.Graveyarded != top || got.Player != me.ID {
		t.Errorf("result = %+v, want manifested %s, graveyarded %s", got, second, top)
	}
	if n := eventsOfKind(g, EventManifestDread); n != 1 {
		t.Errorf("EventManifestDread fired %d times, want 1", n)
	}
}

// CR 701.62a says "look at", and CR 708.5 keeps the manifested card
// private: the pair is known to the looker alone, and so is the card
// that entered.
func TestManifestDreadIsKnownToItsControllerAlone(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, second := dreadLibrary(g)

	runManifestDread(t, g, nil)
	// While the prompt is open, nobody but the looker knows the pair.
	g.ReadSnapshot(func() {
		for _, id := range chooseCardsPromptFor(g, me.ID).ChooseCards {
			z := g.findCardZoneLocked(id)
			c, _ := g.cardInZoneLocked(z, id)
			if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
				t.Errorf("looked-at card known to me=%v opp=%v, want only the looker", c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID))
			}
		}
	})
	answerChooseCards(t, g, me.ID, second)
	c, _, _ := cardAnywhere(g, second)
	if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Errorf("manifested card known to me=%v opp=%v, want only its controller", c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID))
	}
}

// A library of one card has nothing to choose and nothing to put in
// the graveyard: the card is manifested without a prompt.
func TestManifestDreadWithOneCardAsksNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	only := topOfLibraryFor(me, "Only Card", "Creature — Bear")

	var got ManifestDreadResult
	runManifestDread(t, g, func(_ *Game, res ManifestDreadResult) error { got = res; return nil })

	if chooseCardsPromptFor(g, me.ID) != nil {
		t.Fatal("a one-card look still asked a question")
	}
	if _, zone, _ := cardAnywhere(g, only); zone != ZoneBattlefield {
		t.Errorf("the card is in %q, want the battlefield", zone)
	}
	if got.Manifested != only || got.Graveyarded != uuid.Nil {
		t.Errorf("result = %+v, want the card manifested and nothing in the graveyard", got)
	}
}

// An empty library does nothing — and does not say it did. The
// continuation still runs, so a "then put counters on that creature"
// can find it has none.
func TestManifestDreadOnAnEmptyLibraryDoesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Library.Cards = nil })

	calls := 0
	var got ManifestDreadResult
	runManifestDread(t, g, func(_ *Game, res ManifestDreadResult) error { calls++; got = res; return nil })

	if calls != 1 || got.Manifested != uuid.Nil || got.Graveyarded != uuid.Nil {
		t.Errorf("continuation ran %d times with %+v, want once with an empty result", calls, got)
	}
	if n := eventsOfKind(g, EventManifestDread); n != 0 {
		t.Errorf("EventManifestDread fired %d times on an empty library, want 0", n)
	}
}

// The graveyard card is not a mill: no EventMill for a mill payoff to
// see (CR 701.17a counts cards off the top; this names one).
func TestManifestDreadsGraveyardCardIsNotAMill(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	_, second := dreadLibrary(g)
	runManifestDread(t, g, nil)
	answerChooseCards(t, g, me.ID, second)
	if n := eventsOfKind(g, EventMill); n != 0 {
		t.Errorf("manifest dread fired %d mill events, want 0", n)
	}
}
