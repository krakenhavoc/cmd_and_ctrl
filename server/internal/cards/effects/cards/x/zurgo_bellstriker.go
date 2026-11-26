package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zurgo Bellstriker — Legendary Creature — Orc Warrior {R}, 2/2 (Dragons of
// Tarkir, S82 old batch 331):
//
//	"Zurgo can't block creatures with power 2 or greater.
//	 Dash {1}{R} (You may cast this spell for its dash cost. If you do, it
//	 gains haste, and it's returned from the battlefield to its owner's hand at
//	 the beginning of the next end step.)"
//
// Cyclops Tyrant's block rule on its own side with the comparison flipped
// (power 2 or greater, read as the block is declared), and the shared dash
// cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "46808926-d95b-4313-89d1-6bdfe13fcb77",
		Name:             "Zurgo Bellstriker",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{1}{R}")},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.And(effects.Creature(), effects.PowerGE(2))),
				"it can't block creatures with power 2 or greater"),
		},
	})
}
