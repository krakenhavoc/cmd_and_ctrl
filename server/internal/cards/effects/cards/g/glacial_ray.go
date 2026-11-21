package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glacial Ray — Instant — Arcane {1}{R}:
//
//	"Glacial Ray deals 2 damage to any target.
//	Splice onto Arcane {1}{R}"
//
// Spliced onto an Arcane spell, "Glacial Ray" is that spell (CR 702.47c):
// the damage is dealt by it, with its colour, so a creature with
// protection from red may be targeted from a blue spell (ADR 0202).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e65a8ab8-84a7-4fe2-b009-ceb87bec6b4e",
		Name:         "Glacial Ray",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{1}{R}"),
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
		},
	})
}
