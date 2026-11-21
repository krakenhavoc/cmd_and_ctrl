package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emerald Charm — Instant {G} (Dominaria Remastered, #5948):
//
//	"Choose one —
//	 • Untap target permanent.
//	 • Destroy target non-Aura enchantment.
//	 • Target creature loses flying until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "904b3db5-70bc-4d49-9887-f504aff8666f",
		Name:         "Emerald Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Untap target permanent.", effects.TargetPermanent("target permanent"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.UntapTarget{Target: t.ID}.Apply(ctx)
				}),
			effects.ModeDoing("Destroy target non-Aura enchantment.",
				effects.TargetPermanent("target non-Aura enchantment", effects.And(effects.Enchantment(), effects.Not(effects.Subtype("Aura")))),
				effects.TitanOfIndustryDestroy),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature loses flying until end of turn.", effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					f, ok := effects.ModeTarget(ctx, occ)
					if !ok || f.Kind != game.TargetCard {
						return nil
					}
					return effects.UntilEndOfTurn(ctx, f.ID, nil, "Emerald Charm — loses flying until end of turn",
						game.RemoveKeywordsMod("flying"))
				}), effects.ForTargets(effects.HarmUntilEOT(0, 0, 0))),
		),
	})
}
