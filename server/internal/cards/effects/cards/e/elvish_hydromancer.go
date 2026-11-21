package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Hydromancer — Creature — Elf Wizard {2}{G}, 3/2 (Dominaria United):
//
//	"Kicker {3}{U}
//	 When this creature enters, if it was kicked, create a token that's a
//	 copy of target creature you control."
//
// Kor Sanctifiers' kicked enters trigger (an intervening "if", CR 603.4, so an
// unkicked Hydromancer asks for no target) with Donatello's token copy of the
// target: the copiable values only (CR 707.2). The Hydromancer may target
// itself, and a target gone by resolution makes no token (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9a708f82-a4a1-47f9-8ff3-65608269eebd",
		Name:          "Elvish Hydromancer",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Elvish Hydromancer — kicked, create a token that's a copy of target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						return effects.TokenCopyOfSingleTarget(item, effects.NewContext(g, item))
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
