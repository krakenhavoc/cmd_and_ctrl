package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Estwald Shieldbasher — Creature — Human Soldier {3}{W}, 4/2 (Innistrad:
// Crimson Vow, #5495):
//
//	"Whenever this creature attacks, you may pay {1}. If you do, it gains
//	 indestructible until end of turn."
//
// The "you may pay {1}" is a MayPay asked as the trigger resolves
// (Brilliant Wings'); declining, or a pool that cannot pay, grants
// nothing. "It" is this creature: a Shieldbasher that has left and come
// back is a new object (CR 400.7) and gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9f111aa-5d34-4317-9a03-35cd94a888c7",
		Name:         "Estwald Shieldbasher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Estwald Shieldbasher — you may pay {1}; if you do, it gains indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Estwald Shieldbasher — pay {1} to give it indestructible until end of turn?",
						Then:     estwaldShieldbasherPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// estwaldShieldbasherPaid grants the keyword. An on-disk key: never
// renamed, never reused.
var estwaldShieldbasherPaid = PayUnlessBranches("pay-unless/estwald-shieldbasher-indestructible",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if ctx.Item == nil || sourceIsNewObject(ctx.Game, ctx.Item) {
			return nil
		}
		return GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"indestructible"},
			Label: "Estwald Shieldbasher — indestructible"}.Apply(ctx)
	}, nil)
