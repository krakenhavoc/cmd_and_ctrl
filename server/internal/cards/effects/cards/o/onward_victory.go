package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Onward // Victory — Instant // Sorcery, Aftermath (Amonkhet, S82 OLD batch 243, #5535):
//
//	Onward {2}{R}: "Target creature gets +X/+0 until end of turn, where X is its
//	 power."
//	Victory {2}{W}: "Aftermath. Target creature gains double strike until end of
//	 turn."
//
// Two registered faces. Onward reads the creature's power through the layers as
// it resolves, once, and gives that much +X/+0 (a creature with negative power
// gets nothing). Victory is cast only from a graveyard and exiled as it
// resolves (CR 702.127a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a52c1ea-35ac-4028-8c1f-b7971c1df886",
		Name:         "Onward",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0)), Answers: game.AnswerPump},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := effects.B31CurrentPowerOf(ctx.Game, id)
			if x <= 0 {
				return nil
			}
			return effects.BoostUntilEOT{Target: id, Power: x, Label: "Onward — +X/+0 until end of turn"}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "9a52c1ea-35ac-4028-8c1f-b7971c1df886#1",
		Name:         "Victory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike")), Answers: game.AnswerCombatGrant},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Victory — double strike"}.Apply(ctx)
		},
	})
}
