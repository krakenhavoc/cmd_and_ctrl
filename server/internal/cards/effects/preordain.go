package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Preordain — Sorcery {U}:
//
//	"Scry 2, then draw a card."
//
// One mana to see three cards deep and keep the best one. "Then" is
// load-bearing: the scry finishes before the draw, so the card you
// leave on top is the card you draw — which is the whole point, and
// why this cannot be modelled as "draw a card, then scry 2".
//
// That ordering is why the draw goes in Scry.Then rather than on the
// next line: the scry only QUEUES a prompt, so a draw written after it
// would resolve first — drawing one of the two cards the player is
// still deciding about, and leaving the prompt unanswerable because
// that card is no longer in the library to put back.
func init() {
	Register(Spec{
		OracleID:     "ac641490-ca14-48d7-8cc4-b69ce984befa",
		Name:         "Preordain",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return Scry{
				Player: controller,
				N:      2,
				Then: func(g *game.Game) error {
					return g.DrawNForEffect(controller, 1)
				},
			}.Apply(ctx)
		},
	})
}
