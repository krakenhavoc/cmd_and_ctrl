package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tiger-Dillo — Creature — Cat Armadillo {1}{R}, 4/3 (Avatar: The Last
// Airbender, #4459):
//
//	"This creature can't attack or block unless you control another
//	 creature with power 4 or greater."
//
// The condition reads another creature's FINISHED power, which a layer-6
// static cannot see (layer 7 has not run), so the static only records
// the threshold and the engine folds it into CantAttack | CantBlock once
// every layer has run (CR 613.11, game/power_floor.go), as Technodrome's
// own-power floor does. An anthem or counters that lift another creature
// to 4 free it; Tiger-Dillo's own power never counts. Like every
// restriction it is checked at declaration only (CR 508.1c, 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6e82d4f-4fbf-495c-bebd-982d2a855ad0",
		Name:         "Tiger-Dillo",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature can't attack or block unless you control another creature with power 4 or greater.",
			Layer:     game.Layer6Ability,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if c.AttackBlockNeedsOtherPower < 4 {
					c.AttackBlockNeedsOtherPower = 4
				}
			},
		}},
	})
}
