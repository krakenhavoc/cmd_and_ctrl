package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared with mandate_of_peace_test.go ----------------------------

// instantCardFor is a throwaway printed instant, for asking the cast
// gate a hypothetical question rather than actually casting it.
func InstantCardFor(t *testing.T, g *game.Game) game.Card {
	t.Helper()
	c := game.NewCard("Test Spell", g.Seats[0].ID)
	c.TypeLine = "Instant"
	return c
}

// castGateFor is the *effects*-package window onto
// game.Game.CastGateLocked, for a seam whose proof is a refusal
// rather than a resolved effect.
func CastGateFor(g *game.Game, seat uuid.UUID, card game.Card, zone game.ZoneKind) error {
	var err error
	g.ReadSnapshot(func() { err = g.CastGateLocked(seat, card, zone, game.CastSpellParams{}) })
	return err
}
