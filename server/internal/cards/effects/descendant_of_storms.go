package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Descendant of Storms — Creature — Human Soldier {W}, 2/1 (Tarkir:
// Dragonstorm):
//
//	"Whenever this creature attacks, you may pay {1}{W}. If you do, it
//	 endures 1. (Put a +1/+1 counter on it or create a 1/1 white Spirit
//	 creature token.)"
//
// The payment is Krovod Haunch's MayPay, asked as the trigger resolves
// (CR 603.5) and made once per trigger; the paid half is the keyword
// action endure (CR 701.63, EndureN): the counter-or-token pick is made
// on resolution, and a Descendant that has left the battlefield simply
// makes the Spirit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca773c52-ca64-463c-99cf-e71b59d2cff7",
		Name:         "Descendant of Storms",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Descendant of Storms — you may pay {1}{W}; if you do, it endures 1", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}{W}",
					Question: "Descendant of Storms — pay {1}{W} to have it endure 1?",
					Then:     descendantOfStormsEndure,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// descendantOfStormsEndure is the paid half, keyed. An on-disk key: never
// renamed, never reused.
var descendantOfStormsEndure = PayUnlessBranches("pay-unless/descendant-of-storms-endure",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return EndureN(ctx, ctx.Source(), 1, "Descendant of Storms")
	}, nil)
