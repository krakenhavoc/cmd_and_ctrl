package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hearth Charm — Instant {R} (Visions, S82 older sets batch 790, #5631):
//
//	"Choose one —
//	 • Destroy target artifact creature.
//	 • Attacking creatures get +1/+0 until end of turn.
//	 • Target creature with power 2 or less can't be blocked this turn."
//
// The pump is a mass effect whose set is locked as the mode resolves (CR
// 611.2c): the creatures attacking then, anyone's. The power limit on the
// third mode is checked when it is cast and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ce95005-82b3-4921-b1c4-60d775a012a2",
		Name:         "Hearth Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Destroy target artifact creature.",
				effects.TargetCreature("target artifact creature", effects.Artifact()),
				effects.TitanOfIndustryDestroy), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeDoing("Attacking creatures get +1/+0 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.IsAttackingCreature(), Power: 1, Label: "Hearth Charm — +1/+0"}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature with power 2 or less can't be blocked this turn.",
				effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBeBlocked, Label: "Hearth Charm — can't be blocked"}.Apply(ctx)
				}), effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0))),
		),
	})
}
