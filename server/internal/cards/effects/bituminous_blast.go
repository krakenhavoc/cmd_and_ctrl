package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d291d8f7-bb67-470c-8794-cd553c91e5cc",
		Name:         "Bituminous Blast",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		Triggered:    []game.TriggeredAbility{Cascade()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}.Apply(ctx)
		},
	})
}
