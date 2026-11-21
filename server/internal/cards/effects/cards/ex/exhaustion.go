package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exhaustion — Sorcery {2}{U}:
//
//	"Creatures and lands target opponent controls don't untap during
//	 their next untap step."
//
// #3995's player-wide skip over two card types: a permanent that is a
// creature or a land (or both) as that opponent's next untap step begins
// stays tapped, whenever it came under their control (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e7b9caf-8285-4386-98bc-9a809827f447",
		Name:         "Exhaustion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayersPermanentsDontUntap(ctx, "Exhaustion — creatures and lands don't untap during that player's next untap step", "creature", "land")
		},
	})
}
