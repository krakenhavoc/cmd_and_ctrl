package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embrace the Paradox — Instant {3}{G}{U} (Secrets of Strixhaven, #4456):
//
//	"Draw three cards. You may put a land card from your hand onto the
//	 battlefield tapped."
//
// The land question opens only once the three draws have fully resolved
// (H.E.R.B.I.E. Scout Unit's shape), so a land just drawn is on offer.
// Putting a land onto the battlefield is not playing it (the ruling), so
// it works after the turn's land drop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9110be0-3f0c-449b-8baf-689e92fdb8f0",
		Name:         "Embrace the Paradox",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 3},
			effects.MayPutALandFromHandTapped("Embrace the Paradox"),
		),
	})
}
