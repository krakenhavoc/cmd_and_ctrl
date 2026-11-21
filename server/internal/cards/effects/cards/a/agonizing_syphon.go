package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "11669fe5-9298-4e1a-a4f7-6a4598dbc8dd",
		Name:         "Agonizing Syphon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageXToTheTargetAndGainX(item, ctx, 3)
		},
	})
}
