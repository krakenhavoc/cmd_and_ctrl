package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foresee — Sorcery {3}{U} (Future Sight, #5594):
//
//	"Scry 4, then draw two cards."
//
// Preordain's shape at Foresee's numbers: the scry prompt, then the draws
// as the scry's registered tail, so the prompt is a restore point and the
// cards are drawn only after the top of the library is set.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "441de890-7b9f-42f7-859a-a154b8ed74ac",
		Name:         "Foresee",
		Purpose:      game.Purpose{Draws: 2, Scry: 4},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return Scry{
				Player:   controller,
				N:        4,
				ThenTail: game.KeyedTail{Ref: thenDrawCards, Player: controller, N: 2},
			}.Apply(ctx)
		},
	})
}
