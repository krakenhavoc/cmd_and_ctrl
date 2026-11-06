package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sami's Curiosity — Sorcery {G} (Edge of Eternities, #4461):
//
//	"You gain 2 life. Create a Lander token. (It's an artifact with "{2},
//	 {T}, Sacrifice this token: Search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle.")"
//
// Cram Session's life gain followed by Galactic Wayfarer's Lander
// (b16LanderToken), in the printed order. The search ability rides the
// token template.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f532b7f-cb76-4be7-b361-0106c561dcb2",
		Name:         "Sami's Curiosity",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve:    ResolveSteps(GainLife{Amount: 2}, CreateToken{Template: b16LanderToken(), N: 1}),
	})
}
