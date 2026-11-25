package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dual Shot — Instant {R} (Ixalan, #5532):
//
//	"Dual Shot deals 1 damage to each of up to two target creatures."
//
// Each target still legal as the spell resolves takes 1 at once
// (CR 608.2b); with none chosen it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e56056dc-637e-4229-ae2c-dfd06d24086b",
		Name:         "Dual Shot",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 1)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SourceDealsDamageToEachLegalTarget(1)(ctx.Game, item)
		},
	})
}
