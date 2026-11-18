package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	trigger := WhenThisEnters("Kami of Restless Shadows — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Return up to one target Ninja or Rogue creature card from your graveyard to your hand.",
			TargetCardInGraveyard("up to one target Ninja or Rogue creature card from your graveyard",
				YouOwn(), Creature(), Or(HasSubtype("Ninja"), HasSubtype("Rogue"))).WithCount(0, 1),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				for _, t := range ctx.ModeTargets(occ) {
					if t.Kind != game.TargetCard {
						continue
					}
					return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
				}
				return nil
			}),
		ModeDoing("Put target creature card from your graveyard on top of your library.",
			TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				for _, t := range ctx.ModeTargets(occ) {
					if t.Kind != game.TargetCard {
						continue
					}
					return PutIntoLibrary{Card: t.ID}.Apply(ctx)
				}
				return nil
			}),
	)
	Register(Spec{
		OracleID:     "7e770c08-910e-4394-88c5-79c874750b6a",
		Name:         "Kami of Restless Shadows",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
