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
// The alternative cost (#3175) is a pitch whose card is bound to X: an
// alternative cost that includes X, so CR 107.3b does not lock X at 0.
// The caster announces X (CR 107.3a, 601.2b) and exiles a white card
// whose mana value is exactly that X (WithManaValueEqualsX on the pitch;
// a card with {X} in its own cost counts it as 0 in hand, CR 202.3e).
// The client sends the picked card's mana value as the X; the bot is
// offered one cast per distinct mana value among its white cards. On the
// stack X is the announced value, so the charge is the pitched card's
// mana value.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9937dae8-e639-46ad-849b-7e7be93bbbad",
		Name:         "Shining Shoal",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		AlternativeCosts: []game.AlternativeCost{
			Pitch(
				"Exile a white card with mana value X from your hand",
				0,
				CardInYourHand("a white card with mana value X from your hand", OfColor("W")).WithManaValueEqualsX(),
				"a white card with mana value X",
			),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.X() < 1 {
				return nil
			}
			return RedirectDamage{Choose: true, Protect: ShieldYouAndYourCreatures, Amount: ctx.X(), To: RedirectToClause(0)}.Apply(ctx)
		},
	})
}
