package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_of_leng_order_test.go — #2949, ADR 0088 amendment 2026-10-14:
// the cards of one discard that Library of Leng puts on top of the
// library are put there at the same time, so their owner orders them
// (CR 401.4), and the rest of the effect waits for that order.

// lengBoard is Library of Leng under seat 0, an empty hand, a library of
// filler, and n cards in hand to discard.
func LengBoard(t *testing.T, n int) (*game.Game, *game.Player, []uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[0]
	PushTokenReplacementCard(g, LibraryOfLengOracle, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	AwbPadLibrary(me)
	var cards []uuid.UUID
	for i := 0; i < n; i++ {
		cards = append(cards, HandCardFull(me, "Pitch", "Sorcery", "{1}", "", nil))
	}
	return g, me, cards
}

// lengDiscardThen is lengDiscard with a "then" that counts its runs.
func LengDiscardThen(t *testing.T, g *game.Game, who *game.Player, ran *int, cards ...uuid.UUID) {
	t.Helper()
	var promptID uuid.UUID
	// #4361 S7b: the "then" is a keyed discard-then.
	then := game.RegisterDiscardThen("test/4361-s7b-leng-"+uuid.NewString(), func(*game.Game, game.DiscardsDone) error { *ran++; return nil })
	g.WithWriteLock(func() {
		if err := g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{Player: who.ID, N: len(cards)}, then, nil); err != nil {
			t.Fatalf("PlayerDiscardsThenKeyedForEffect: %v", err)
		}
		if n := len(g.PendingChoices); n > 0 {
			promptID = g.PendingChoices[n-1].ID
		}
	})
	if err := g.ResolveChooseCards(promptID, who.ID, cards); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

func PutInLibraryPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePutInLibrary {
			return c
		}
	}
	return nil
}

func LengLibraryTop(p *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := len(p.Library.Cards) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, p.Library.Cards[i].InstanceID)
	}
	return out
}
