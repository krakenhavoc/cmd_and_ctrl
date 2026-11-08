package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stonesplitter Bolt — Instant {X}{R} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Stonesplitter Bolt deals X damage to target creature or planeswalker.
//	 If this spell was bargained, it deals twice X damage to that
//	 permanent instead."
//
// Bargain is the shared optional cost (g2Bargain), read back from the
// payment record at resolution as Torch the Tower reads it. A bargained
// cast deals 2X in place of X, never both. The declared purpose is the
// unbargained X, the least the spell does, and X=0 does nothing at all,
// so XMatters is set.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "1322734c-3c3e-4885-a6d2-d6460a16fca2",
		Name:          "Stonesplitter Bolt",
		Completeness:  CompletenessFull,
		XMatters:      true,
		Purpose:       game.Purpose{Targets: game.ForTargets(DamageXToTarget(0)), Answers: game.AnswerRemove},
		Targets:       TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			amount := ctx.X()
			if g2WasBargained(ctx) {
				amount *= 2
			}
			return DealDamageToTheTarget(amount)(ctx.Game, item)
		},
	})
}
