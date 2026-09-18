package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_of_leng_test.go — the catalog half of #650. The engine's own
// contract for the discard event is pinned in
// game/discard_replacement_test.go; what is here is the card.

const libraryOfLengOracle = "867def48-4be8-4056-bcf1-d6b00450b9a3"

// lengDiscard asks `who` to discard `cards` as an EFFECT's instruction
// (Mind Rot's shape), answering the pick prompt straight away, and
// returns with whatever the CR 614 window queued still open.
func lengDiscard(t *testing.T, g *game.Game, who *game.Player, cards ...uuid.UUID) {
	t.Helper()
	var promptID uuid.UUID
	g.WithWriteLock(func() {
		promptID = g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: who.ID, N: len(cards)})
	})
	if promptID == uuid.Nil {
		t.Fatal("no discard prompt was queued")
	}
	if err := g.ResolveChooseCards(promptID, who.ID, cards); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// answerLengPrompt answers the single queued CR 614.10 "may".
func answerLengPrompt(t *testing.T, g *game.Game, chooser uuid.UUID, apply bool) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want exactly one Library of Leng prompt", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Kind != game.PendingChoiceOptionalReplacement {
		t.Fatalf("prompt kind = %q, want %q", c.Kind, game.PendingChoiceOptionalReplacement)
	}
	if err := g.ResolveOptionalReplacement(c.ID, chooser, apply); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
}

// lengWatcher records the discard events the card is supposed not to
// suppress.
type lengWatcher struct{ discards []game.Event }

func (w *lengWatcher) OnEvent(_ *game.Game, ev game.Event) {
	if ev.Kind == game.EventDiscardCard {
		w.discards = append(w.discards, ev)
	}
}

// TestLibraryOfLengPutsAnEffectDiscardOnTopOfYourLibrary — "discard
// it, BUT you may put it on top of your library instead": the discard
// happened, and only the destination changed.
func TestLibraryOfLengPutsAnEffectDiscardOnTopOfYourLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTokenReplacementCard(g, libraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := handCardFull(me, "Pitch", "Sorcery", "{1}", "", nil)
	w := &lengWatcher{}
	g.RegisterListener(w)

	lengDiscard(t, g, me, card)
	answerLengPrompt(t, g, me.ID, true)

	if me.Graveyard.Contains(card) || me.Hand.Contains(card) {
		t.Error("the card did not leave for the library")
	}
	if !me.Library.Contains(card) {
		t.Fatal("the card is not in the library")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != card {
		t.Error("the card is in the library but not on TOP of it")
	}
	// CR 701.8a: it was still discarded, so every "whenever you
	// discard" payoff still sees it.
	if len(w.discards) != 1 {
		t.Fatalf("EventDiscardCard x %d, want 1", len(w.discards))
	}
	if got := w.discards[0].NewZone; got != game.ZoneLibrary {
		t.Errorf("EventDiscardCard new zone = %q, want %q", got, game.ZoneLibrary)
	}
	if got := w.discards[0].DiscardCause; got != game.DiscardCauseEffect {
		t.Errorf("EventDiscardCard cause = %q, want %q", got, game.DiscardCauseEffect)
	}
}

// TestLibraryOfLengDeclinedSendsTheCardToTheGraveyard — it is a "may",
// so "no" is the printed outcome and not an error.
func TestLibraryOfLengDeclinedSendsTheCardToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTokenReplacementCard(g, libraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := handCardFull(me, "Pitch", "Sorcery", "{1}", "", nil)

	lengDiscard(t, g, me, card)
	answerLengPrompt(t, g, me.ID, false)

	if !me.Graveyard.Contains(card) {
		t.Error("declining did not send the card to the graveyard")
	}
	if me.Library.Contains(card) {
		t.Error("declining still put the card in the library")
	}
}

// TestLibraryOfLengOnlyReplacesItsControllersDiscards — "if an effect
// causes YOU to discard a card". An opponent's discard is untouched,
// and nobody is asked anything.
func TestLibraryOfLengOnlyReplacesItsControllersDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTokenReplacementCard(g, libraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	theirs := handCardFull(opp, "Their Pitch", "Sorcery", "{1}", "", nil)

	lengDiscard(t, g, opp, theirs)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("an opponent's discard queued %d prompt(s), want none", len(g.PendingChoices))
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's discard did not reach their graveyard")
	}
}

// TestLibraryOfLengIsDeclaredWithItsOneCaveat — the ordering clause for
// a multi-card save is the only thing it does not do.
func TestLibraryOfLengIsDeclaredWithItsOneCaveat(t *testing.T) {
	spec, ok := Lookup(libraryOfLengOracle)
	if !ok {
		t.Fatal("Library of Leng is not registered")
	}
	if !spec.NoMaxHandSize {
		t.Error("the first line — no maximum hand size — is missing")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("completeness %q with %d caveat(s), want caveats with exactly 1", spec.Completeness, len(spec.Caveats))
	}
}
