package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9a294cbd-df18-4ee0-acaa-e31390a7f508",
		Name:         "Tweeze",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return damageToFirstTarget(3)(ctx.Item, ctx) }),
			Instruction(b39MayDiscardThenDraw(1, false, rummage, drawOneIfDiscarded)),
		),
	})
}
