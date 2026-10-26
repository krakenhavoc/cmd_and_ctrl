package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diviner's Portent — Instant {X}{U}{U}{U}:
//
//	"Roll a d20 and add the number of cards in your hand.
//	 1—14 | Draw X cards.
//	 15+ | Scry X, then draw X cards."
//
// One d20 from the keyed random stream (ADR 0054), plus the size of
// the caster's hand as the spell resolves — the Portent itself is on
// the stack by then and is not counted. The total picks one of the two
// rows; they are exclusive, so this is an `if`, not Spec.Modes.
//
// The draw after the scry goes in Scry.Then, because the scry only
// queues a prompt: a draw written on the next line would take cards
// the player is still deciding about.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "119585c7-ddfa-47ed-b2f8-488ebc156222",
		Name:         "Diviner's Portent",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return RollDice{Sides: 20, N: 1, Then: divinersPortentResult, TailN: ctx.X()}.Apply(ctx)
		},
	})
}

// divinersPortentResult is the results table, after the d20 (#4307: a
// replaced roll waits for its chooser). X was read as the spell resolved
// (the tail's N); the hand is counted as the roll settles.
var divinersPortentResult = CardTail("dice/diviners-portent-result", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 {
		return nil
	}
	controller := ctx.Controller()
	x := o.Tail.N
	total := o.Rolls[0]
	if p := ctx.Game.PlayerByIDForEffect(controller); p != nil && p.Hand != nil {
		total += p.Hand.Size()
	}
	if total < 15 {
		return DrawCards{Player: controller, N: x}.Apply(ctx)
	}
	return Scry{Player: controller, N: x, ThenTail: ctx.TailN(thenDrawCards, x)}.Apply(ctx)
})
