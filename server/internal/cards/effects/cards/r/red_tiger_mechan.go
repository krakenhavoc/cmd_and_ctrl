package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Red Tiger Mechan — Artifact Creature — Robot Cat {3}{R}, 3/3 (Edge of
// Eternities, #4461):
//
//	"Haste
//	 Warp {1}{R}"
//
// A hasty body with the shared warp cost (CR 702.185a): warped, it
// attacks the turn it arrives and is exiled at the next end step, to be
// cast again from exile on a later turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "94c9aea7-564d-4fc8-ae14-7b22cba18b95",
		Name:             "Red Tiger Mechan",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"haste"},
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{1}{R}")},
	})
}
