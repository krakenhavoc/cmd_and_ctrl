package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agonizing Syphon — Sorcery {3}{B} (Foundations Jumpstart, #5929):
//
//	"Agonizing Syphon deals 3 damage to any target and you gain 3 life."
//
// Consume Spirit's two steps with a fixed 3: the damage, then the gain. The
// life is a flat 3, not the damage dealt, so a prevented hit still gains it;
// a target gone as the spell resolves stops the whole spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11669fe5-9298-4e1a-a4f7-6a4598dbc8dd",
		Name:         "Agonizing Syphon",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return damageXToTheTargetAndGainX(item, ctx, 3)
		},
	})
}
