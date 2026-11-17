package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "9a708f82-a4a1-47f9-8ff3-65608269eebd",
		Name:          "Elvish Hydromancer",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{3}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
					"Elvish Hydromancer — kicked, create a token that's a copy of target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						return TokenCopyOfSingleTarget(item, NewContext(g, item))
					}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
