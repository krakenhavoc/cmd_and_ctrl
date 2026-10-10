package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ranger's Firebrand — Sorcery {R}:
//
//	"Ranger's Firebrand deals 2 damage to any target. The Ring tempts
//	 you."
//
// The tempt runs once the damage is dealt (a damage replacement can
// stop to ask, CR 616.1). State-based actions are not checked until
// the spell has resolved (CR 704.3), so a creature of yours dealt
// lethal damage may still be chosen as the Ring-bearer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "796d3836-d820-440c-b0ad-5907b3cfcb10",
		Name:         "Ranger's Firebrand",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			tempt := ringTemptsYouNext(item)
			return ctx.Game.DealDamageEachThenForEffect(ctx.Source(), legalTargetIDs(ctx), 2,
				func(g *game.Game, _ int) error { return tempt(g) })
		},
	})
}
