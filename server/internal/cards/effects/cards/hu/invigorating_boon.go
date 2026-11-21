package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invigorating Boon — Enchantment {1}{G} (Dominaria Remastered, #5948):
//
//	"Whenever a player cycles a card, you may put a +1/+1 counter on target
//	 creature."
//
// Astral Slide's shape: the cycle trigger of any player, with a target
// creature chosen as it goes on the stack and the "may" asked on
// resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d00329c2-e65c-4774-83cd-d68e1d1a6e31",
		Name:         "Invigorating Boon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.On(game.EventCycle, effects.AnyPlayer, "Invigorating Boon — put a +1/+1 counter on target creature",
						effects.PlusOneCounterOnTheTarget),
					"Invigorating Boon — put a +1/+1 counter on target creature?"),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
