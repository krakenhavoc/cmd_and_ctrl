package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triton Waverider — Creature — Merfolk Wizard {3}{U}, 3/3:
//
//	"Constellation — Whenever an enchantment you control enters, this creature
//	 gains flying until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ee66c03-3684-404f-8f2c-dc91e99249bc",
		Name:         "Triton Waverider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.GapB103EnchantmentEnteredUnderYourControl,
				"Triton Waverider — gains flying until end of turn (constellation)",
				effects.ThisGainsKeywordUntilEndOfTurn("flying", "Triton Waverider — flying until end of turn")),
		},
	})
}
