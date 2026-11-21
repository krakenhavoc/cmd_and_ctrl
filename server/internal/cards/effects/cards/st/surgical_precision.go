package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surgical Precision — Sorcery {1}{W} (Reality Fracture, tracker #2795):
//
//	"Choose one —
//	 • Destroy target creature with toughness 4 or greater. You gain 1 life.
//	 • You draw a card and gain 2 life."
//
// The life is gained whether or not the creature was destroyed (an
// indestructible one survives), as printed; a target that has gone makes
// the bullet do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0caccec3-4942-499a-9434-0c94389afcf8",
		Name:         "Surgical Precision",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			// The life waits for a prompt the destruction or the draw asks (a replacement
			// offer; CR 608.2c, #4155). The first bullet's destruction is skipped when its
			// target has gone, the life is not (CR 608.2b).
			effects.ModeSteps("Destroy target creature with toughness 4 or greater. You gain 1 life.",
				effects.TargetCreature("target creature with toughness 4 or greater", effects.ToughnessGE(4)),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok {
						return nil
					}
					// The target is gone once it is destroyed, so the life step reads
					// from the record whether the target was there to be destroyed.
					ctx.RecordStep(surgicalPrecisionRecord, nil, 1)
					return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					if len(ctx.StepRecord(surgicalPrecisionRecord).Ns) == 0 {
						return nil
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
				})),
			effects.ModeSteps("You draw a card and gain 2 life.", nil,
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
				})),
		),
	})
}

// surgicalPrecisionRecord files that the first bullet's target was legal when
// the spell resolved, for the life gained after the destruction (#4155).
const surgicalPrecisionRecord = "surgical-precision-target"
