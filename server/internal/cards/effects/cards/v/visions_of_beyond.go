package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Visions of Beyond — Instant {U} (EDHREC rank 3456):
//
//	"Draw a card. If a graveyard has twenty or more cards in it, draw
//	 three cards instead."
//
// Ancestral Recall in a long game. The threshold is any player's
// graveyard — yours, an opponent's, an eliminated player's — read as
// the spell resolves (b33AnyGraveyardHasAtLeast), and the "instead"
// is one draw of one or three, not a draw of one and then two more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c68964df-51ae-43b0-abea-74735016c13f",
		Name:         "Visions of Beyond",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.B33DrawOneOrThreeIfAGraveyardIsFull,
	})
}
