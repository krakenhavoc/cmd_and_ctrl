package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trekB42Pick answers the open choose_cards prompt of `chooser`.
func TrekB42Pick(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	ch := LatestChooseCards(g, chooser)
	if ch == nil {
		t.Fatalf("no choose_cards prompt for %s", chooser)
	}
	if err := g.ResolveChooseCards(ch.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}
