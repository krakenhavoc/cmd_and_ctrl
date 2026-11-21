package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Triumph of Ferocity — Enchantment {2}{G}:
//
//	"At the beginning of your upkeep, draw a card if you control the creature
//	 with the greatest power or tied for the greatest power."
//
// The "if" is in the effect, not an intervening if-clause, so it is read as
// the ability resolves, with the same helper Thickest in the Thicket uses
// (controlsGreatestPower: layered power of every creature on the battlefield,
// negative power included; with no creatures nobody has the greatest).
func init() {
	Register(Spec{
		OracleID:     "eabcbaca-6df9-455a-85e5-32d6ad4d1b44",
		Name:         "Triumph of Ferocity",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Triumph of Ferocity — draw a card if you control the creature with the greatest power or tied for the greatest power",
				Do(If{
					Cond: CondBoard{Name: "you control the creature with the greatest power or tied for it", Test: func(ctx *Context) bool {
						return controlsGreatestPower(ctx.Game, ctx.Controller())
					}},
					Then: []Applier{DrawCards{N: 1}},
				})),
		},
	})
}
