package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fillHandTo draws until the player's hand holds at least n cards.
func FillHandTo(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < 60 && p.Hand.Size() < n; i++ {
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
	}
	if p.Hand.Size() < n {
		t.Fatalf("hand only reached %d, want %d", p.Hand.Size(), n)
	}
}

// runCleanup walks the cursor to the end step and then takes the one
// AdvanceStep that enters cleanup, which is where
// populateDiscardPendingLocked decides who owes a discard.
func RunCleanup(t *testing.T, g *game.Game) {
	t.Helper()
	AdvanceTo(t, g, game.StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into cleanup: %v", err)
	}
}

// leaveBattlefield sends a battlefield permanent to its owner's
// graveyard.
func LeaveBattlefield(t *testing.T, g *game.Game, owner, cardID uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner},
		cardID,
	); err != nil {
		t.Fatalf("MoveCardByID off battlefield: %v", err)
	}
}
