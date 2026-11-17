package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artillery Blast — Instant {1}{W}:
//
//	"Domain — Artillery Blast deals X damage to target tapped creature,
//	 where X is 1 plus the number of basic land types among lands you
//	 control."
//
// Deadly Riposte's tapped-creature target (tapped as the spell is cast and
// again as it resolves, CR 608.2b) with Frantic Firebolt's amount read as
// the spell resolves: 1 plus the shared domain count (DomainCount, #5666).
// The Purpose declares the floor of the damage, 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dad0b69f-ef33-475b-aebc-3a3ca880feb1",
		Name:         "Artillery Blast",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 1)),
		Targets:      TargetCreature("target tapped creature", b751Tapped()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageTheClauseTarget(0, 1+DomainCount(ctx.Game, ctx.Controller()))(ctx)
		},
	})
}
