package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disturbed Slumber — Instant {1}{G} (The Lost Caverns of Ixalan, #4470):
//
//	"Until end of turn, target land you control becomes a 4/4 Dinosaur
//	 creature with reach and haste. It's still a land. It must be blocked
//	 this turn if able."
//
// Wrenn and Realmbreaker's +1 for a turn: one ScopedEffectFor record
// pinned to the land (ADR 0041 phase 3), with the creature type added
// (it keeps its land types, so it is still a land), the Dinosaur
// creature type read through the spell's text (CR 612.2), base 4/4, and
// reach and haste, all until end of turn. The block requirement is
// Irresistible Prey's data record on the same land for the same turn
// (BlockRequirementUntilEOT). A land that left before the spell resolves
// is not a legal target, and nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a381297-ebce-403c-b1dc-b7405fa6902c",
		Name:         "Disturbed Slumber",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "reach", "haste")),
		Targets:      effects.TargetPermanent("target land you control", effects.And(effects.Land(), effects.YouControl())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			land, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.ScopedEffectFor{
				Target: land,
				Mods: []game.Mod{
					game.AddTypesMod("Creature"),
					game.AddSubtypesMod(ctx.TextSubtypes("Dinosaur")...),
					game.SetBasePowerMod(4),
					game.SetBaseToughnessMod(4),
					game.AddKeywordsMod("reach", "haste"),
				},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Disturbed Slumber — a 4/4 Dinosaur creature with reach and haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.BlockRequirementUntilEOT{
				Target: land,
				Kind:   game.BlockRequirementMustBeBlocked,
				Label:  "Disturbed Slumber — must be blocked this turn",
			}.Apply(ctx)
		},
	})
}
