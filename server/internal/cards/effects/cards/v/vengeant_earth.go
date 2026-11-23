package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vengeant Earth — Instant {1}{G} (March of the Machine):
//
//	"Target creature or land you control becomes a 4/4 Elemental
//	 creature with haste in addition to its other types until end of
//	 turn. It must be blocked this turn if able."
//
// Brave the Wilds' animation (a creature type, haste and a base 4/4,
// layers 4, 6 and 7b) pinned to the object (CR 400.7), for the turn; the
// block requirement is Enlarge's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "237b8528-a199-45f4-8efc-96d67119f738",
		Name:         "Vengeant Earth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or land you control", effects.And(effects.Or(effects.Creature(), effects.Land()), effects.YouControl())),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.ScopedEffectFor{
				Target: t.ID,
				Mods: []game.Mod{
					game.AddTypesMod("Creature"),
					game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
					game.AddKeywordsMod("haste"),
					game.SetBasePowerMod(4),
					game.SetBaseToughnessMod(4),
				},
				Duration: ctx.Game.PinnedTo(ctx.Game.UntilEndOfTurnDuration(), t.ID),
				Label:    "Vengeant Earth — a 4/4 Elemental creature with haste",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.BlockRequirementUntilEOT{
				Target: t.ID,
				Kind:   game.BlockRequirementMustBeBlocked,
				Label:  "Vengeant Earth — must be blocked this turn",
			}.Apply(ctx)
		},
	})
}
