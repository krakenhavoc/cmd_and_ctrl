package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Misstep — Sorcery {1}{U}:
//
//	"Creatures target player controls don't untap during that player's
//	 next untap step."
//
// #3995's player-wide skip, over creatures: a creature that player casts
// or takes after Misstep resolved stays tapped too (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ee994a1-fd3b-43cf-be3c-9c705e60754a",
		Name:         "Misstep",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayersPermanentsDontUntap(ctx, "Misstep — creatures don't untap during that player's next untap step", "creature")
		},
	})
}
