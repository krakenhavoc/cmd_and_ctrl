package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stingerquill Charm — Instant {B}{R} (Reality Fracture, tracker #2795):
//
//	"Choose one —
//	 • Stingerquill Charm deals 3 damage to any target.
//	 • Target creature gains first strike and deathtouch until end of turn.
//	 • Create a 2/2 colorless Wizard Soldier creature token named Cadet.
//	   It gains haste until end of turn."
//
// The haste is an until-end-of-turn grant on the new token, not a
// printed keyword of the template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "987aadac-d6a6-4e50-9468-9e8e53cc5529",
		Name:         "Stingerquill Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Stingerquill Charm deals 3 damage to any target.", effects.TargetAny(),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains first strike and deathtouch until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike", "deathtouch"}, Label: "Stingerquill Charm"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike", "deathtouch"))),
			effects.ModeDoing("Create a 2/2 colorless Wizard Soldier creature token named Cadet. It gains haste until end of turn.",
				nil, effects.RfSpellBCadetWithHaste),
		),
	})
}
