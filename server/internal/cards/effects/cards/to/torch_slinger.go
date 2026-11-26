package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torch Slinger — Creature — Goblin Shaman {2}{R}, 2/2 (Duel Decks: Zendikar
// vs. Eldrazi, #5548):
//
//	"Kicker {1}{R}
//	 When this creature enters, if it was kicked, it deals 2 damage to target
//	 creature."
//
// The trigger exists only if the kicker was paid (CR 603.4), so its target is
// asked only then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "64231f03-ea89-4589-8ef8-a87dc9448e7b",
		Name:          "Torch Slinger",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Torch Slinger — kicked, it deals 2 damage to target creature",
					effects.DealDamageToTheTarget(2)),
				effects.TargetCreature("target creature")),
		},
	})
}
