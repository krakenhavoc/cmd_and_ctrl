package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func MoveToGraveyard(t *testing.T, g *game.Game, owner, card uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, card,
	); err != nil {
		t.Fatalf("move to graveyard: %v", err)
	}
}
