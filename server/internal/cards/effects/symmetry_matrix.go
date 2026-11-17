package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Symmetry Matrix — Artifact {4} (The Brothers' War, #5482):
//
//	"Whenever a creature you control with power equal to its toughness
//	 enters, you may pay {1}. If you do, draw a card."
//
// Power and toughness are read as the creature enters, through the layers
// (counters and enter-with effects included). The Matrix is not a
// creature, so it never triggers on itself. "You may pay {1}" is a MayPay
// asked as the trigger resolves; a pool that cannot pay, or a decline,
// draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f6d129c-e6ad-4f35-b8b1-56a101fdd5db",
		Name:         "Symmetry Matrix",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.CurrentPower() == c.CurrentToughness()
			}, "Symmetry Matrix — you may pay {1}. If you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Pay {1} to draw a card?",
						Then:     s82BroB07MatrixPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var s82BroB07MatrixPaid = PayUnlessBranches("pay-unless/symmetry-matrix-draw",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}, nil)
