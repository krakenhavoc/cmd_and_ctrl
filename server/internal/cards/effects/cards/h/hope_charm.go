package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hope Charm — Instant {W} (Visions, S82 older sets batch 790, #5631):
//
//	"Choose one —
//	 • Target creature gains first strike until end of turn.
//	 • Target player gains 2 life.
//	 • Destroy target Aura."
//
// Each bullet has its own target group (CR 700.2); a target that is no longer
// legal as the spell resolves does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1b5c83e-cf8b-4644-8315-35387a4f491f",
		Name:         "Hope Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains first strike until end of turn.",
				effects.TargetCreature("target creature"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"}, Label: "Hope Charm — first strike"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike"))),
			effects.ModeWithPurpose(effects.ModeDoing("Target player gains 2 life.", effects.TargetPlayer("target player"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.GainLife{Player: t.ID, Amount: 2}.Apply(ctx)
				}), effects.ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 2})),
			effects.ModeWithPurpose(effects.ModeDoing("Destroy target Aura.", effects.TargetPermanent("target Aura", effects.HasSubtype("Aura")),
				effects.TitanOfIndustryDestroy), game.Purpose{Answers: game.AnswerRemove}),
		),
	})
}
