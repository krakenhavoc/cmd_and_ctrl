package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undergrowth Recon — Enchantment {1}{G}{G} (Murders at Karlov Manor,
// #4469):
//
//	"At the beginning of your upkeep, return target land card from your
//	 graveyard to the battlefield tapped."
//
// The trigger is targeted, so it is removed without a prompt when your
// graveyard holds no land card (CR 603.3d); a target that has left by
// resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3fbe46fa-88da-439a-994e-c8f1791332a8",
		Name:         "Undergrowth Recon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourUpkeep("Undergrowth Recon — return target land card from your graveyard to the battlefield tapped",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetCardInGraveyard("target land card from your graveyard", effects.Land(), effects.YouOwn())),
		},
	})
}
