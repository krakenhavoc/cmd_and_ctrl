package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stolen Grain — Sorcery {4}{B}{B} (Portal Three Kingdoms, S82 #5622):
//
//	"Stolen Grain deals 5 damage to target opponent or planeswalker. You gain 5
//	 life."
//
// Burning Fields' damage, then the life: gained whether or not the damage was
// prevented (CR 608.2c, it is a separate instruction), but not at all when the
// only target is illegal on resolution and the spell is countered by the rules
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc17836d-8edf-44bd-a9ee-346d6af25dbf",
		Name:         "Stolen Grain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 5, Targets: effects.ForTargets(effects.DamageToTarget(0, 5)).Targets},
		Targets:      effects.TargetOpponentOrPlaneswalker(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			dealt, err := effects.OldB708DamageToTargetOpponentOrPlaneswalker(ctx, 5)
			if err != nil || !dealt {
				return err
			}
			return effects.GainLife{Amount: 5}.Apply(ctx)
		},
	})
}
