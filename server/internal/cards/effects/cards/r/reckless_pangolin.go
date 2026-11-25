package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckless Pangolin — Creature — Pangolin {2}{G}, 2/2 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	"Whenever this creature attacks, it gets +1/+1 until end of turn."
//
// Benalish Veteran's attack trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f384f19-7fb4-4120-8001-5b401b4debbd",
		Name:         "Reckless Pangolin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Reckless Pangolin — it gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Reckless Pangolin — +1/+1")),
		},
	})
}
