package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tweeze — Instant {2}{R} (Lorwyn Eclipsed, #4458):
//
//	"Tweeze deals 3 damage to any target. You may discard a card. If you
//	 do, draw a card."
//
// Lightning Bolt's damage, then Incinerating Blast's rummage
// (b39MayDiscardThenDraw): each is its own step, so the optional discard
// waits for any prompt the damage asks (CR 608.2c), and a card is drawn
// only for a card discarded. With the target illegal as the spell resolves
// nothing happens, and there is no discard (the ruling, CR 608.2b).
//
// No simplification.
func init() {
	const rummage = "Tweeze — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "9a294cbd-df18-4ee0-acaa-e31390a7f508",
		Name:         "Tweeze",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DamageToFirstTargetStep(3)(ctx.Item, ctx) }),
			effects.Instruction(effects.B39MayDiscardThenDraw(1, false, rummage, effects.DrawOneIfDiscarded)),
		),
	})
}
