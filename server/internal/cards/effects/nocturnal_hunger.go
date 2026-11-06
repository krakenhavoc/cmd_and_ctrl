package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nocturnal Hunger — Instant {2}{B} (Bloomburrow, #4467):
//
//	"Gift a Food (You may promise an opponent a gift as you cast this
//	 spell. If you do, they create a Food token before its other effects.
//	 It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3
//	 life.")
//	 Destroy target creature. If the gift wasn't promised, you lose 2
//	 life."
//
// The gift keyword (gift.go, CR 702.174): a promised Food is made for the
// chosen opponent before the destruction (CR 702.174j). The life loss is
// the price of keeping the Food: it reads the promise (CR 702.174k), not
// whether the creature was destroyed, so an indestructible target still
// costs 2 life when nothing was promised. The two sentences are two steps
// (CR 608.2c), so a question the destruction asks is answered before the
// life is lost. A target that left in response counters the spell
// (CR 608.2b): no Food, and no life lost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed7f64d1-69ba-40f5-9863-3edba10e23c7",
		Name:         "Nocturnal Hunger",
		Completeness: CompletenessFull,
		Gift:         GiftAFood(),
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return destroyTheTargetPermanent(ctx.Item, ctx) }),
			Instruction(nocturnalHungerLoss),
		),
	})
}

// nocturnalHungerLoss is "If the gift wasn't promised, you lose 2 life."
func nocturnalHungerLoss(ctx *Context) error {
	if ctx.GiftPromised() {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
}
