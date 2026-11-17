package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_one_b06_helpers.go — shared bodies for the ONE batch 6 proliferate
// creatures (S82, #5479).

// s82OneB06ProliferatePaid is "you may pay {cost}. If you do, proliferate"
// as MayPay's paid branch (Venomous Brutalizer, Tainted Observer). An
// on-disk key: never renamed, never reused.
var s82OneB06ProliferatePaid = PayUnlessBranches("pay-unless/s82-one-b06-proliferate",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return Proliferate{}.Apply(ctx)
	}, nil)

// s82OneB06ProliferateIfDone is "If you do, proliferate" as the tail of a
// may-instruction: it proliferates only when the instruction's one object
// went (Cacophony Scamp's sacrifice). An on-disk key.
var s82OneB06ProliferateIfDone = CardTail("then/s82-one-b06-proliferate-if-done", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return Proliferate{}.Apply(ctx)
})
