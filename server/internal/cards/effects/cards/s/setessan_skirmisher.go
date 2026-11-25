package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Setessan Skirmisher — Creature — Human Warrior {1}{G}, 2/1:
//
//	"Constellation — Whenever an enchantment you control enters, this creature
//	 gets +1/+1 until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "789d2858-f342-422c-9187-c6ace9485ed5",
		Name:         "Setessan Skirmisher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.GapB103EnchantmentEnteredUnderYourControl,
				"Setessan Skirmisher — gets +1/+1 until end of turn (constellation)",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Setessan Skirmisher — +1/+1 until end of turn")),
		},
	})
}
