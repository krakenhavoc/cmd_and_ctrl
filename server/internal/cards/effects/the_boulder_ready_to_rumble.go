package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Boulder, Ready to Rumble — Legendary Creature — Human Warrior
// Performer {3}{G}, 4/4 (Avatar: The Last Airbender, #4459):
//
//	"Whenever The Boulder attacks, earthbend X, where X is the number of
//	 creatures you control with power 4 or greater."
//
// The earthbend keyword action with its own target clause, chosen as the
// trigger goes on the stack (CR 603.3d). X is counted as the ability
// resolves (CR 608.2h), with the layers current, so The Boulder counts
// itself while its power is 4 or more, and a land earthbent to 4 power
// earlier counts too. X = 0 still animates the land as a 0/0 (the
// keyword's own rule).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d33ee141-a880-4368-a274-a6ae2f2c624b",
		Name:         "The Boulder, Ready to Rumble",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("The Boulder, Ready to Rumble — earthbend X, where X is the number of creatures you control with power 4 or greater",
				avB14Earthbend(func(ctx *Context) int {
					return msc13CreaturesWithPowerAtLeast(ctx.Game, ctx.Controller(), 4)
				})), EarthbendTargets()),
		},
	})
}
