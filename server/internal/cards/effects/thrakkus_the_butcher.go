package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrakkus the Butcher — Legendary Creature — Dragon Peasant {3}{R}{G},
// 3/4 (Commander Legends: Battle for Baldur's Gate, #5490):
//
//	"Trample
//	 Whenever Thrakkus attacks, double the power of each Dragon you
//	 control until end of turn."
//
// Trample is printed. The attack trigger is Double Trouble's body over
// the Dragon creatures you control as it resolves (CR 611.2c), Thrakkus
// among them while he is still there: each gets +X/+0 where X is its
// power then, every X read before any is applied (CR 701.10b). "Dragon"
// is a word of Thrakkus's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f7cd7ede-51c1-4117-99b5-484d0d80c95f",
		Name:            "Thrakkus the Butcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Thrakkus the Butcher — double the power of each Dragon you control until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					dragons := creaturesOfTypeYouControl(g, item.Controller, ctx.TextSubtype("Dragon"))
					return DoublePowerOfEachUntilEOT(ctx, dragons, "Thrakkus the Butcher — double power")
				}),
		},
	})
}
