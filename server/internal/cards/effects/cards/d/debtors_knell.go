package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Debtors' Knell — Enchantment {4}{W/B}{W/B}{W/B} (Commander 2021, #5504):
//
//	"At the beginning of your upkeep, put target creature card from a
//	 graveyard onto the battlefield under your control."
//
// The target is a creature card in any graveyard, chosen as the trigger
// goes on the stack (with none, the trigger is removed, CR 603.3d), and it
// comes back under the Knell's controller's control (Reanimate's shape),
// not its owner's. One that left in response makes the ability do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9ff368b-5208-4ada-bb99-eccbf8eaea28",
		Name:         "Debtors' Knell",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourUpkeep("Debtors' Knell — put target creature card from a graveyard onto the battlefield under your control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
						return nil
					}),
				effects.TargetCreatureInAnyGraveyard()),
		},
	})
}
