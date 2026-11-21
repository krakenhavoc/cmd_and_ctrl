package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collective Resistance — Instant {1}{G}:
//
//	"Escalate {G} (Pay this cost for each mode chosen beyond the
//	 first.)
//	 Choose one or more —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Target creature gains hexproof and indestructible until end of
//	   turn."
//
// Escalate (CR 702.120a, #2126) is mana here: {G} joins the total at
// CR 601.2f once per mode beyond the first, so all three modes cost
// {1}{G}{G}{G}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bffdfe7b-f17a-41b6-a460-80280fe497d4",
		Name:         "Collective Resistance",
		Completeness: effects.CompletenessFull,
		Modes: effects.Escalating(effects.ChooseOneOrMore(
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.DestroyTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains hexproof and indestructible until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"hexproof", "indestructible"},
						Label:    "Collective Resistance",
					}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof", "indestructible"))),
		), effects.EscalateMana("{G}")),
	})
}
