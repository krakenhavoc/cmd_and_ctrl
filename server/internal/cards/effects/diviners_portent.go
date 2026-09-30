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
			controller := ctx.Controller()
			x := ctx.X()
			results, err := rollDice(ctx, 20, 1)
			if err != nil || len(results) == 0 {
				return err
			}
			total := results[0]
			if p := ctx.Game.PlayerByIDForEffect(controller); p != nil && p.Hand != nil {
				total += p.Hand.Size()
			}
			if total < 15 {
				return DrawCards{Player: controller, N: x}.Apply(ctx)
			}
			return Scry{Player: controller, N: x, Then: func(g *game.Game) error {
				return g.DrawNForEffect(controller, x)
			}}.Apply(ctx)
		},
	})
}
