package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shining Shoal — Instant — Arcane {X}{W}{W}:
//
//	"You may exile a white card with mana value X from your hand rather
//	 than pay this spell's mana cost.
//	 The next X damage that a source of your choice would deal to you
//	 and/or creatures you control this turn is dealt to any target
//	 instead."
//
// ADR 0108 §9 and owner decision 1 (#1905): Harm's Way's charged
// redirection with a charge of X (CR 615.7). The rulings: damage to
// several of your things at once is divided — you choose which X is
// redirected (divide_shield); a target gone or out of the game leaves the
// damage where it was.
//
// Simplification: the alternative cost is not offered, because nothing
// lets X be the mana value of the card exiled to pay it.
func init() {
	Register(Spec{
		OracleID:     "9937dae8-e639-46ad-849b-7e7be93bbbad",
		Name:         "Shining Shoal",
		XMatters:     true,
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You can't cast it by exiling a white card from your hand."},
		Targets:      TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.X() < 1 {
				return nil
			}
			return RedirectDamage{Choose: true, Protect: ShieldYouAndYourCreatures, Amount: ctx.X(), To: RedirectToClause(0)}.Apply(ctx)
		},
	})
}
