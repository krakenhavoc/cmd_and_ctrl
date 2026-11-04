package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b9110be0-3f0c-449b-8baf-689e92fdb8f0",
		Name:         "Embrace the Paradox",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: ResolveSteps(
			DrawCards{N: 3},
			MayPutALandFromHandTapped("Embrace the Paradox"),
		),
	})
}
