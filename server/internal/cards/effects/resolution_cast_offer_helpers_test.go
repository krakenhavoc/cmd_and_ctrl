package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_cast_offer_helpers_test.go — #4250, part 2: the hand-cast
// offers and the one-off grants ask the resolution_cast prompt inside the
// resolution that offers the cast (CR 608.2g). These helpers find and
// answer it.

// resolutionCastOfferFor is chooser's open resolution_cast prompt, or nil.
func resolutionCastOfferFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// mustResolutionCastOffer fails unless chooser owes a resolution_cast
// under keyword, held inside the resolution that made it, offering every
// card of want, and returns it.
func mustResolutionCastOffer(t *testing.T, g *game.Game, chooser uuid.UUID, keyword string, want ...uuid.UUID) *game.PendingChoice {
	t.Helper()
	c := resolutionCastOfferFor(g, chooser)
	if c == nil {
		t.Fatalf("no resolution_cast offer is waiting (pending: %d)", len(g.PendingChoices))
	}
	if c.MayCastKeyword != keyword {
		t.Fatalf("the offer's keyword is %q, want %q", c.MayCastKeyword, keyword)
	}
	if !g.ResolutionPaused() {
		t.Fatal("the offer is not inside the resolution that made it")
	}
	for _, id := range want {
		if !slices.Contains(c.DiscardOptions, id) {
			t.Fatalf("card %s is not offered: %v", id, c.DiscardOptions)
		}
	}
	return c
}

// castThroughResolutionCast casts card through chooser's open
// resolution_cast prompt.
func castThroughResolutionCast(t *testing.T, g *game.Game, chooser, card uuid.UUID, params game.CastSpellParams) error {
	t.Helper()
	c := resolutionCastOfferFor(g, chooser)
	if c == nil {
		t.Fatal("no resolution_cast offer to cast through")
	}
	if params.FromZone == "" {
		params.FromZone = string(offeredZone(g, chooser, card))
	}
	return g.CastForResolutionChoice(chooser, c.ID, card, params)
}

// declineResolutionCast answers chooser's open resolution_cast prompt
// "cast no more of them".
func declineResolutionCast(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	c := resolutionCastOfferFor(g, chooser)
	if c == nil {
		t.Fatal("no resolution_cast offer to decline")
	}
	if err := g.ResolvePendingChoice(c.ID, chooser, nil); err != nil {
		t.Fatalf("declining the offer: %v", err)
	}
}

// castInsideTheResolution fails unless spell was cast before the
// resolution that offered it finished: its EventCast is followed by an
// EventResolutionFinished, the held resolution's last line (CR 608.2g:
// the spell is cast while the other object is resolving).
func castInsideTheResolution(t *testing.T, g *game.Game, spell uuid.UUID) {
	t.Helper()
	cast := -1
	for i, ev := range g.Events {
		if ev.Kind == game.EventCast && ev.CardID == spell {
			cast = i
		}
	}
	if cast < 0 {
		t.Fatalf("spell %s was never cast", spell)
	}
	for _, ev := range g.Events[cast+1:] {
		if ev.Kind == game.EventResolutionFinished {
			return
		}
	}
	t.Fatal("the spell was cast after the resolution that offered it had finished")
}
