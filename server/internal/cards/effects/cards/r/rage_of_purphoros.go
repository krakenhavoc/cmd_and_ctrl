package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rage of Purphoros — Sorcery {4}{R}:
//
//	"Rage of Purphoros deals 4 damage to target creature. It can't be
//	 regenerated this turn. Scry 1."
//
// The no-regeneration rider is the spell's, not the damage's (ADR 0108
// §2). The scry happens after both, as printed, and only if the target is
// still legal (an illegal target means the spell does not resolve at all).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6667a93-c706-451f-a83d-ffe5b9b9e53e",
		Name:         "Rage of Purphoros",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.WithRiders(effects.ForTargets(effects.DamageToTarget(0, 4)), game.Purpose{Scry: 1}),
		OnResolve: effects.DamageFirstTargetThenNoRegen(4, func(ctx *effects.Context) error {
			return effects.Scry{N: 1}.Apply(ctx)
		}),
	})
}
