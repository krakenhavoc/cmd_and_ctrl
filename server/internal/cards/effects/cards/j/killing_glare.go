package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Killing Glare — Instant {X}{B}:
//
//	"Destroy target creature with power X or less."
//
// ADR 0109 §9 (#1842): the power bound reads the announced X, which is
// chosen before the target (CR 601.2b–c) and re-checked as the spell
// resolves (CR 608.2b) — a creature pumped past X in response is an
// illegal target and the spell does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b47891f-130a-4d76-b262-53f5e3bb6b95",
		Name:         "Killing Glare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with power X or less").WithPowerAtMostX(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyTheFirstLegalCardTarget(ctx)
		},
	})
}
