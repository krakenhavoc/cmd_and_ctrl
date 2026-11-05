package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serum Visions — Sorcery {U} (EDHREC rank 888):
//
//	"Draw a card. Scry 2."
//
// Preordain backwards: the draw happens first, then the scry looks
// at the two cards below it. That order is the whole difference
// between the two cards, and it is why the Scry primitive here has
// no `Then` — nothing is printed after the scry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56956afd-db53-4542-816b-490c8b0bbcf7",
		Name:         "Serum Visions",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The scry waits for a prompt the draw asks (CR 608.2c, #4155).
			DrawCards{N: 1},
			Scry{N: 2},
		),
	})
}
