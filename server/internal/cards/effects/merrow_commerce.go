package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merrow Commerce — Kindred Enchantment — Merfolk {1}{U} (Lorwyn, #5593):
//
//	"At the beginning of your end step, untap all Merfolk you control."
//
// Every Merfolk permanent you control as the trigger resolves, a Kindred
// one as well as a creature, read through the layers (a changeling counts,
// CR 702.73a). The set is the one on the battlefield then, and the untap is
// one action over all of them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a931b609-1486-447b-9479-61a45d64f2ef",
		Name:         "Merrow Commerce",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Merrow Commerce — untap all Merfolk you control", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				merfolk := ctx.TextSubtype("Merfolk")
				return b16UntapAllYouControlMatching(ctx, item.Controller, func(c game.Card) bool {
					return c.HasSubtype(merfolk)
				})
			}),
		},
	})
}
