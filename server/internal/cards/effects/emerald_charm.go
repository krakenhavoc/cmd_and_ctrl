package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emerald Charm — Instant {G} (Dominaria Remastered, #5948):
//
//	"Choose one —
//	 • Untap target permanent.
//	 • Destroy target non-Aura enchantment.
//	 • Target creature loses flying until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "904b3db5-70bc-4d49-9887-f504aff8666f",
		Name:         "Emerald Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Untap target permanent.", TargetPermanent("target permanent"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return UntapTarget{Target: t.ID}.Apply(ctx)
				}),
			ModeDoing("Destroy target non-Aura enchantment.",
				TargetPermanent("target non-Aura enchantment", And(Enchantment(), Not(Subtype("Aura")))),
				titanOfIndustryDestroy),
			ModeWithPurpose(ModeDoing("Target creature loses flying until end of turn.", TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					f, ok := ModeTarget(ctx, occ)
					if !ok || f.Kind != game.TargetCard {
						return nil
					}
					return untilEndOfTurn(ctx, f.ID, nil, "Emerald Charm — loses flying until end of turn",
						game.RemoveKeywordsMod("flying"))
				}), ForTargets(HarmUntilEOT(0, 0, 0))),
		),
	})
}
