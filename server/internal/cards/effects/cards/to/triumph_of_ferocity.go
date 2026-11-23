package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "eabcbaca-6df9-455a-85e5-32d6ad4d1b44",
		Name:         "Triumph of Ferocity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Triumph of Ferocity — draw a card if you control the creature with the greatest power or tied for the greatest power",
				effects.Do(effects.If{
					Cond: effects.CondBoard{Name: "you control the creature with the greatest power or tied for it", Test: func(ctx *effects.Context) bool {
						return effects.ControlsGreatestPower(ctx.Game, ctx.Controller())
					}},
					Then: []effects.Applier{effects.DrawCards{N: 1}},
				})),
		},
	})
}
