package legal

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrificeAllLabel is the move-label suffix of a cast whose additional
// cost sacrifices all (#2097): the printed clause and every permanent it
// takes, or "none" — " (Sacrifice all creatures you control: Grizzly
// Bears, Llanowar Elves)". The clause is spelled out because a reader of
// the move list (a model seat, an MCP client) must see that the move
// gives up the whole board, not one chosen permanent.
func sacrificeAllLabel(g *game.Game, cost *game.AdditionalCost, ids []uuid.UUID) string {
	names := "none"
	if len(ids) > 0 {
		parts := make([]string, len(ids))
		for i, id := range ids {
			parts[i] = cardName(g, id)
		}
		names = strings.Join(parts, ", ")
	}
	return " (" + cost.Label + ": " + names + ")"
}
