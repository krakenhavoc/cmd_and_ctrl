package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6a381297-ebce-403c-b1dc-b7405fa6902c",
		Name:         "Disturbed Slumber",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "reach", "haste")),
		Targets:      TargetPermanent("target land you control", And(Land(), YouControl())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			land, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (ScopedEffectFor{
				Target: land,
				Mods: []game.Mod{
					game.AddTypesMod("Creature"),
					game.AddSubtypesMod(ctx.TextSubtypes("Dinosaur")...),
					game.SetBasePowerMod(4),
					game.SetBaseToughnessMod(4),
					game.AddKeywordsMod("reach", "haste"),
				},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Disturbed Slumber — a 4/4 Dinosaur creature with reach and haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return BlockRequirementUntilEOT{
				Target: land,
				Kind:   game.BlockRequirementMustBeBlocked,
				Label:  "Disturbed Slumber — must be blocked this turn",
			}.Apply(ctx)
		},
	})
}
