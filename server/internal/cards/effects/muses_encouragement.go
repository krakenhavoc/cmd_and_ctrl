package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Muse's Encouragement — Instant {4}{U} (Secrets of Strixhaven, #4456):
//
//	"Create a 3/3 blue and red Elemental creature token with flying.
//	 Surveil 2."
//
// The token first, then the surveil, in printed order. The surveil is the
// shared Surveil primitive (look at the top two, any to the graveyard, the
// rest back in any order).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62421084-96e6-40dd-8d34-7c8c1091d4a4",
		Name:         "Muse's Encouragement",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("3/3 blue and red Elemental with flying"), N: 1}, Surveil{N: 2}),
	})
}
