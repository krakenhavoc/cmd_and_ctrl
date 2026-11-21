package testsl

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTokenReplacementCard(g, effectstest.LibraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := effectstest.HandCardFull(me, "Pitch", "Sorcery", "{1}", "", nil)
	w := &lengWatcher{}
	g.RegisterListener(w)

	effectstest.LengDiscard(t, g, me, card)
	effectstest.AnswerLengPrompt(t, g, me.ID, true)

	if me.Graveyard.Contains(card) || me.Hand.Contains(card) {
		t.Error("the card did not leave for the library")
	}
	if !me.Library.Contains(card) {
		t.Fatal("the card is not in the library")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != card {
		t.Error("the card is in the library but not on TOP of it")
	}
	// CR 701.9a: it was still discarded, so every "whenever you
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTokenReplacementCard(g, effectstest.LibraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := effectstest.HandCardFull(me, "Pitch", "Sorcery", "{1}", "", nil)

	effectstest.LengDiscard(t, g, me, card)
	effectstest.AnswerLengPrompt(t, g, me.ID, false)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushTokenReplacementCard(g, effectstest.LibraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	theirs := effectstest.HandCardFull(opp, "Their Pitch", "Sorcery", "{1}", "", nil)

	effectstest.LengDiscard(t, g, opp, theirs)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("an opponent's discard queued %d prompt(s), want none", len(g.PendingChoices))
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's discard did not reach their graveyard")
	}
}

// TestLibraryOfLengIsDeclaredFull — the multi-card order is the
// owner's since #2949, so nothing is left out.
func TestLibraryOfLengIsDeclaredFull(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.LibraryOfLengOracle)
	if !ok {
		t.Fatal("Library of Leng is not registered")
	}
	if !spec.NoMaxHandSize {
		t.Error("the first line — no maximum hand size — is missing")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %q with %d caveat(s), want full with none", spec.Completeness, len(spec.Caveats))
	}
}
