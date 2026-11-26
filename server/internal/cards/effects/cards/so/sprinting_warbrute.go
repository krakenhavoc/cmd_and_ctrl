package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sprinting Warbrute — Creature — Ogre Berserker {4}{R}, 5/4 (Dragons of
// Tarkir, S82 old batch 331):
//
//	"This creature attacks each combat if able.
//	 Dash {3}{R} (You may cast this spell for its dash cost. If you do, it
//	 gains haste, and it's returned from the battlefield to its owner's hand at
//	 the beginning of the next end step.)"
//
// Barricade Breaker's AttacksEachCombat requirement and the shared dash cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e680f3e9-5c76-41e6-a4a5-662ee05ac3af",
		Name:             "Sprinting Warbrute",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{3}{R}")},
		Static:           []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
