package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// departed_tax_test.go — #961 / CR 800.4f at the card level: "if an
// object requires a player who has left the game to pay a cost or
// choose whether to pay a cost, that cost is not paid". The three
// pay-unless cards in the catalog, each asked of a player who then
// concedes with the prompt still open.
//
// The two taxes a SURVIVOR's object charges still pay out — the
// payment is what did not happen, not the consequence. The one a
// player's own permanent charges them does nothing at all: CR 800.4a
// takes that permanent out of the game in the same breath.
//
// The engine-side seam (the departure table's drop action, the sweep
// that runs it, undo) is pinned in internal/game's
// leave_game_pay_unless_test.go.

// sacrificedInLog reports whether the log records `id` being
// sacrificed.
func SacrificedInLog(g *game.Game, id uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventSacrifice && ev.CardID == id {
			return true
		}
	}
	return false
}
