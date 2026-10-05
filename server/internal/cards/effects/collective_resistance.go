package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bffdfe7b-f17a-41b6-a460-80280fe497d4",
		Name:         "Collective Resistance",
		Completeness: CompletenessFull,
		Modes: Escalating(ChooseOneOrMore(
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
			ModeDoing("Destroy target enchantment.",
				TargetPermanent("target enchantment", Enchantment()),
				DestroyTheModesTarget),
			ModeDoing("Target creature gains hexproof and indestructible until end of turn.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"hexproof", "indestructible"},
						Label:    "Collective Resistance",
					}.Apply(ctx)
				}),
		), EscalateMana("{G}")),
	})
}
