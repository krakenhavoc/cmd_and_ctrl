package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// etChooseCards answers `chooser`'s newest choose_cards prompt with
// ids, after checking it offered `offered` cards.
func EtChooseCards(t *testing.T, g *game.Game, chooser uuid.UUID, offered int, ids ...uuid.UUID) {
	t.Helper()
	ask := LatestChooseCardsFor(g, chooser)
	if ask == nil {
		t.Fatal("no choose_cards prompt")
	}
	if len(ask.ChooseCards) != offered {
		t.Errorf("offered %d cards, want %d", len(ask.ChooseCards), offered)
	}
	if err := g.ResolveChooseCards(ask.ID, chooser, ids); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}
