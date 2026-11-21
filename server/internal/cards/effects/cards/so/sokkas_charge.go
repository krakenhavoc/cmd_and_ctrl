package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sokka's Charge — Enchantment {3}{W} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"During your turn, Allies you control have double strike and
//	 lifelink."
//
// Two layer-6 keyword grants over the Ally creatures the enchantment's
// controller controls, each re-read every layer pass, so an Ally that
// enters mid-turn has both at once and every Ally loses both as the
// turn passes to another player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcc29e59-341c-4175-aa54-f535e2c3e355",
		Name:         "Sokka's Charge",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.AvB27AlliesYouControlHaveDuringYourTurn("double strike"),
			effects.AvB27AlliesYouControlHaveDuringYourTurn("lifelink"),
		},
	})
}
