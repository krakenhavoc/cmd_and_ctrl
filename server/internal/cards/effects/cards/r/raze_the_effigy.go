package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raze the Effigy — Instant {R} (Innistrad: Midnight Hunt, #5497):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Target attacking creature gets +2/+2 until end of turn."
//
// Abrade's modal shape; each bullet carries its own target clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9286ddc5-8151-4129-aaf6-2a2abe7047e5",
		Name:         "Raze the Effigy",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact())), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.Mode("Target attacking creature gets +2/+2 until end of turn.",
				effects.TargetCreature("target attacking creature", effects.AttackingCreature())), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2))),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				return effects.DestroyTarget{Target: target}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.BoostUntilEOT{Target: target, Power: 2, Toughness: 2, Label: "Raze the Effigy — +2/+2"}.Apply(ctx)
			}
			return nil
		},
	})
}
