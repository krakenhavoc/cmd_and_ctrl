package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bituminous Blast — Instant {3}{B}{R}:
//
//	"Cascade
//	 Bituminous Blast deals 4 damage to target creature."
//
// Cascade is the catalog keyword (it triggers on cast and resolves before
// the Blast, CR 702.85a); the damage is a plain targeted spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d291d8f7-bb67-470c-8794-cd553c91e5cc",
		Name:         "Bituminous Blast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		Triggered:    []game.TriggeredAbility{effects.Cascade()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}.Apply(ctx)
		},
	})
}
