package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3fbe46fa-88da-439a-994e-c8f1791332a8",
		Name:         "Undergrowth Recon",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourUpkeep("Undergrowth Recon — return target land card from your graveyard to the battlefield tapped",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
						}
						return nil
					}),
				TargetCardInGraveyard("target land card from your graveyard", Land(), YouOwn())),
		},
	})
}
