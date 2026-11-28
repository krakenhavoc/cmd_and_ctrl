package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seedling Charm — Instant {G} (Mirage, S82 OLD batch 812, #5633):
//
//	"Choose one —
//	 • Return target Aura attached to a creature to its owner's hand.
//	 • Regenerate target green creature.
//	 • Target creature gains trample until end of turn."
//
// Three modal bullets, each with its own target clause. The Aura must be on a
// creature when it is targeted and again as the spell resolves (Devout
// Harpist's clause, CR 608.2b); the green clause reads the colour word through
// the Charm's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8d66cd6-15db-41ff-bae3-60f1aad3ea23",
		Name:         "Seedling Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target Aura attached to a creature to its owner's hand.",
				effects.BattlefieldSpec("permanent", "target Aura attached to a creature", effects.OldB717AuraOnACreature),
				effects.BounceTheModesTarget),
			effects.ModeDoing("Regenerate target green creature.",
				effects.TargetCreature("target green creature", effects.OfColor("G")),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.Regenerate{Target: t.ID}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains trample until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Seedling Charm — trample"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample"))),
		),
	})
}
