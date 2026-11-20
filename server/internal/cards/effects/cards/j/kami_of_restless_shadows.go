package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of Restless Shadows — Creature — Spirit {4}{B}, 3/3 (Kamigawa: Neon
// Dynasty, S82):
//
//	"When this creature enters, choose one —
//	 • Return up to one target Ninja or Rogue creature card from your
//	   graveyard to your hand.
//	 • Put target creature card from your graveyard on top of your
//	   library."
//
// A modal enter trigger (Aether Channeler): the first bullet is "up to one"
// so it has no target at all when none is chosen; the second needs a target.
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Kami of Restless Shadows — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Return up to one target Ninja or Rogue creature card from your graveyard to your hand.",
			effects.TargetCardInGraveyard("up to one target Ninja or Rogue creature card from your graveyard",
				effects.YouOwn(), effects.Creature(), effects.Or(effects.HasSubtype("Ninja"), effects.HasSubtype("Rogue"))).WithCount(0, 1),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				for _, t := range ctx.ModeTargets(occ) {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
				}
				return nil
			}),
		effects.ModeDoing("Put target creature card from your graveyard on top of your library.",
			effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				for _, t := range ctx.ModeTargets(occ) {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.PutIntoLibrary{Card: t.ID}.Apply(ctx)
				}
				return nil
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "7e770c08-910e-4394-88c5-79c874750b6a",
		Name:         "Kami of Restless Shadows",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
