package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "dcc29e59-341c-4175-aa54-f535e2c3e355",
		Name:         "Sokka's Charge",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			avB27AlliesYouControlHaveDuringYourTurn("double strike"),
			avB27AlliesYouControlHaveDuringYourTurn("lifelink"),
		},
	})
}
